import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { LocaleProvider } from "../../lib/locale";
import { ProblemList } from "./problem-list";

describe("ProblemList", () => {
  it("renders problem rows", () => {
    render(
      <LocaleProvider>
        <ProblemList
          problems={[
            {
              id: "two-sum",
              slug: "two-sum",
              title: "Two Sum",
              difficulty: "EASY",
              tags: ["array"],
              acceptedRate: 60,
              status: "SOLVED",
            },
          ]}
        />
      </LocaleProvider>,
    );

    expect(screen.getByText("Two Sum")).toBeInTheDocument();
    expect(screen.getByText("简单")).toBeInTheDocument();
    expect(screen.getByText("60.0%")).toBeInTheDocument();
  });
});
