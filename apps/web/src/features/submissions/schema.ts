import { z } from "zod";

import { createSubmissionSchema, languageSchema } from "@oj/shared";

export const createSubmissionRequestSchema = createSubmissionSchema.extend({
  userId: z.string().min(1),
});

export const submissionSummarySchema = z.object({
  id: z.string().min(1),
  userId: z.string().min(1),
  problemId: z.string().min(1),
  problem: z
    .object({
      id: z.string().min(1),
      slug: z.string().min(1),
      title: z.string().min(1),
    })
    .optional(),
  language: languageSchema,
  sourceObjectKey: z.string().min(1),
  status: z.string().min(1),
  createdAt: z.string().datetime().optional(),
});

export const submissionDetailSchema = submissionSummarySchema.extend({
  compileSummary: z.object({
    compileOutput: z.string().optional().default(""),
    maxTimeMs: z.number().int().nonnegative().optional(),
    maxMemoryKb: z.number().int().nonnegative().optional(),
    judgedAt: z.string().datetime().optional(),
  }),
  results: z.array(
    z.object({
      testCaseId: z.string().min(1),
      status: z.string().min(1),
      timeMs: z.number().int().nonnegative().optional(),
      memoryKb: z.number().int().nonnegative().optional(),
      outputSnippet: z.string().optional(),
      errorSnippet: z.string().optional(),
    }),
  ),
  artifactAvailability: z.object({
    sourceObjectKey: z.string().optional().default(""),
    artifacts: z.array(
      z.object({
        artifactType: z.string().min(1),
        objectKey: z.string().min(1),
        contentType: z.string().optional(),
        contentEncoding: z.string().optional(),
        expiresAt: z.string().datetime().optional(),
      }),
    ),
  }),
});

export const submissionListSchema = z.object({
  items: z.array(submissionSummarySchema),
  total: z.number().int().nonnegative(),
  page: z.number().int().positive(),
  pageSize: z.number().int().positive(),
});

export const submissionLogContextSchema = z.object({
  event: z.string().min(1),
  feature: z.literal("submissions"),
  route: z.literal("/submissions"),
  requestId: z.string().min(1),
  apiMode: z.enum(["remote", "fallback"]),
  language: languageSchema,
  userIdPresent: z.boolean(),
  problemIdPresent: z.boolean(),
  userIdPreview: z.string().optional(),
  problemIdPreview: z.string().optional(),
  submissionId: z.string().optional(),
  status: z.string().optional(),
  sourceObjectKey: z.string().optional(),
  errorMessage: z.string().optional(),
  httpStatus: z.number().int().optional(),
  field: z.enum(["userId", "problemId", "language", "source"]).optional(),
  sourceLength: z.number().int().nonnegative().optional(),
});

export type CreateSubmissionRequest = z.infer<
  typeof createSubmissionRequestSchema
>;
export type SubmissionSummary = z.infer<typeof submissionSummarySchema>;
export type SubmissionDetail = z.infer<typeof submissionDetailSchema>;
export type SubmissionList = z.infer<typeof submissionListSchema>;
export type SubmissionLogContext = z.infer<typeof submissionLogContextSchema>;
