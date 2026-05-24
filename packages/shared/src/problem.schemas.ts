import { z } from "zod";

const problemDifficultySchema = z.enum(["EASY", "MEDIUM", "HARD"]);

export const problemSummarySchema = z.object({
  id: z.string(),
  slug: z.string(),
  title: z.string(),
  difficulty: problemDifficultySchema,
  tags: z.array(z.string()),
  acceptedRate: z.number().min(0).max(100),
});

export const problemDetailSchema = z.object({
  id: z.string(),
  slug: z.string(),
  title: z.string(),
  difficulty: problemDifficultySchema,
  statementJson: z.array(
    z.object({
      kind: z.string().min(1),
      section: z.string().min(1),
      content: z.string().min(1),
    }),
  ),
  samplesJson: z.array(
    z.object({
      inputObjectKey: z.string().min(1),
      outputObjectKey: z.string().min(1),
      weight: z.number().int().positive(),
    }),
  ),
  limitsJson: z.object({
    timeLimitMs: z.number().int().positive(),
    memoryLimitKb: z.number().int().positive(),
  }),
  metadataJson: z.record(z.string(), z.unknown()),
  updatedAt: z.string().datetime(),
});

const adminProblemMutationInputBaseSchema = z.object({
  slug: z.string().trim().min(1).max(120),
  title: z.string().trim().min(1).max(160),
  difficulty: problemDifficultySchema,
  timeLimitMs: z.number().int().positive(),
  memoryLimitKb: z.number().int().positive(),
  reason: z.string().trim().max(280).default(""),
});

export const adminProblemCreateInputSchema = adminProblemMutationInputBaseSchema;
export const adminProblemUpdateInputSchema = adminProblemMutationInputBaseSchema;
export const adminProblemPublishInputSchema = z.object({
  reason: z.string().trim().max(280).default(""),
});

export type ProblemSummary = z.infer<typeof problemSummarySchema>;
export type ProblemDetail = z.infer<typeof problemDetailSchema>;
export type AdminProblemCreateInput = z.infer<typeof adminProblemCreateInputSchema>;
export type AdminProblemUpdateInput = z.infer<typeof adminProblemUpdateInputSchema>;
export type AdminProblemPublishInput = z.infer<typeof adminProblemPublishInputSchema>;
