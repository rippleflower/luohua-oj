import { env } from "../../lib/env";
import { ApiError, getJSON } from "../../lib/http-client";
import { getSubmissionHistoryItem, listSubmissionHistory } from "./history";
import {
  submissionDetailSchema,
  submissionListSchema,
  type SubmissionDetail,
  type SubmissionList,
} from "./schema";

export async function getSubmission(
  submissionId: string,
): Promise<SubmissionDetail | undefined> {
  if (env.apiBaseUrl !== "") {
    const response = await getJSON<unknown>(`/submissions/${submissionId}`);
    return submissionDetailSchema.parse(response);
  }

  if (!env.demoMode) {
    throw new ApiError(
      "submission detail api is unavailable because VITE_API_BASE_URL is empty and VITE_DEMO_MODE is false",
      503,
    );
  }

  const fallback = getSubmissionHistoryItem(submissionId);
  if (!fallback) {
    return undefined;
  }

  return submissionDetailSchema.parse({
    ...fallback,
    compileSummary: {
      compileOutput: "",
    },
    results: [],
    artifactAvailability: {
      sourceObjectKey: fallback.sourceObjectKey,
      artifacts: [],
    },
  });
}

type ListSubmissionsOptions = {
  page?: number;
  pageSize?: number;
};

export async function listSubmissions(
  username: string,
  options: ListSubmissionsOptions = {},
): Promise<SubmissionList> {
  const normalized = username.trim();
  const page = options.page ?? 1;
  const pageSize = options.pageSize ?? 20;

  if (env.apiBaseUrl !== "" && normalized !== "") {
    const response = await getJSON<unknown>(
      `/users/${encodeURIComponent(normalized)}/submissions?page=${page}&pageSize=${pageSize}`,
    );
    return submissionListSchema.parse(response);
  }

  if (!env.demoMode) {
    throw new ApiError(
      "submission history api is unavailable because demo fallback is disabled",
      503,
    );
  }

  const items = listSubmissionHistory();
  return {
    items,
    total: items.length,
    page: 1,
    pageSize: items.length === 0 ? pageSize : items.length,
  };
}
