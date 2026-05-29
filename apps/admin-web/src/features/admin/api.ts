import type {
  AdminProblemContentInput,
  AdminProblemDetail,
  AdminContestCreateInput,
  AdminContestFreezeInput,
  AdminContestProblemBindingsInput,
  AdminProblemCreateInput,
  AdminProblemPublishInput,
  AdminAuditEvent,
  AdminContestSummary,
  AdminContestUpdateInput,
  AdminProblemSummary,
  AdminSubmissionRejudgeInput,
  AdminSubmissionRejudgeResult,
  AdminProblemUpdateInput,
  AdminSubmissionSummary,
  AdminUserDetail,
  AdminUserSummary,
  Announcement,
  JudgeQueueSummary,
  PermissionKey,
  SystemSettings,
} from "@oj/shared";
import {
  adminContestCreateInputSchema,
  adminContestFreezeInputSchema,
  adminContestProblemBindingsInputSchema,
  adminProblemContentInputSchema,
  adminProblemCreateInputSchema,
  adminProblemDetailSchema,
  adminProblemPublishInputSchema,
  adminProblemUpdateInputSchema,
  adminContestUpdateInputSchema,
  adminSubmissionRejudgeInputSchema,
  adminSubmissionRejudgeResultSchema,
  adminAuditEventSchema,
  adminContestSummarySchema,
  adminProblemSummarySchema,
  adminSubmissionSummarySchema,
  adminUserDetailSchema,
  adminUserSummarySchema,
  announcementSchema,
  judgeQueueSummarySchema,
  systemSettingsSchema,
} from "@oj/shared";

import { getJSON, patchJSON, postJSON, putJSON } from "../../lib/http-client";

export async function listUsers(): Promise<AdminUserSummary[]> {
  const response = await getJSON<unknown[]>("/admin/users");
  return response.map((item) => adminUserSummarySchema.parse(item));
}

export async function getUser(userId: string): Promise<AdminUserDetail> {
  const response = await getJSON<unknown>(`/admin/users/${userId}`);
  return adminUserDetailSchema.parse(response);
}

export async function getUserPermissions(
  userId: string,
): Promise<PermissionKey[]> {
  return getJSON<PermissionKey[]>(`/admin/users/${userId}/permissions`);
}

export async function updateUser(
  userId: string,
  input: {
    status: string;
    displayName: string;
    bio: string;
    avatarUrl: string;
    reason: string;
  },
) {
  const response = await patchJSON<unknown>(`/admin/users/${userId}`, {
    body: input,
  });
  return adminUserDetailSchema.parse(response);
}

export async function updateUserRole(
  userId: string,
  role: string,
  reason: string,
) {
  await postJSON(`/admin/users/${userId}/role`, { body: { role, reason } });
}

export async function updateUserPermissions(
  userId: string,
  permissions: PermissionKey[],
  reason: string,
) {
  await putJSON(`/admin/users/${userId}/permissions`, {
    body: { permissions, reason },
  });
}

export async function listProblems(): Promise<AdminProblemSummary[]> {
  const response = await getJSON<unknown[]>("/admin/problems");
  return response.map((item) => adminProblemSummarySchema.parse(item));
}

export async function createProblem(
  input: AdminProblemCreateInput,
): Promise<AdminProblemSummary> {
  const response = await postJSON<unknown>("/admin/problems", {
    body: adminProblemCreateInputSchema.parse(input),
  });
  return adminProblemSummarySchema.parse(response);
}

export async function updateProblem(
  problemId: string,
  input: AdminProblemUpdateInput,
): Promise<AdminProblemSummary> {
  const response = await patchJSON<unknown>(`/admin/problems/${problemId}`, {
    body: adminProblemUpdateInputSchema.parse(input),
  });
  return adminProblemSummarySchema.parse(response);
}

export async function getProblemDetail(
  problemId: string,
): Promise<AdminProblemDetail> {
  const response = await getJSON<unknown>(`/admin/problems/${problemId}`);
  return adminProblemDetailSchema.parse(response);
}

export async function updateProblemContent(
  problemId: string,
  input: AdminProblemContentInput,
): Promise<AdminProblemDetail> {
  const response = await patchJSON<unknown>(`/admin/problems/${problemId}/content`, {
    body: adminProblemContentInputSchema.parse(input),
  });
  return adminProblemDetailSchema.parse(response);
}

export async function publishProblem(
  problemId: string,
  input: AdminProblemPublishInput,
): Promise<AdminProblemSummary> {
  const response = await postJSON<unknown>(
    `/admin/problems/${problemId}/publish`,
    {
      body: adminProblemPublishInputSchema.parse(input),
    },
  );
  return adminProblemSummarySchema.parse(response);
}

export async function listContests(): Promise<AdminContestSummary[]> {
  const response = await getJSON<unknown[]>("/admin/contests");
  return response.map((item) => adminContestSummarySchema.parse(item));
}

export async function createContest(
  input: AdminContestCreateInput,
): Promise<AdminContestSummary> {
  const response = await postJSON<unknown>("/admin/contests", {
    body: adminContestCreateInputSchema.parse(input),
  });
  return adminContestSummarySchema.parse(response);
}

export async function updateContest(
  contestId: string,
  input: AdminContestUpdateInput,
): Promise<AdminContestSummary> {
  const response = await patchJSON<unknown>(`/admin/contests/${contestId}`, {
    body: adminContestUpdateInputSchema.parse(input),
  });
  return adminContestSummarySchema.parse(response);
}

export async function replaceContestProblems(
  contestId: string,
  input: AdminContestProblemBindingsInput,
): Promise<AdminContestSummary> {
  const response = await putJSON<unknown>(
    `/admin/contests/${contestId}/problems`,
    {
      body: adminContestProblemBindingsInputSchema.parse(input),
    },
  );
  return adminContestSummarySchema.parse(response);
}

export async function freezeContest(
  contestId: string,
  input: AdminContestFreezeInput,
): Promise<AdminContestSummary> {
  const response = await postJSON<unknown>(
    `/admin/contests/${contestId}/freeze`,
    {
      body: adminContestFreezeInputSchema.parse(input),
    },
  );
  return adminContestSummarySchema.parse(response);
}

export async function listSubmissions(): Promise<AdminSubmissionSummary[]> {
  const response = await getJSON<unknown[]>("/admin/submissions");
  return response.map((item) => adminSubmissionSummarySchema.parse(item));
}

export async function getJudgeQueueSummary(): Promise<JudgeQueueSummary> {
  const response = await getJSON<unknown>("/admin/judge/queue");
  return judgeQueueSummarySchema.parse(response);
}

export async function rejudgeSubmission(
  submissionId: string,
  input: AdminSubmissionRejudgeInput,
): Promise<AdminSubmissionRejudgeResult> {
  const response = await postJSON<unknown>(
    `/admin/submissions/${submissionId}/rejudge`,
    {
      body: adminSubmissionRejudgeInputSchema.parse(input),
    },
  );
  return adminSubmissionRejudgeResultSchema.parse(response);
}

export async function listAuditEvents(): Promise<AdminAuditEvent[]> {
  const response = await getJSON<unknown[]>("/admin/audit");
  return response.map((item) => adminAuditEventSchema.parse(item));
}

export async function getSystemSettings(): Promise<SystemSettings> {
  const response = await getJSON<unknown>("/admin/system/settings");
  return systemSettingsSchema.parse(response);
}

export async function updateSystemSettings(input: {
  registrationEnabled: boolean;
  judgeQueuePaused: boolean;
  storageMode: "LOCAL" | "S3";
}) {
  const response = await patchJSON<unknown>("/admin/system/settings", {
    body: input,
  });
  return systemSettingsSchema.parse(response);
}

export async function listAnnouncements(): Promise<Announcement[]> {
  const response = await getJSON<unknown[]>("/admin/announcements");
  return response.map((item) => announcementSchema.parse(item));
}

export async function createAnnouncement(input: {
  title: string;
  content: string;
  status: "DRAFT" | "PUBLISHED";
  audience: "ALL" | "USERS" | "ADMINS";
}) {
  const response = await postJSON<unknown>("/admin/announcements", {
    body: input,
  });
  return announcementSchema.parse(response);
}

export async function updateAnnouncement(
  id: string,
  input: {
    title: string;
    content: string;
    status: "DRAFT" | "PUBLISHED";
    audience: "ALL" | "USERS" | "ADMINS";
  },
) {
  const response = await patchJSON<unknown>(`/admin/announcements/${id}`, {
    body: input,
  });
  return announcementSchema.parse(response);
}
