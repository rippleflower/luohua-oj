import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { LocaleProvider } from "../../lib/locale";
import { ContestDetailRoute } from "./detail";

vi.mock("../../features/auth/hooks", () => ({
  useAuthUser: vi.fn(() => ({ data: null })),
}));

vi.mock("../../features/contests/hooks", () => ({
  useContest: vi.fn(),
}));

import { useContest } from "../../features/contests/hooks";

describe("ContestDetailRoute", () => {
  beforeEach(() => {
    vi.mocked(useContest).mockReturnValue({
      data: {
        id: "contest-1",
        slug: "spring-open",
        title: "2026 春季公开赛",
        status: "RUNNING",
        startsAt: "2026-05-16T11:00:00Z",
        endsAt: "2026-05-16T13:00:00Z",
        duration: "2 小时",
        problemCount: 2,
        participantCount: 128,
        blurb: "两小时混合场",
        rankSummary: "当前第 42 / 128 名",
        remaining: "00:30:00",
        recentSubmissions: [
          {
            id: "sub-1",
            problemCode: "A",
            status: "ACCEPTED",
            at: "1 分钟前",
          },
        ],
        problems: [
          {
            code: "A",
            slug: "two-sum",
            title: "两数之和",
            difficulty: "EASY",
            status: "SOLVED",
            firstSolve: "00:12",
          },
          {
            code: "B",
            title: "未公开题目",
            difficulty: "HARD",
            status: "LOCKED",
          },
        ],
      },
    } as unknown as ReturnType<typeof useContest>);
  });

  it("renders real contest overview and removes placeholder statement blocks", () => {
    render(
      <LocaleProvider>
        <QueryClientProvider client={new QueryClient()}>
          <ContestDetailRoute slug="spring-open" />
        </QueryClientProvider>
      </LocaleProvider>,
    );

    expect(
      screen.getByRole("heading", { name: "2026 春季公开赛", level: 1 }),
    ).toBeInTheDocument();
    expect(screen.getByText("两小时混合场")).toBeInTheDocument();
    expect(screen.queryByText("样例输入")).not.toBeInTheDocument();
    expect(screen.queryByText("输入说明")).not.toBeInTheDocument();
  });

  it("links problem card when slug exists and degrades to readonly without slug", () => {
    render(
      <LocaleProvider>
        <QueryClientProvider client={new QueryClient()}>
          <ContestDetailRoute slug="spring-open" />
        </QueryClientProvider>
      </LocaleProvider>,
    );

    const linkedProblem = screen.getByRole("link", { name: /两数之和/i });
    expect(linkedProblem).toHaveAttribute("href", "/problems/two-sum");
    expect(screen.getByText("未公开题目")).toBeInTheDocument();
  });

  it("shows makeup-list action when contest has ended", () => {
    vi.mocked(useContest).mockReturnValue({
      data: {
        id: "contest-1",
        slug: "april-grand-prix",
        title: "四月大奖赛",
        status: "ENDED",
        startsAt: "2026-04-27T11:00:00Z",
        endsAt: "2026-04-27T13:00:00Z",
        duration: "2 小时",
        problemCount: 2,
        participantCount: 128,
        blurb: "已结束比赛",
        rankSummary: "最终第 42 名",
        remaining: "比赛已结束",
        recentSubmissions: [],
        problems: [
          { code: "A", title: "A", difficulty: "EASY", status: "SOLVED" },
          { code: "B", title: "B", difficulty: "MEDIUM", status: "ATTEMPTED" },
        ],
      },
    } as unknown as ReturnType<typeof useContest>);

    render(
      <LocaleProvider>
        <QueryClientProvider client={new QueryClient()}>
          <ContestDetailRoute slug="april-grand-prix" />
        </QueryClientProvider>
      </LocaleProvider>,
    );

    const button = screen.getByRole("link", { name: "生成补题清单" });
    expect(button).toHaveAttribute("href", "/contests/april-grand-prix/makeup-list");
  });
});
