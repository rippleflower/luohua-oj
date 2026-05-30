import { z } from "zod";

import { problemDetailSchema, problemSummarySchema } from "@oj/shared";

export const problemListSchema = z.array(problemSummarySchema);
export { problemDetailSchema };
