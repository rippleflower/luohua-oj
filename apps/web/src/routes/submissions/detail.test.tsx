import { render, screen } from "@testing-library/react";
import { QueryClientProvider } from "@tanstack/react-query";
import { beforeEach, describe, expect, it } from "vitest";

import { recordSubmissionHistory } from "../../features/submissions/history";
import { LocaleProvider } from "../../lib/locale";
import { queryClient } from "../../lib/query-client";
import { SubmissionDetailRoute } from "./detail";

describe("SubmissionDetailRoute", () => {
  beforeEach(() => {
    window.localStorage.clear();
  });

  it("renders stored submission details", async () => {
    recordSubmissionHistory({
      id: "sub-1",
      userId: "user-1",
      problemId: "two-sum",
      problem: {
        id: "problem-1",
        slug: "two-sum",
        title: "Two Sum",
      },
      language: "CPP17",
      sourceObjectKey: "submissions/2026/05/sub-1/source.cpp.zst",
      status: "PENDING",
    });
    recordSubmissionHistory({
      id: "sub-2",
      userId: "user-2",
      problemId: "binary-search",
      problem: {
        id: "problem-2",
        slug: "binary-search",
        title: "Binary Search",
      },
      language: "CPP17",
      sourceObjectKey: "submissions/2026/05/sub-2/source.cpp.zst",
      status: "ACCEPTED",
    });

    render(
      <QueryClientProvider client={queryClient}>
        <LocaleProvider>
          <SubmissionDetailRoute submissionId="sub-1" />
        </LocaleProvider>
      </QueryClientProvider>,
    );

    expect(await screen.findByText("Two Sum")).toBeInTheDocument();
    expect(screen.getByText("binary-search")).toBeInTheDocument();
    expect(screen.getByText("已通过")).toBeInTheDocument();
    expect(
      screen.getByText("submissions/2026/05/sub-1/source.cpp.zst"),
    ).toBeInTheDocument();
  });
});
