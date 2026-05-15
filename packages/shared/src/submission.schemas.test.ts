import { describe, expect, it } from "vitest";

import { createSubmissionSchema } from "./submission.schemas";

describe("createSubmissionSchema", () => {
  it("rejects unsupported languages", () => {
    expect(() =>
      createSubmissionSchema.parse({
        problemId: "problem-1",
        language: "RUBY",
        source: "puts 1",
      }),
    ).toThrow();
  });

  it("accepts a supported language and source", () => {
    const result = createSubmissionSchema.parse({
      problemId: "problem-1",
      language: "CPP17",
      source: "int main(){return 0;}",
    });

    expect(result.language).toBe("CPP17");
  });
});
