import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { LocaleProvider } from "../../lib/locale";
import { SubmissionHistoryList } from "./submission-history-list";

describe("SubmissionHistoryList", () => {
  it("renders history rows", () => {
    render(
      <LocaleProvider>
        <SubmissionHistoryList
          submissions={[
            {
              id: "submission-1234",
              userId: "user-1",
              problemId: "two-sum",
              problem: {
                id: "problem-1",
                slug: "two-sum",
                title: "Two Sum",
              },
              language: "CPP17",
              sourceObjectKey:
                "submissions/2026/05/submission-1234/source.cpp.zst",
              status: "PENDING",
            },
          ]}
          page={1}
          pageSize={20}
          sourceLabel="显示本地提交记录"
          total={1}
          username="demo"
          onUsernameChange={() => {}}
        />
      </LocaleProvider>,
    );

    expect(screen.getByText("Two Sum")).toBeInTheDocument();
    expect(screen.getByText("two-sum")).toBeInTheDocument();
    expect(screen.getByText("CPP17")).toBeInTheDocument();
    expect(screen.getByText("等待中")).toBeInTheDocument();
    expect(screen.getByDisplayValue("demo")).toBeInTheDocument();
    expect(screen.getByText("1 条提交 · 第 1 / 1 页")).toBeInTheDocument();
  });
});
