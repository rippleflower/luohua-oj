import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { LocaleProvider } from "../lib/locale";
import { HomeRoute } from "./home";

describe("HomeRoute", () => {
  it("renders key entry points", async () => {
    const queryClient = new QueryClient();

    render(
      <LocaleProvider>
        <QueryClientProvider client={queryClient}>
          <HomeRoute />
        </QueryClientProvider>
      </LocaleProvider>,
    );

    expect(await screen.findByText("面向刷题与比赛的在线判题台")).toBeInTheDocument();
    expect(screen.getAllByText("开始做题").length).toBeGreaterThan(0);
    expect(screen.getAllByText("查看比赛").length).toBeGreaterThan(0);
    expect(screen.getAllByText("题库").length).toBeGreaterThan(0);
    expect(screen.getAllByText("比赛").length).toBeGreaterThan(0);
    expect(screen.getAllByText("最近提交").length).toBeGreaterThan(0);
  });
});
