import type { ProblemDetail, ProblemSummary } from "@oj/shared";

import { env } from "../../lib/env";
import { ApiError, getJSON } from "../../lib/http-client";
import { problemDetailSchema, problemListSchema } from "./schema";

const problemSeed: ProblemSummary[] = [
  {
    id: "two-sum",
    problemNo: 1,
    routeCode: "LOCAL1",
    slug: "two-sum",
    title: "Two Sum",
    difficulty: "EASY",
    tags: ["array", "hash-table"],
    acceptedRate: 62.4,
  },
  {
    id: "shortest-path",
    problemNo: 2,
    routeCode: "LOCAL2",
    slug: "shortest-path",
    title: "Shortest Path",
    difficulty: "MEDIUM",
    tags: ["graph", "dijkstra"],
    acceptedRate: 41.8,
  },
  {
    id: "dynamic-ranking",
    problemNo: 3,
    routeCode: "LOCAL3",
    slug: "dynamic-ranking",
    title: "Dynamic Ranking",
    difficulty: "HARD",
    tags: ["segment-tree", "offline"],
    acceptedRate: 18.9,
  },
];

const problemDetailSeed: ProblemDetail[] = [
  {
    id: "two-sum",
    problemNo: 1,
    routeCode: "LOCAL1",
    slug: "two-sum",
    title: "Two Sum",
    difficulty: "EASY",
    statementJson: [
      {
        kind: "markdown",
        section: "statement",
        content: "给定一个整数数组和目标值，找出和为目标值的两个下标。",
      },
      {
        kind: "markdown",
        section: "input",
        content: "第一行输入目标值，第二行输入数组元素。",
      },
      {
        kind: "markdown",
        section: "output",
        content: "输出任意一组满足条件的下标。",
      },
      {
        kind: "markdown",
        section: "constraints",
        content: "数组长度不超过 1e5，值域在 32 位整数范围内。",
      },
    ],
    samplesJson: [
      {
        input: "9\n2 7 11 15\n",
        output: "0 1\n",
        weight: 1,
      },
    ],
    limitsJson: {
      timeLimitMs: 1000,
      memoryLimitKb: 262144,
    },
    metadataJson: {
      tags: ["array", "hash-table"],
      acceptedRate: 62.4,
      published: true,
    },
    updatedAt: "2026-05-17T00:00:00Z",
  },
  {
    id: "shortest-path",
    problemNo: 2,
    routeCode: "LOCAL2",
    slug: "shortest-path",
    title: "Shortest Path",
    difficulty: "MEDIUM",
    statementJson: [
      {
        kind: "markdown",
        section: "statement",
        content: "给定带权图，求从起点到所有点的最短路径。",
      },
      {
        kind: "markdown",
        section: "input",
        content: "输入点数、边数和边信息。",
      },
      {
        kind: "markdown",
        section: "output",
        content: "输出每个点的最短距离。",
      },
      {
        kind: "markdown",
        section: "constraints",
        content: "边权非负，适合 Dijkstra。",
      },
    ],
    samplesJson: [
      {
        input: "4 4 1\n1 2 1\n2 3 2\n1 4 7\n3 4 1\n",
        output: "0 1 3 4\n",
        weight: 1,
      },
    ],
    limitsJson: {
      timeLimitMs: 2000,
      memoryLimitKb: 262144,
    },
    metadataJson: {
      tags: ["graph", "dijkstra"],
      acceptedRate: 41.8,
      published: true,
    },
    updatedAt: "2026-05-17T00:00:00Z",
  },
  {
    id: "dynamic-ranking",
    problemNo: 3,
    routeCode: "LOCAL3",
    slug: "dynamic-ranking",
    title: "Dynamic Ranking",
    difficulty: "HARD",
    statementJson: [
      {
        kind: "markdown",
        section: "statement",
        content: "维护动态序列排名并支持离线查询。",
      },
      {
        kind: "markdown",
        section: "input",
        content: "输入修改和查询混合操作流。",
      },
      {
        kind: "markdown",
        section: "output",
        content: "对每个查询输出对应排名结果。",
      },
      {
        kind: "markdown",
        section: "constraints",
        content: "需要线段树或离线分治。",
      },
    ],
    samplesJson: [
      {
        input: "5 3\n1 5 2 4 3\nQ 1 5 3\nU 3 6\nQ 1 5 3\n",
        output: "3\n4\n",
        weight: 1,
      },
    ],
    limitsJson: {
      timeLimitMs: 4000,
      memoryLimitKb: 524288,
    },
    metadataJson: {
      tags: ["segment-tree", "offline"],
      acceptedRate: 18.9,
      published: true,
    },
    updatedAt: "2026-05-17T00:00:00Z",
  },
];

export async function listProblems(): Promise<ProblemSummary[]> {
  try {
    const response = await getJSON<unknown>("/problems");
    return problemListSchema.parse(response);
  } catch {
    if (env.demoMode) {
      return problemListSchema.parse(problemSeed);
    }
    throw missingProblemApiError();
  }
}

export async function getProblem(
  slug: string,
): Promise<ProblemDetail | undefined> {
  try {
    const response = await getJSON<unknown>(
      `/problems/${encodeURIComponent(slug)}`,
    );
    return problemDetailSchema.parse(response);
  } catch {
    if (!env.demoMode) {
      throw missingProblemApiError();
    }
    const found = problemDetailSeed.find((problem) => problem.slug === slug);
    return found ? problemDetailSchema.parse(found) : undefined;
  }
}

export async function getProblemByRouteCode(
  routeCode: string,
): Promise<ProblemDetail | undefined> {
  try {
    const response = await getJSON<unknown>(
      `/problems/code/${encodeURIComponent(routeCode)}`,
    );
    return problemDetailSchema.parse(response);
  } catch {
    if (!env.demoMode) {
      throw missingProblemApiError();
    }
    const found = problemDetailSeed.find((problem) => problem.routeCode === routeCode);
    return found ? problemDetailSchema.parse(found) : undefined;
  }
}

function missingProblemApiError() {
  return new ApiError(
    env.apiBaseUrl === ""
      ? "problem api is unavailable because VITE_API_BASE_URL is empty and VITE_DEMO_MODE is false"
      : "problem api request failed and VITE_DEMO_MODE is false",
    503,
  );
}
