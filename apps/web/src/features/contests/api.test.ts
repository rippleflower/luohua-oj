import { beforeEach, describe, expect, it, vi } from "vitest";

const mockedGetJSON = vi.hoisted(() => vi.fn());

vi.mock("../../lib/env", () => ({
  env: {
    apiBaseUrl: "http://localhost:8080",
    adminBaseUrl: "",
    submissionsUsername: "",
  },
}));

vi.mock("../../lib/http-client", () => ({
  getJSON: mockedGetJSON,
}));

import { getContest, listContests } from "./api";

describe("contests api", () => {
  beforeEach(() => {
    mockedGetJSON.mockReset();
  });

  it("uses remote contests list when api request succeeds", async () => {
    mockedGetJSON.mockResolvedValueOnce([
      {
        id: "contest-remote",
        slug: "remote-open",
        title: "Remote Open",
        status: "RUNNING",
        startsAt: "2026-05-16T11:00:00Z",
        endsAt: "2026-05-16T13:00:00Z",
        duration: "2 hours",
        problemCount: 6,
        participantCount: 1000,
        blurb: "remote contest",
      },
    ]);

    const contests = await listContests();

    expect(mockedGetJSON).toHaveBeenCalledWith("/contests");
    expect(contests).toHaveLength(1);
    expect(contests[0]).toMatchObject({
      id: "contest-remote",
      slug: "remote-open",
      title: "Remote Open",
    });
  });

  it("falls back to local seed when remote detail request fails", async () => {
    mockedGetJSON.mockRejectedValueOnce(new Error("network down"));

    const contest = await getContest("spring-open");

    expect(mockedGetJSON).toHaveBeenCalledWith("/contests/spring-open");
    expect(contest).toBeDefined();
    expect(contest?.title).toBe("2026 春季公开赛");
  });

  it("accepts remote legacy problem snapshots without slug", async () => {
    mockedGetJSON.mockResolvedValueOnce({
      id: "contest-legacy",
      slug: "legacy-open",
      title: "Legacy Open",
      status: "ENDED",
      startsAt: "2026-05-16T11:00:00Z",
      endsAt: "2026-05-16T13:00:00Z",
      duration: "2 hours",
      problemCount: 1,
      participantCount: 16,
      blurb: "legacy contest",
      rankSummary: "Completed",
      remaining: "ended",
      recentSubmissions: [],
      problems: [
        {
          code: "A",
          title: "Legacy Problem",
          difficulty: "EASY",
          status: "LOCKED",
        },
      ],
    });

    const contest = await getContest("legacy-open");

    expect(contest).toBeDefined();
    expect(contest?.problems[0].slug).toBeUndefined();
    expect(contest?.problems[0].title).toBe("Legacy Problem");
  });
});
