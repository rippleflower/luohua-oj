import { beforeEach, describe, expect, it, vi } from "vitest";

const mockedGetJSON = vi.hoisted(() => vi.fn());

vi.mock("../../lib/env", () => ({
  env: {
    apiBaseUrl: "http://localhost:8080",
    adminBaseUrl: "",
    submissionsUsername: "",
  },
}));

vi.mock("../../lib/http-client", async () => {
  const actual = await vi.importActual<typeof import("../../lib/http-client")>("../../lib/http-client");
  return {
    ...actual,
    getJSON: mockedGetJSON,
  };
});

import { getContestMakeupList } from "./api";

describe("contest makeup api", () => {
  beforeEach(() => {
    mockedGetJSON.mockReset();
  });

  it("loads remote makeup list", async () => {
    mockedGetJSON.mockResolvedValueOnce({
      contestSlug: "april-grand-prix",
      generatedAt: "2026-05-26T09:00:00Z",
      items: [
        {
          problemId: "p-1",
          problemCode: "D",
          problemTitle: "D",
          difficulty: "MEDIUM",
          category: "ATTEMPTED_UNSOLVED",
          attemptCount: 2,
          severityRank: 3,
          reasonSummary: "WA",
          suggestedAction: "retry",
        },
      ],
    });

    const result = await getContestMakeupList("april-grand-prix");
    expect(mockedGetJSON).toHaveBeenCalledWith("/contests/april-grand-prix/makeup-list");
    expect(result.items).toHaveLength(1);
    expect(result.items[0].problemCode).toBe("D");
  });
});
