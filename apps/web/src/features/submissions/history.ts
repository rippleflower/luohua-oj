import { submissionSummarySchema, type SubmissionSummary } from "./schema";

const storageKey = "oj.submissions.history";
const maxHistorySize = 20;

function canUseStorage() {
  return typeof window !== "undefined" && typeof window.localStorage !== "undefined";
}

function parseHistory(raw: string | null): SubmissionSummary[] {
  if (!raw) {
    return [];
  }

  try {
    const parsed = JSON.parse(raw) as unknown;
    return Array.isArray(parsed) ? parsed.map((item) => submissionSummarySchema.parse(item)) : [];
  } catch {
    return [];
  }
}

export function listSubmissionHistory(): SubmissionSummary[] {
  if (!canUseStorage()) {
    return [];
  }

  return parseHistory(window.localStorage.getItem(storageKey));
}

export function getSubmissionHistoryItem(submissionId: string): SubmissionSummary | undefined {
  return listSubmissionHistory().find((item) => item.id === submissionId);
}

export function recordSubmissionHistory(submission: SubmissionSummary) {
  if (!canUseStorage()) {
    return;
  }

  const entry = submissionSummarySchema.parse({
    ...submission,
    createdAt: submission.createdAt ?? new Date().toISOString(),
  });

  const next = [
    entry,
    ...listSubmissionHistory().filter((item) => item.id !== entry.id),
  ].slice(0, maxHistorySize);

  window.localStorage.setItem(storageKey, JSON.stringify(next));
  window.dispatchEvent(new CustomEvent("submissions:history-updated"));
}
