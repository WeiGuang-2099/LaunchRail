import { describe, expect, it, vi } from "vitest";
import { listServices } from "./services";

describe("service API client", () => {
  it("serializes list query parameters", async () => {
    const fetchMock = vi.fn<typeof fetch>().mockResolvedValue(
      jsonResponse({
        items: [],
        pagination: { total: 0, limit: 20, offset: 40 },
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await listServices("archived", 20, 40);

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/v1/services?status=archived&limit=20&offset=40",
      expect.objectContaining({ signal: undefined }),
    );
  });

  it("throws a typed API error", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn<typeof fetch>().mockResolvedValue(
        jsonResponse(
          {
            error: {
              code: "service_not_found",
              message: "The requested service was not found.",
              requestId: "request-123",
              details: [],
            },
          },
          404,
        ),
      ),
    );

    await expect(listServices("active", 20, 0)).rejects.toMatchObject({
      status: 404,
      code: "service_not_found",
      requestId: "request-123",
    });
  });
});

function jsonResponse(value: unknown, status = 200): Response {
  return new Response(JSON.stringify(value), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}
