export const languages = ["CPP17", "CPP20", "JAVA17", "PYTHON311"] as const;
export type Language = (typeof languages)[number];

export const submissionStatuses = [
  "PENDING",
  "RUNNING",
  "ACCEPTED",
  "WRONG_ANSWER",
  "TIME_LIMIT_EXCEEDED",
  "MEMORY_LIMIT_EXCEEDED",
  "RUNTIME_ERROR",
  "COMPILE_ERROR",
  "SYSTEM_ERROR",
] as const;

export type SubmissionStatus = (typeof submissionStatuses)[number];
