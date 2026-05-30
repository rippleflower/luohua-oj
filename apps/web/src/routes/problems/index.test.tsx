import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { LocaleProvider } from "../../lib/locale";
import { ProblemsRoute } from ".";

describe("ProblemsRoute", () => {
  it("filters problems by search, difficulty, status, and tag", async () => {
    render(
      <LocaleProvider>
        <QueryClientProvider client={new QueryClient()}>
          <ProblemsRoute />
        </QueryClientProvider>
      </LocaleProvider>,
    );

    expect(await screen.findByText("Two Sum")).toBeInTheDocument();
    expect(screen.getByText("Shortest Path")).toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText("搜索标题或 slug"), {
      target: { value: "short" },
    });
    expect(screen.queryByText("Two Sum")).not.toBeInTheDocument();
    expect(screen.getByText("Shortest Path")).toBeInTheDocument();

    fireEvent.change(screen.getByDisplayValue("全部难度"), {
      target: { value: "MEDIUM" },
    });
    expect(screen.getByText("Shortest Path")).toBeInTheDocument();

    fireEvent.click(screen.getAllByRole("button", { name: "dijkstra" })[0]);
    expect(screen.getByText("Shortest Path")).toBeInTheDocument();

    fireEvent.click(screen.getByRole("button", { name: "尝试过 1" }));
    expect(screen.getByText("Shortest Path")).toBeInTheDocument();
  });
});
