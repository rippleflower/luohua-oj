import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { LocaleProvider } from "../../lib/locale";
import { ContestsRoute } from "./index";

describe("ContestsRoute", () => {
  it("renders contest groups", async () => {
    const queryClient = new QueryClient();

    render(
      <LocaleProvider>
        <QueryClientProvider client={queryClient}>
          <ContestsRoute />
        </QueryClientProvider>
      </LocaleProvider>,
    );

    expect(await screen.findAllByText("进行中")).toHaveLength(2);
    expect(screen.getAllByText("即将开始")).toHaveLength(2);
    expect(screen.getAllByText("已结束")).toHaveLength(2);
    expect(await screen.findByText("2026 春季公开赛")).toBeInTheDocument();
  });
});
