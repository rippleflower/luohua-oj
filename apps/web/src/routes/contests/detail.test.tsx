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
});
