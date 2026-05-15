import { z } from "zod";

import { problemSummarySchema } from "@oj/shared";

export const problemListSchema = z.array(problemSummarySchema);
