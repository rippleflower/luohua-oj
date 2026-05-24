import { z } from "zod";

import { languages } from "./enums";

export const languageSchema = z.enum(languages);

export const createSubmissionSchema = z.object({
  problemId: z.string().min(1),
  language: languageSchema,
  source: z.string().min(1).max(100_000),
});

export const adminSubmissionRejudgeInputSchema = z.object({
  reason: z.string().trim().max(280).default(""),
});

export const adminSubmissionRejudgeResultSchema = z.object({
  submissionId: z.string().min(1),
  queue: z.string().min(1),
  resultSnapshotVersion: z.number().int().nonnegative(),
  problemVersionId: z.string().min(1).nullable(),
});

export const judgeQueueTaskSummarySchema = z.object({
  id: z.string().min(1),
  type: z.string().min(1),
  submissionId: z.string().min(1).nullable(),
  state: z.enum(["pending", "active", "retry", "completed", "archived"]),
  createdAt: z.string().datetime().nullable(),
  nextProcessAt: z.string().datetime().nullable(),
  completedAt: z.string().datetime().nullable(),
  lastErr: z.string(),
});

export const judgeQueueSummarySchema = z.object({
  queue: z.string().min(1),
  paused: z.boolean(),
  latencySeconds: z.number().int().nonnegative(),
  pending: z.number().int().nonnegative(),
  active: z.number().int().nonnegative(),
  scheduled: z.number().int().nonnegative(),
  retry: z.number().int().nonnegative(),
  archived: z.number().int().nonnegative(),
  completed: z.number().int().nonnegative(),
  processedToday: z.number().int().nonnegative(),
  failedToday: z.number().int().nonnegative(),
  recentTasks: z.array(judgeQueueTaskSummarySchema),
  updatedAt: z.string().datetime(),
});

export type CreateSubmissionInput = z.infer<typeof createSubmissionSchema>;
export type AdminSubmissionRejudgeInput = z.infer<
  typeof adminSubmissionRejudgeInputSchema
>;
export type AdminSubmissionRejudgeResult = z.infer<
  typeof adminSubmissionRejudgeResultSchema
>;
export type JudgeQueueTaskSummary = z.infer<typeof judgeQueueTaskSummarySchema>;
export type JudgeQueueSummary = z.infer<typeof judgeQueueSummarySchema>;
