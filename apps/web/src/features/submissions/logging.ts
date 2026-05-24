import { logError, logEvent, logRequest } from "../../lib/logger";
import {
  submissionLogContextSchema,
  type CreateSubmissionRequest,
  type SubmissionLogContext,
  type SubmissionSummary,
} from "./schema";

function previewValue(value: string): string | undefined {
  const trimmed = value.trim();
  if (trimmed === "") {
    return undefined;
  }

  return trimmed.length <= 8 ? trimmed : `${trimmed.slice(0, 8)}...`;
}

export function createSubmissionRequestId() {
  return crypto.randomUUID();
}

function baseContext(
  requestId: string,
  apiMode: SubmissionLogContext["apiMode"],
  form: CreateSubmissionRequest,
): Omit<SubmissionLogContext, "event"> {
  return {
    feature: "submissions",
    route: "/submissions",
    requestId,
    apiMode,
    language: form.language,
    userIdPresent: form.userId.trim() !== "",
    problemIdPresent: form.problemId.trim() !== "",
    userIdPreview: previewValue(form.userId),
    problemIdPreview: previewValue(form.problemId),
  };
}

export function logSubmissionPageViewed(
  requestId: string,
  apiMode: SubmissionLogContext["apiMode"],
  form: CreateSubmissionRequest,
) {
  logEvent(
    submissionLogContextSchema.parse({
      event: "submissions.page_viewed",
      ...baseContext(requestId, apiMode, form),
    }),
  );
}

export function logSubmissionFieldChanged(
  requestId: string,
  apiMode: SubmissionLogContext["apiMode"],
  form: CreateSubmissionRequest,
  field: NonNullable<SubmissionLogContext["field"]>,
) {
  logEvent(
    submissionLogContextSchema.parse({
      event: "submissions.form_field_changed",
      ...baseContext(requestId, apiMode, form),
      field,
      sourceLength: field === "source" ? form.source.length : undefined,
    }),
  );
}

export function logSubmissionRequested(
  requestId: string,
  apiMode: SubmissionLogContext["apiMode"],
  form: CreateSubmissionRequest,
) {
  logRequest(
    submissionLogContextSchema.parse({
      event: "submissions.submit_clicked",
      ...baseContext(requestId, apiMode, form),
      sourceLength: form.source.length,
    }),
  );
}

export function logSubmissionStarted(
  requestId: string,
  apiMode: SubmissionLogContext["apiMode"],
  form: CreateSubmissionRequest,
) {
  logRequest(
    submissionLogContextSchema.parse({
      event: "submissions.request_started",
      ...baseContext(requestId, apiMode, form),
      sourceLength: form.source.length,
    }),
  );
}

export function logSubmissionFallback(
  requestId: string,
  form: CreateSubmissionRequest,
) {
  logEvent(
    submissionLogContextSchema.parse({
      event: "submissions.fallback_used",
      ...baseContext(requestId, "fallback", form),
      sourceLength: form.source.length,
    }),
  );
}

export function logSubmissionSucceeded(
  requestId: string,
  apiMode: SubmissionLogContext["apiMode"],
  form: CreateSubmissionRequest,
  response: SubmissionSummary,
) {
  logEvent(
    submissionLogContextSchema.parse({
      event: "submissions.request_succeeded",
      ...baseContext(requestId, apiMode, form),
      submissionId: response.id,
      status: response.status,
      sourceObjectKey: response.sourceObjectKey,
    }),
  );
}

export function logSubmissionFailed(
  requestId: string,
  apiMode: SubmissionLogContext["apiMode"],
  form: CreateSubmissionRequest,
  error: unknown,
) {
  const message = error instanceof Error ? error.message : "unknown error";
  const status = extractStatus(message);
  logError(
    submissionLogContextSchema.parse({
      event: "submissions.request_failed",
      ...baseContext(requestId, apiMode, form),
      errorMessage: message,
      httpStatus: status,
    }),
  );
}

export function logSubmissionValidationFailed(
  requestId: string,
  apiMode: SubmissionLogContext["apiMode"],
  form: CreateSubmissionRequest,
  error: unknown,
) {
  const message = error instanceof Error ? error.message : "validation error";
  logError(
    submissionLogContextSchema.parse({
      event: "submissions.validation_failed",
      ...baseContext(requestId, apiMode, form),
      errorMessage: message,
    }),
  );
}

function extractStatus(message: string): number | undefined {
  const matched = message.match(/\b(\d{3})\b/);
  return matched ? Number(matched[1]) : undefined;
}
