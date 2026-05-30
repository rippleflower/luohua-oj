import { render, screen } from "@testing-library/react";
import { QueryClientProvider } from "@tanstack/react-query";
import { vi } from "vitest";

vi.mock("./routes/contests", () => ({
  ContestsRoute: () => <div>比赛管理页面</div>,
}));

import { App } from "./app";
import { queryClient } from "./lib/query-client";

test("renders admin login route", () => {
  window.history.pushState({}, "", "/login");
  render(
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>,
  );
  expect(screen.getByText("管理后台登录")).toBeInTheDocument();
});

test("routes contests path to contests route", () => {
  window.history.pushState({}, "", "/contests");
  render(
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>,
  );
  expect(screen.getByText("比赛管理页面")).toBeInTheDocument();
});
