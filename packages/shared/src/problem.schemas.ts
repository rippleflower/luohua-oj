import { z } from "zod";

export const problemSummarySchema = z.object({
  id: z.string(),
  slug: z.string(),
  title: z.string(),
  difficulty: z.enum(["EASY", "MEDIUM", "HARD"]),
  tags: z.array(z.string()),
  acceptedRate: z.number().min(0).max(100),
});

export type ProblemSummary = z.infer<typeof problemSummarySchema>;
