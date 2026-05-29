import { z } from "zod";

import {
  authUserSchema,
  permissionKeySchema,
  profileSettingsSchema,
  userRoleSchema,
} from "./auth.schemas";
import { contestStatusSchema } from "./contest.schemas";
import {
  problemSampleSchema,
  problemStatementSectionSchema,
} from "./problem.schemas";

export const adminUserSummarySchema = z.object({
  id: z.string().min(1),
  email: z.string().email(),
  username: z.string().min(1),
  role: userRoleSchema,
  status: z.string().min(1),
  displayName: z.string().min(1),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
});

export const adminUserDetailSchema = adminUserSummarySchema.extend({
  permissions: z.array(permissionKeySchema),
  profile: profileSettingsSchema.extend({
    preferredLocale: z.string().min(2),
    preferredLanguage: z.string().min(1),
  }),
  stats: z.object({
    solvedCount: z.number().int().nonnegative(),
    submissionCount: z.number().int().nonnegative(),
    acceptedCount: z.number().int().nonnegative(),
    lastActiveAt: z.string().datetime().nullable(),
  }),
});

export const adminProblemSummarySchema = z.object({
  id: z.string().min(1),
  problemNo: z.number().int().positive(),
  routeCode: z.string().min(1),
  slug: z.string().min(1),
  title: z.string().min(1),
  difficulty: z.enum(["EASY", "MEDIUM", "HARD"]),
  timeLimitMs: z.number().int().positive(),
  memoryLimitKb: z.number().int().positive(),
  status: z.string().min(1),
  currentVersionNo: z.number().int().positive(),
  isPublished: z.boolean(),
  submissionCount: z.number().int().nonnegative(),
  acceptedRate: z.number().min(0).max(100),
  updatedAt: z.string().datetime(),
});

export const adminProblemDetailSchema = adminProblemSummarySchema.extend({
  statementJson: z.array(problemStatementSectionSchema),
  samples: z.array(problemSampleSchema),
  tags: z.array(z.string().min(1)),
});

export const adminContestSummarySchema = z.object({
  id: z.string().min(1),
  slug: z.string().min(1),
  title: z.string().min(1),
  description: z.string(),
  status: contestStatusSchema,
  startsAt: z.string().datetime(),
  endsAt: z.string().datetime(),
  participantCount: z.number().int().nonnegative(),
  problemCount: z.number().int().nonnegative(),
  latestSnapshotNo: z.number().int().nonnegative(),
  updatedAt: z.string().datetime(),
  problems: z.array(
    z.object({
      problemId: z.string().min(1),
      problemSlug: z.string().min(1),
      problemTitle: z.string().min(1),
      code: z.string().min(1),
      position: z.number().int().positive(),
    }),
  ),
  snapshots: z.array(
    z.object({
      id: z.string().min(1),
      snapshotNo: z.number().int().positive(),
      frozenAt: z.string().datetime(),
      problemCount: z.number().int().nonnegative(),
    }),
  ),
});

export const adminSubmissionSummarySchema = z.object({
  id: z.string().min(1),
  username: z.string().min(1),
  problemTitle: z.string().min(1),
  language: z.string().min(1),
  status: z.string().min(1),
  createdAt: z.string().datetime(),
});

export const adminAuditEventSchema = z.object({
  id: z.string().min(1),
  actor: z.object({
    id: z.string().min(1).nullable(),
    username: z.string().min(1).nullable(),
    role: userRoleSchema.nullable(),
  }),
  action: z.string().min(1),
  targetType: z.string().min(1),
  targetId: z.string().min(1),
  reason: z.string(),
  diff: z.record(z.string(), z.unknown()),
  ip: z.string(),
  createdAt: z.string().datetime(),
});

export const adminDashboardSchema = z.object({
  actor: authUserSchema,
  metrics: z.object({
    totalUsers: z.number().int().nonnegative(),
    activeProblems: z.number().int().nonnegative(),
    runningContests: z.number().int().nonnegative(),
    submissions24h: z.number().int().nonnegative(),
    pendingSubmissions: z.number().int().nonnegative(),
  }),
  alerts: z.array(
    z.object({
      title: z.string().min(1),
      tone: z.enum(["info", "warning", "critical"]),
      description: z.string().min(1),
    }),
  ),
});

export const systemSettingsSchema = z.object({
  registrationEnabled: z.boolean(),
  judgeQueuePaused: z.boolean(),
  storageMode: z.enum(["LOCAL", "S3"]),
  sourceRoot: z.string().min(1),
  redisAddr: z.string().min(1),
  updatedAt: z.string().datetime(),
});

export const announcementSchema = z.object({
  id: z.string().min(1),
  title: z.string().min(1),
  content: z.string().min(1),
  status: z.enum(["DRAFT", "PUBLISHED"]),
  audience: z.enum(["ALL", "USERS", "ADMINS"]),
  createdAt: z.string().datetime(),
  updatedAt: z.string().datetime(),
});

export type AdminUserSummary = z.infer<typeof adminUserSummarySchema>;
export type AdminUserDetail = z.infer<typeof adminUserDetailSchema>;
export type AdminProblemSummary = z.infer<typeof adminProblemSummarySchema>;
export type AdminProblemDetail = z.infer<typeof adminProblemDetailSchema>;
export type AdminContestSummary = z.infer<typeof adminContestSummarySchema>;
export type AdminSubmissionSummary = z.infer<
  typeof adminSubmissionSummarySchema
>;
export type AdminAuditEvent = z.infer<typeof adminAuditEventSchema>;
export type AdminDashboard = z.infer<typeof adminDashboardSchema>;
export type SystemSettings = z.infer<typeof systemSettingsSchema>;
export type Announcement = z.infer<typeof announcementSchema>;
