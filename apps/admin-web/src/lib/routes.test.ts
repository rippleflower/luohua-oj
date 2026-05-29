import { describe, expect, it } from "vitest";

import { matchAdminRoute } from "./routes";

describe("matchAdminRoute", () => {
  it("matches contests page", () => {
    expect(matchAdminRoute("/contests")).toEqual({ kind: "contests" });
  });

  it("matches user permissions page", () => {
    expect(matchAdminRoute("/users/user-1/permissions")).toEqual({
      kind: "user-permissions",
      userId: "user-1",
    });
  });

  it("falls back to dashboard for unknown paths", () => {
    expect(matchAdminRoute("/unknown/route")).toEqual({ kind: "dashboard" });
  });
});
