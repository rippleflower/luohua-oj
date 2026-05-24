import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { LocaleProvider } from "../../lib/locale";
import { ProblemDetailRoute } from "./detail";

describe("ProblemDetailRoute", () => {
  it("renders local fallback detail data", async () => {
    const queryClient = new QueryClient();

    render(
      <LocaleProvider>
        <QueryClientProvider client={queryClient}>
          <ProblemDetailRoute slug="two-sum" />
        </QueryClientProvider>
      </LocaleProvider>,
    );

    expect(
      await screen.findByRole("heading", { name: "Two Sum", level: 1 }),
    ).toBeInTheDocument();
    expect(
      screen.getByText("给定一个整数数组和目标值，找出和为目标值的两个下标。"),
    ).toBeInTheDocument();
    expect(
      screen.getByText("problems/two-sum/versions/1/public/sample-1.in"),
    ).toBeInTheDocument();
  });
});
