import { useMutation } from "@tanstack/react-query";

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

export function useCreateSubmission() {
  return useMutation({
    mutationFn: async (input: { form: CreateSubmissionRequest; requestId: string }) => {
      const apiMode = env.apiBaseUrl !== "" ? "remote" : "fallback";
      logSubmissionStarted(input.requestId, apiMode, input.form);
      if (apiMode === "fallback") {
        logSubmissionFallback(input.requestId, input.form);
      }

      return createSubmission(input.form);
    },
    onSuccess: (response, variables) => {
      const apiMode = env.apiBaseUrl !== "" ? "remote" : "fallback";
      recordSubmissionHistory(response);
      logSubmissionSucceeded(variables.requestId, apiMode, variables.form, response);
    },
    onError: (error, variables) => {
      const apiMode = env.apiBaseUrl !== "" ? "remote" : "fallback";
      logSubmissionFailed(variables.requestId, apiMode, variables.form, error);
    },
  });
}
