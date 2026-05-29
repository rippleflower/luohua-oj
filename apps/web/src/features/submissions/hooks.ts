import { useEffect } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { createSubmission } from "./api";
import { recordSubmissionHistory } from "./history";
import {
  logSubmissionFailed,
  logSubmissionFallback,
  logSubmissionStarted,
  logSubmissionSucceeded,
} from "./logging";
import type { CreateSubmissionRequest } from "./schema";
import { env } from "../../lib/env";
import { getSubmission, listSubmissions } from "./read";

export function useCreateSubmission() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (input: { form: CreateSubmissionRequest; requestId: string }) => {
      const apiMode = env.demoMode && env.apiBaseUrl === "" ? "fallback" : "remote";
      logSubmissionStarted(input.requestId, apiMode, input.form);
      if (apiMode === "fallback") {
        logSubmissionFallback(input.requestId, input.form);
      }

      return createSubmission(input.form);
    },
    onSuccess: (response, variables) => {
      const apiMode = env.demoMode && env.apiBaseUrl === "" ? "fallback" : "remote";
      recordSubmissionHistory(response);
      void queryClient.invalidateQueries({ queryKey: ["submissions", "history"] });
      void queryClient.invalidateQueries({ queryKey: ["submission", response.id] });
      logSubmissionSucceeded(variables.requestId, apiMode, variables.form, response);
    },
    onError: (error, variables) => {
      const apiMode = env.demoMode && env.apiBaseUrl === "" ? "fallback" : "remote";
      logSubmissionFailed(variables.requestId, apiMode, variables.form, error);
    },
  });
}

export function useSubmissionHistory(username: string, page: number, pageSize: number) {
  const normalizedUsername = username.trim();
  const apiMode = env.demoMode && env.apiBaseUrl === "" ? "fallback" : "remote";
  const queryClient = useQueryClient();

  useEffect(() => {
    const syncHistory = () => {
      void queryClient.invalidateQueries({ queryKey: ["submissions", "history"] });
    };

    window.addEventListener("submissions:history-updated", syncHistory);
    return () => {
      window.removeEventListener("submissions:history-updated", syncHistory);
    };
  }, [queryClient]);

  return useQuery({
    queryKey: ["submissions", "history", normalizedUsername, page, pageSize, apiMode],
    queryFn: () => listSubmissions(normalizedUsername, { page, pageSize }),
  });
}

export function useSubmissionDetail(submissionId: string) {
  const apiMode = env.demoMode && env.apiBaseUrl === "" ? "fallback" : "remote";

  return useQuery({
    queryKey: ["submission", submissionId, apiMode],
    queryFn: () => getSubmission(submissionId),
    enabled: submissionId.trim() !== "",
    refetchInterval(query) {
      const submission = query.state.data;
      if (apiMode !== "remote") {
        return false;
      }
      if (!submission) {
        return 2_000;
      }
      return submission.status === "PENDING" || submission.status === "RUNNING" ? 2_000 : false;
    },
    refetchIntervalInBackground: false,
  });
}
