import { expect, test } from "@playwright/test";

const environment =
  (
    globalThis as typeof globalThis & {
      process?: { env?: Record<string, string | undefined> };
    }
  ).process?.env ?? {};
const apiBaseURL = environment.E2E_API_URL ?? "http://127.0.0.1:8080";

test("the seeded service is available through the full stack", async ({
  page,
  request,
}) => {
  const readiness = await request.get(`${apiBaseURL}/readyz`);
  expect(readiness.ok()).toBeTruthy();
  await expect(readiness.json()).resolves.toMatchObject({ status: "ready" });

  await page.goto("/services");
  await expect(
    page.getByRole("heading", { name: "Registered services" }),
  ).toBeVisible();

  const serviceLink = page.getByRole("link", {
    name: "View LaunchRail E2E Service",
  });
  await expect(serviceLink).toContainText("launchrail-e2e");
  await expect(serviceLink).toContainText("platform-team");
  await serviceLink.click();

  await expect(
    page.getByRole("heading", { name: "LaunchRail E2E Service" }),
  ).toBeVisible();
  await expect(
    page.getByText("Deterministic service record for full-stack smoke tests."),
  ).toBeVisible();
  await expect(page.getByText("platform-team", { exact: true })).toBeVisible();
  await expect(
    page.getByRole("link", {
      name: "https://github.com/WeiGuang-2099/LaunchRail",
    }),
  ).toHaveAttribute("href", "https://github.com/WeiGuang-2099/LaunchRail");
});
