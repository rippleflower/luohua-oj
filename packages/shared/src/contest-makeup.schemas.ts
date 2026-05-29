import { z } from "zod";

const makeupCategorySchema = z.enum([
  "ATTEMPTED_UNSOLVED",
  "UNATTEMPTED_RECOMMENDED",
]);

export const contestMakeupItemSchema = z.object({
  problemId: z.string().min(1),
  problemCode: z.string().min(1),
  problemSlug: z.string().min(1).optional(),
  problemTitle: z.string().min(1),
  difficulty: z.enum(["EASY", "MEDIUM", "HARD"]),
  category: makeupCategorySchema,
  lastStatus: z.string().optional(),
  attemptCount: z.number().int().positive().optional(),
  severityRank: z.number().int().nonnegative(),
  reasonSummary: z.string().min(1),
  suggestedAction: z.string().min(1),
});

export const contestMakeupListSchema = z.object({
  contestSlug: z.string().min(1),
  generatedAt: z.string().datetime(),
  items: z.array(contestMakeupItemSchema),
});

export type ContestMakeupItem = z.infer<typeof contestMakeupItemSchema>;
export type ContestMakeupList = z.infer<typeof contestMakeupListSchema>;
