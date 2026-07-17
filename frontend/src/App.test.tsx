import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import { App } from "./App";
import type { Service } from "./api/services";

const service: Service = {
  id: "6b276f3d-0fd5-4572-9dfa-875f4dd824ef",
  serviceKey: "orders-api",
  displayName: "Orders API",
  description: "Core order management",
  ownerTeam: "orders-team",
  repositoryUrl: "https://github.com/example/orders-api",
  lifecycle: "production",
  createdAt: "2026-07-17T01:00:00Z",
  updatedAt: "2026-07-17T02:00:00Z",
  archivedAt: null,
};

describe("service catalog application", () => {
  it("renders the active catalog empty state", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn<typeof fetch>().mockResolvedValue(pageResponse([])),
    );

    renderApp("/services");

    expect(
      await screen.findByRole("heading", {
        name: "Register the first service",
      }),
    ).toBeInTheDocument();
  });

  it("recovers from a catalog loading error", async () => {
    const fetchMock = vi
      .fn<typeof fetch>()
      .mockRejectedValueOnce(new Error("network unavailable"))
      .mockResolvedValueOnce(pageResponse([]));
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    renderApp("/services");

    expect(
      await screen.findByRole("heading", {
        name: "We could not load services",
      }),
    ).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Try again" }));

    expect(
      await screen.findByRole("heading", {
        name: "Register the first service",
      }),
    ).toBeInTheDocument();
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });

  it("shows field feedback for a duplicate service key", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn<typeof fetch>().mockResolvedValue(
        jsonResponse(
          {
            error: {
              code: "service_key_conflict",
              message: "A service with this service key already exists.",
              requestId: "request-duplicate",
              details: [{ field: "serviceKey", message: "must be unique" }],
            },
          },
          409,
        ),
      ),
    );
    const user = userEvent.setup();
    renderApp("/services/new");

    await user.type(screen.getByLabelText(/Service key/), "orders-api");
    await user.type(screen.getByLabelText(/Display name/), "Orders API");
    await user.type(screen.getByLabelText(/Owner team/), "orders-team");
    await user.type(
      screen.getByLabelText(/GitHub repository/),
      "https://github.com/example/orders-api",
    );
    await user.click(screen.getByRole("button", { name: "Register service" }));

    expect(await screen.findByText("must be unique")).toBeInTheDocument();
  });

  it("edits service metadata and returns to details", async () => {
    const updated = {
      ...service,
      displayName: "Orders Platform",
      updatedAt: "2026-07-17T03:00:00Z",
    };
    const fetchMock = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(jsonResponse(service))
      .mockResolvedValueOnce(jsonResponse(updated))
      .mockResolvedValueOnce(jsonResponse(updated));
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    renderApp(`/services/${service.id}/edit`);

    const displayName = await screen.findByLabelText("Display name");
    await user.clear(displayName);
    await user.type(displayName, "Orders Platform");
    await user.click(screen.getByRole("button", { name: "Save changes" }));

    expect(
      await screen.findByRole("heading", { level: 1, name: "Orders Platform" }),
    ).toBeInTheDocument();
    expect(fetchMock.mock.calls[1]?.[1]).toEqual(
      expect.objectContaining({ method: "PUT" }),
    );
  });

  it("confirms archival and opens the archived catalog", async () => {
    const fetchMock = vi
      .fn<typeof fetch>()
      .mockResolvedValueOnce(jsonResponse(service))
      .mockResolvedValueOnce(new Response(null, { status: 204 }))
      .mockResolvedValueOnce(
        pageResponse([{ ...service, archivedAt: "2026-07-17T04:00:00Z" }]),
      );
    vi.stubGlobal("fetch", fetchMock);
    const user = userEvent.setup();
    renderApp(`/services/${service.id}`);

    await user.click(
      await screen.findByRole("button", { name: "Archive service" }),
    );
    const dialog = screen.getByRole("dialog", {
      name: `Archive ${service.displayName}?`,
    });
    await user.click(
      within(dialog).getByRole("button", { name: "Archive service" }),
    );

    expect(
      await screen.findByText("Service archived successfully."),
    ).toBeInTheDocument();
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(3));
    expect(fetchMock.mock.calls[1]?.[1]).toEqual(
      expect.objectContaining({ method: "DELETE" }),
    );
  });
});

function renderApp(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <App />
    </MemoryRouter>,
  );
}

function pageResponse(items: Service[]): Response {
  return jsonResponse({
    items,
    pagination: { total: items.length, limit: 20, offset: 0 },
  });
}

function jsonResponse(value: unknown, status = 200): Response {
  return new Response(JSON.stringify(value), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
