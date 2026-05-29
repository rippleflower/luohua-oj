import { describe, expect, it } from "vitest";

import { matchWebRoute } from "./routes";

describe("matchWebRoute", () => {
  it("matches canonical problem detail routes", () => {
    expect(matchWebRoute("/problems/p/ABC123")).toEqual({
      kind: "problem-detail",
      routeCode: "ABC123",
    });
  });

  it("matches legacy slug routes", () => {
    expect(matchWebRoute("/problems/two-sum")).toEqual({
      kind: "problem-detail",
      slug: "two-sum",
    });
  });

  it("matches contest makeup route", () => {
    expect(matchWebRoute("/contests/spring-open/makeup-list")).toEqual({
      kind: "contest-makeup",
      slug: "spring-open",
    });
  });
});
