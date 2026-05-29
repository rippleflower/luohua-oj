import { contestMakeupListSchema, type ContestDetail, type ContestMakeupList } from "@oj/shared";

import { env } from "../../lib/env";
import { ApiError, getJSON } from "../../lib/http-client";
import { getContest } from "../contests/api";

export async function getContestMakeupList(slug: string): Promise<ContestMakeupList> {
  if (env.apiBaseUrl !== "") {
    const response = await getJSON<unknown>(`/contests/${encodeURIComponent(slug)}/makeup-list`);
    return contestMakeupListSchema.parse(response);
  }

  if (!env.demoMode) {
    throw new ApiError(
      "contest makeup api is unavailable because VITE_API_BASE_URL is empty and VITE_DEMO_MODE is false",
      503,
    );
  }

  const contest = await getContest(slug);
  return buildFallbackMakeup(slug, contest);
}

function buildFallbackMakeup(slug: string, contest?: ContestDetail): ContestMakeupList {
  if (!contest) {
    throw new ApiError("contest not found", 404);
  }
  if (contest.status !== "ENDED") {
    throw new ApiError("contest is not ended", 409);
  }

  const attempted = contest.problems
    .filter((problem) => problem.status === "ATTEMPTED")
    .sort(compareProblems)
    .map((problem) => ({
      problemId: `${slug}:${problem.code}`,
      problemCode: problem.code,
      problemSlug: problem.slug,
      problemTitle: problem.title,
      difficulty: problem.difficulty,
      category: "ATTEMPTED_UNSOLVED" as const,
      lastStatus: "WRONG_ANSWER",
      attemptCount: 1,
      severityRank: statusSeverityRank("WRONG_ANSWER"),
      reasonSummary: "比赛中尝试过但尚未通过",
      suggestedAction: "优先复查边界条件并重提",
    }));

  const unattempted = contest.problems
    .filter((problem) => problem.status === "LOCKED")
    .sort(compareProblems)
    .map((problem) => ({
      problemId: `${slug}:${problem.code}`,
      problemCode: problem.code,
      problemSlug: problem.slug,
      problemTitle: problem.title,
      difficulty: problem.difficulty,
      category: "UNATTEMPTED_RECOMMENDED" as const,
      severityRank: statusSeverityRank(""),
      reasonSummary: "比赛中未尝试该题",
      suggestedAction: "先完成一次可运行提交，再按结果迭代",
    }));

  return contestMakeupListSchema.parse({
    contestSlug: contest.slug,
    generatedAt: new Date().toISOString(),
    items: [...attempted, ...unattempted],
  });
}

function compareProblems(
  left: Pick<ContestDetail["problems"][number], "difficulty" | "code">,
  right: Pick<ContestDetail["problems"][number], "difficulty" | "code">,
) {
  const leftRank = difficultyRank(left.difficulty);
  const rightRank = difficultyRank(right.difficulty);
  if (leftRank !== rightRank) {
    return leftRank - rightRank;
  }
  return left.code.localeCompare(right.code);
}

function difficultyRank(difficulty: "EASY" | "MEDIUM" | "HARD"): number {
  switch (difficulty) {
    case "EASY":
      return 0;
    case "MEDIUM":
      return 1;
    case "HARD":
    default:
      return 2;
  }
}

function statusSeverityRank(status: string): number {
  switch (status) {
    case "RUNTIME_ERROR":
      return 0;
    case "MEMORY_LIMIT_EXCEEDED":
      return 1;
    case "TIME_LIMIT_EXCEEDED":
      return 2;
    case "WRONG_ANSWER":
      return 3;
    case "COMPILE_ERROR":
      return 4;
    case "SYSTEM_ERROR":
      return 5;
    case "RUNNING":
      return 6;
    case "PENDING":
      return 7;
    default:
      return 99;
  }
}
