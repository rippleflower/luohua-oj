export const languages = ["CPP17", "CPP20", "JAVA17", "PYTHON311"] as const;
export type Language = (typeof languages)[number];

export const userRoles = ["USER", "ADMIN", "SUPER_ADMIN"] as const;
export type UserRole = (typeof userRoles)[number];

export const permissionKeys = [
  "dashboard.view",
  "users.view",
  "users.edit",
  "users.roles",
  "problems.view",
  "problems.edit",
  "problems.publish",
  "contests.view",
  "contests.edit",
  "contests.publish",
  "submissions.view",
  "submissions.rejudge",
  "announcements.view",
  "announcements.edit",
  "system.view",
  "system.edit",
  "audit.view",
] as const;
export type PermissionKey = (typeof permissionKeys)[number];

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
