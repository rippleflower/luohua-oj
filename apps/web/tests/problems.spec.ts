import { expect, test } from "@playwright/test";

test("problem list renders", async ({ page }) => {
  await page.goto("/problems");
  await expect(page.getByRole("heading", { name: "Problems" })).toBeVisible();
  await expect(page.getByRole("link", { name: "Two Sum" })).toBeVisible();
});
