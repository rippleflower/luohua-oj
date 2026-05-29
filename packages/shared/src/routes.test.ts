import { describe, expect, it } from "vitest";

import { adminRoutes, isRouteActive, normalizePathname, webRoutes, webSectionForPath } from "./routes";

describe("routes", () => {
  it("builds canonical problem detail paths", () => {
    expect(webRoutes.problemDetail("ABC123")).toBe("/problems/p/ABC123");
    expect(webRoutes.legacyProblemDetail("two-sum")).toBe("/problems/two-sum");
    expect(webRoutes.contestMakeup("spring-open")).toBe("/contests/spring-open/makeup-list");
    expect(adminRoutes.userPermissions("user-1")).toBe("/users/user-1/permissions");
  });

  it("normalizes trailing slashes", () => {
    expect(normalizePathname("/problems/")).toBe("/problems");
    expect(normalizePathname("/")).toBe("/");
  });

  it("matches active routes", () => {
    expect(isRouteActive("/problems/p/ABC123", webRoutes.problems)).toBe(true);
    expect(isRouteActive("/contests", webRoutes.problems)).toBe(false);
  });

  it("maps web sections from paths", () => {
    expect(webSectionForPath("/")).toBe("home");
    expect(webSectionForPath("/problems/p/ABC123")).toBe("problems");
    expect(webSectionForPath("/settings/security")).toBe("me");
  });
});
