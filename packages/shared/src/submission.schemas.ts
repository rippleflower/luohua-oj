import { z } from "zod";

import { languages } from "./enums";

export const languageSchema = z.enum(languages);

export const createSubmissionSchema = z.object({
  problemId: z.string().min(1),
  language: languageSchema,
  source: z.string().min(1).max(100_000),
});

export type CreateSubmissionInput = z.infer<typeof createSubmissionSchema>;
