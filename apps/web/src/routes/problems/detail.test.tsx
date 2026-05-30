import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { LocaleProvider } from "../../lib/locale";
import { ProblemDetailRoute } from "./detail";

vi.mock("@monaco-editor/react", () => ({
  default: ({ value }: { value: string }) => (
    <textarea data-testid="monaco-editor" readOnly value={value} />
  ),
}));

describe("ProblemDetailRoute", () => {
  beforeEach(() => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL) => {
        const url =
          typeof input === "string"
            ? input
            : input instanceof URL
              ? input.toString()
              : input.url;

        if (url.endsWith("/auth/me")) {
          return new Response(
            JSON.stringify({
              id: "admin-1",
              email: "admin@example.com",
              username: "admin",
              role: "SUPER_ADMIN",
              permissions: [],
              displayName: "Admin",
            }),
            {
              status: 200,
              headers: { "Content-Type": "application/json" },
            },
          );
        }

        return new Response(JSON.stringify({ error: "not found" }), {
          status: 404,
          headers: { "Content-Type": "application/json" },
        });
      }),
    );
  });

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
    expect(screen.getByText("代码编辑与提交")).toBeInTheDocument();
    expect(screen.getByTestId("monaco-editor")).toHaveValue(
      "#include <bits/stdc++.h>\nusing namespace std;\n\nint main() {\n  ios::sync_with_stdio(false);\n  cin.tie(nullptr);\n\n  return 0;\n}\n",
    );
    expect(
      screen.getByText("problems/two-sum/versions/1/public/sample-1.in"),
    ).toBeInTheDocument();
  });
});
