import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { ApiError } from "../../lib/http-client";
import { LocaleProvider } from "../../lib/locale";
import { ContestMakeupRoute } from "./makeup";

vi.mock("../../features/auth/hooks", () => ({
  useAuthUser: vi.fn(() => ({ data: null })),
}));

vi.mock("../../features/contest-makeup/hooks", () => ({
  useContestMakeupList: vi.fn(),
}));

import { useContestMakeupList } from "../../features/contest-makeup/hooks";

describe("ContestMakeupRoute", () => {
  beforeEach(() => {
    vi.mocked(useContestMakeupList).mockReturnValue({
      data: {
        contestSlug: "april-grand-prix",
        generatedAt: "2026-05-26T09:00:00Z",
        items: [
          {
            problemId: "p-1",
            problemCode: "D",
            problemSlug: "archive-queue",
            problemTitle: "归档队列",
            difficulty: "MEDIUM",
            category: "ATTEMPTED_UNSOLVED",
            attemptCount: 2,
            severityRank: 3,
            reasonSummary: "比赛中尝试过但尚未通过",
            suggestedAction: "优先复查边界条件并重提",
          },
        ],
      },
      error: null,
      isPending: false,
    } as unknown as ReturnType<typeof useContestMakeupList>);
  });

  it("renders makeup items when list has data", () => {
    render(
      <LocaleProvider>
        <QueryClientProvider client={new QueryClient()}>
          <ContestMakeupRoute slug="april-grand-prix" />
        </QueryClientProvider>
      </LocaleProvider>,
    );

    expect(screen.getByText("D. 归档队列")).toBeInTheDocument();
    expect(screen.getByText(/建议动作/)).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "打开题目" })).toHaveAttribute(
      "href",
      "/problems/archive-queue?fromMakeup=april-grand-prix",
    );
  });

  it("renders empty state when list has no items", () => {
    vi.mocked(useContestMakeupList).mockReturnValue({
      data: {
        contestSlug: "april-grand-prix",
        generatedAt: "2026-05-26T09:00:00Z",
        items: [],
      },
      error: null,
      isPending: false,
    } as unknown as ReturnType<typeof useContestMakeupList>);

    render(
      <LocaleProvider>
        <QueryClientProvider client={new QueryClient()}>
          <ContestMakeupRoute slug="april-grand-prix" />
        </QueryClientProvider>
      </LocaleProvider>,
    );

    expect(screen.getByText("本场比赛题目已全部通过，无需补题。")).toBeInTheDocument();
  });

  it("renders not-ended state when contest is not ended", () => {
    vi.mocked(useContestMakeupList).mockReturnValue({
      data: undefined,
      error: new ApiError("contest is not ended", 409),
      isPending: false,
    } as unknown as ReturnType<typeof useContestMakeupList>);

    render(
      <LocaleProvider>
        <QueryClientProvider client={new QueryClient()}>
          <ContestMakeupRoute slug="spring-open" />
        </QueryClientProvider>
      </LocaleProvider>,
    );

    expect(screen.getByText("比赛尚未结束，暂时无法生成补题清单。")).toBeInTheDocument();
  });
});
