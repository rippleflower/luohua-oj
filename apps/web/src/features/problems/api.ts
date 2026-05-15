import type { ProblemSummary } from "@oj/shared";

import { env } from "../../lib/env";
import { getJSON } from "../../lib/http-client";
import { problemListSchema } from "./schema";

const problemSeed: ProblemSummary[] = [
  {
    id: "two-sum",
    slug: "two-sum",
    title: "Two Sum",
    difficulty: "EASY",
    tags: ["array", "hash-table"],
    acceptedRate: 62.4,
  },
  {
    id: "shortest-path",
    slug: "shortest-path",
    title: "Shortest Path",
    difficulty: "MEDIUM",
    tags: ["graph", "dijkstra"],
    acceptedRate: 41.8,
  },
  {
    id: "dynamic-ranking",
    slug: "dynamic-ranking",
    title: "Dynamic Ranking",
    difficulty: "HARD",
    tags: ["segment-tree", "offline"],
    acceptedRate: 18.9,
  },
];

export async function listProblems(): Promise<ProblemSummary[]> {
  if (env.apiBaseUrl !== "") {
    const response = await getJSON<unknown>("/problems");
    return problemListSchema.parse(response);
  }

  return problemListSchema.parse(problemSeed);
}
