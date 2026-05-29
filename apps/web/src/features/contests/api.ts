import { z } from "zod";

import { env } from "../../lib/env";
import { ApiError, getJSON } from "../../lib/http-client";
import {
  contestDetailSchema,
  contestSummarySchema,
  type ContestDetail,
  type ContestSummary,
} from "./schema";

const contestSeed: ContestDetail[] = [
  {
    id: "spring-open",
    slug: "spring-open",
    title: "2026 春季公开赛",
    status: "RUNNING",
    startsAt: "2026-05-16 19:00",
    endsAt: "2026-05-16 21:00",
    duration: "2 小时",
    problemCount: 6,
    participantCount: 1248,
    blurb: "两小时混合场，偏重实现速度、图论直觉和稳定提交节奏。",
    rankSummary: "当前第 42 / 1,248 名",
    remaining: "01:17:24",
    recentSubmissions: [
      { id: "sub-9001", problemCode: "B", status: "ACCEPTED", at: "1 分钟前" },
      {
        id: "sub-9000",
        problemCode: "D",
        status: "WRONG_ANSWER",
        at: "7 分钟前",
      },
    ],
    problems: [
      {
        code: "A",
        slug: "two-sum",
        title: "位窗口",
        difficulty: "EASY",
        status: "SOLVED",
        firstSolve: "00:12",
      },
      {
        code: "B",
        slug: "shortest-path",
        title: "地铁换乘",
        difficulty: "MEDIUM",
        status: "SOLVED",
        firstSolve: "00:34",
      },
      {
        code: "C",
        slug: "dynamic-ranking",
        title: "零知识路径",
        difficulty: "MEDIUM",
        status: "ATTEMPTED",
      },
      {
        code: "D",
        title: "数据包风暴",
        difficulty: "HARD",
        status: "ATTEMPTED",
      },
      { code: "E", title: "稀疏幻象", difficulty: "HARD", status: "LOCKED" },
      { code: "F", title: "最后检查点", difficulty: "HARD", status: "LOCKED" },
    ],
  },
  {
    id: "nightly-qualifier",
    slug: "nightly-qualifier",
    title: "夜场资格赛",
    status: "UPCOMING",
    startsAt: "2026-05-17 20:30",
    endsAt: "2026-05-17 23:30",
    duration: "3 小时",
    problemCount: 7,
    participantCount: 832,
    blurb: "资格赛题面更直接，强调稳定读题、连续拿分和中段提速。",
    rankSummary: "尚未开始",
    remaining: "1 天 03 小时后开始",
    recentSubmissions: [],
    problems: [
      { code: "A", title: "待公布", difficulty: "EASY", status: "LOCKED" },
      { code: "B", title: "待公布", difficulty: "MEDIUM", status: "LOCKED" },
      { code: "C", title: "待公布", difficulty: "MEDIUM", status: "LOCKED" },
      { code: "D", title: "待公布", difficulty: "HARD", status: "LOCKED" },
      { code: "E", title: "待公布", difficulty: "HARD", status: "LOCKED" },
      { code: "F", title: "待公布", difficulty: "HARD", status: "LOCKED" },
      { code: "G", title: "待公布", difficulty: "HARD", status: "LOCKED" },
    ],
  },
  {
    id: "april-grand-prix",
    slug: "april-grand-prix",
    title: "四月大奖赛",
    status: "ENDED",
    startsAt: "2026-04-27 19:00",
    endsAt: "2026-04-27 22:00",
    duration: "3 小时",
    problemCount: 8,
    participantCount: 1984,
    blurb: "月赛题量更长，数据结构和后程拉开分差的压力更明显。",
    rankSummary: "最终第 118 / 1,984 名",
    remaining: "比赛已结束",
    recentSubmissions: [
      {
        id: "sub-8870",
        problemCode: "E",
        status: "TIME_LIMIT_EXCEEDED",
        at: "最后一小时",
      },
    ],
    problems: [
      {
        code: "A",
        slug: "two-sum",
        title: "校验和",
        difficulty: "EASY",
        status: "SOLVED",
        firstSolve: "00:08",
      },
      {
        code: "B",
        slug: "shortest-path",
        title: "轨道",
        difficulty: "MEDIUM",
        status: "SOLVED",
        firstSolve: "00:25",
      },
      {
        code: "C",
        slug: "dynamic-ranking",
        title: "拆分合并",
        difficulty: "MEDIUM",
        status: "SOLVED",
        firstSolve: "00:51",
      },
      { code: "D", title: "归档队列", difficulty: "HARD", status: "ATTEMPTED" },
      { code: "E", title: "热寂", difficulty: "HARD", status: "ATTEMPTED" },
      { code: "F", title: "遥测", difficulty: "HARD", status: "LOCKED" },
      { code: "G", title: "镜湖", difficulty: "HARD", status: "LOCKED" },
      { code: "H", title: "棱镜", difficulty: "HARD", status: "LOCKED" },
    ],
  },
];

export async function listContests(): Promise<ContestSummary[]> {
  if (env.apiBaseUrl !== "") {
    try {
      const response = await getJSON<unknown>("/contests");
      return z.array(contestSummarySchema).parse(response);
    } catch {
      if (!env.demoMode) {
        throw missingContestApiError();
      }
    }
  } else if (!env.demoMode) {
    throw missingContestApiError();
  }
  return contestSeed.map((contest) => contestSummarySchema.parse(contest));
}

export async function getContest(
  slug: string,
): Promise<ContestDetail | undefined> {
  if (env.apiBaseUrl !== "") {
    try {
      const response = await getJSON<unknown>(
        `/contests/${encodeURIComponent(slug)}`,
      );
      return contestDetailSchema.parse(response);
    } catch {
      if (!env.demoMode) {
        throw missingContestApiError();
      }
    }
  } else if (!env.demoMode) {
    throw missingContestApiError();
  }
  const found = contestSeed.find((contest) => contest.slug === slug);
  return found ? contestDetailSchema.parse(found) : undefined;
}

function missingContestApiError() {
  return new ApiError(
    env.apiBaseUrl === ""
      ? "contest api is unavailable because VITE_API_BASE_URL is empty and VITE_DEMO_MODE is false"
      : "contest api request failed and VITE_DEMO_MODE is false",
    503,
  );
}
