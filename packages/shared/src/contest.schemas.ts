import { z } from "zod";

export const contestStatusSchema = z.enum(["RUNNING", "UPCOMING", "ENDED"]);
const auditReasonSchema = z.string().trim().max(280).default("");

export const contestProblemSchema = z.object({
  code: z.string().min(1),
  slug: z.string().min(1).optional(),
  title: z.string().min(1),
  difficulty: z.enum(["EASY", "MEDIUM", "HARD"]),
  status: z.enum(["LOCKED", "ATTEMPTED", "SOLVED"]),
  firstSolve: z.string().optional(),
});

export const contestSummarySchema = z.object({
  id: z.string().min(1),
  slug: z.string().min(1),
  title: z.string().min(1),
  status: contestStatusSchema,
  startsAt: z.string().min(1),
  endsAt: z.string().min(1),
  duration: z.string().min(1),
  problemCount: z.number().int().nonnegative(),
  participantCount: z.number().int().nonnegative(),
  blurb: z.string().min(1),
});

export const contestDetailSchema = contestSummarySchema.extend({
  rankSummary: z.string().min(1),
  remaining: z.string().min(1),
  recentSubmissions: z.array(
    z.object({
      id: z.string().min(1),
      problemCode: z.string().min(1),
      status: z.string().min(1),
      at: z.string().min(1),
    }),
  ),
  problems: z.array(contestProblemSchema),
});

const adminContestMutationInputBaseSchema = z.object({
  slug: z.string().trim().min(1).max(120),
  title: z.string().trim().min(1).max(160),
  description: z.string().trim().max(600).default(""),
  status: contestStatusSchema,
  startsAt: z.string().datetime(),
  endsAt: z.string().datetime(),
  reason: auditReasonSchema,
});

export const adminContestProblemBindingInputSchema = z.object({
  problemId: z.string().min(1),
  code: z.string().trim().min(1).max(16),
  position: z.number().int().positive(),
});

export const adminContestCreateInputSchema =
  adminContestMutationInputBaseSchema;
export const adminContestUpdateInputSchema =
  adminContestMutationInputBaseSchema;
export const adminContestProblemBindingsInputSchema = z.object({
  problems: z.array(adminContestProblemBindingInputSchema),
  reason: auditReasonSchema,
});
export const adminContestFreezeInputSchema = z.object({
  reason: auditReasonSchema,
});

export type ContestSummary = z.infer<typeof contestSummarySchema>;
export type ContestDetail = z.infer<typeof contestDetailSchema>;
export type AdminContestCreateInput = z.infer<
  typeof adminContestCreateInputSchema
>;
export type AdminContestUpdateInput = z.infer<
  typeof adminContestUpdateInputSchema
>;
export type AdminContestProblemBindingInput = z.infer<
  typeof adminContestProblemBindingInputSchema
>;
export type AdminContestProblemBindingsInput = z.infer<
  typeof adminContestProblemBindingsInputSchema
>;
export type AdminContestFreezeInput = z.infer<
  typeof adminContestFreezeInputSchema
>;
