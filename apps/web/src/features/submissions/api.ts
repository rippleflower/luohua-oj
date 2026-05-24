import { env } from "../../lib/env";
import { postJSON } from "../../lib/http-client";
import {
  createSubmissionRequestSchema,
  submissionSummarySchema,
  type CreateSubmissionRequest,
  type SubmissionSummary,
} from "./schema";

const demoSourceRoot = "tmp/submissions";

export async function createSubmission(
  input: CreateSubmissionRequest,
): Promise<SubmissionSummary> {
  const request = createSubmissionRequestSchema.parse(input);

  if (env.apiBaseUrl !== "") {
    const response = await postJSON<unknown>("/submissions", {
      body: request,
    });
    return submissionSummarySchema.parse(response);
  }

  return submissionSummarySchema.parse({
    id: crypto.randomUUID(),
    userId: request.userId,
    problemId: request.problemId,
    language: request.language,
    sourceObjectKey: `${demoSourceRoot}/${request.problemId}-${Date.now()}.txt`,
    status: "PENDING",
  });
}
