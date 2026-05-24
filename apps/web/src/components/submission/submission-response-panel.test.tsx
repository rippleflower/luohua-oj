import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { LocaleProvider } from "../../lib/locale";
import { SubmissionResponsePanel } from "./submission-response-panel";

describe("SubmissionResponsePanel", () => {
  it("renders empty state", () => {
    render(
      <LocaleProvider>
        <SubmissionResponsePanel />
      </LocaleProvider>,
    );
    expect(screen.getByText("还没有提交记录。")).toBeInTheDocument();
  });

  it("renders submission data", () => {
    render(
      <LocaleProvider>
        <SubmissionResponsePanel
          data={{
            id: "sub-1",
            userId: "user-1",
            problemId: "problem-1",
            language: "CPP17",
            sourceObjectKey: "submissions/2026/05/sub-1/source.cpp.zst",
            status: "PENDING",
            compileSummary: {
              compileOutput: "",
            },
            results: [],
            artifactAvailability: {
              sourceObjectKey: "submissions/2026/05/sub-1/source.cpp.zst",
              artifacts: [],
            },
          }}
        />
      </LocaleProvider>,
    );

    expect(screen.getByText("sub-1")).toBeInTheDocument();
    expect(screen.getByText("等待中")).toBeInTheDocument();
    expect(
      screen.getByText("submissions/2026/05/sub-1/source.cpp.zst"),
    ).toBeInTheDocument();
  });
});
