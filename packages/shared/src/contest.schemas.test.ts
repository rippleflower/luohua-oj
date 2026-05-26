import { describe, expect, it } from "vitest";

import { contestDetailSchema } from "./contest.schemas";

function baseDetail() {
  return {
    id: "contest-1",
    slug: "spring-open",
    title: "Spring Open",
    status: "RUNNING",
    startsAt: "2026-05-16T11:00:00Z",
    endsAt: "2026-05-16T13:00:00Z",
    duration: "2 hours",
    problemCount: 2,
    participantCount: 128,
    blurb: "fast contest",
    rankSummary: "Top 10%",
    remaining: "00:30:00",
    recentSubmissions: [
      { id: "sub-1", problemCode: "A", status: "ACCEPTED", at: "1 minute ago" },
    ],
  } as const;
}

describe("contestDetailSchema", () => {
  it("accepts problems with slug", () => {
    const parsed = contestDetailSchema.parse({
      ...baseDetail(),
      problems: [
        {
          code: "A",
          slug: "two-sum",
          title: "Two Sum",
          difficulty: "EASY",
          status: "SOLVED",
        },
      ],
    });

    expect(parsed.problems[0].slug).toBe("two-sum");
  });

  it("accepts legacy problems without slug", () => {
    const parsed = contestDetailSchema.parse({
      ...baseDetail(),
      problems: [
        {
          code: "A",
          title: "Two Sum",
          difficulty: "EASY",
          status: "SOLVED",
        },
      ],
    });

    expect(parsed.problems[0].slug).toBeUndefined();
  });
});
