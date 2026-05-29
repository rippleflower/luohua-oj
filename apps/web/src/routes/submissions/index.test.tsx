import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { LocaleProvider } from "../../lib/locale";
import { resetLogTransport, setLogTransport } from "../../lib/logger";
import { SubmissionsRoute } from "./index";

vi.mock("../../lib/env", () => ({
  env: {
    apiBaseUrl: "",
    adminBaseUrl: "",
    submissionsUsername: "",
    demoMode: true,
  },
}));

describe("SubmissionsRoute", () => {
  const events: Record<string, unknown>[] = [];

  beforeEach(() => {
    events.length = 0;
    setLogTransport((entry) => {
      events.push(entry);
    });
  });

  afterEach(() => {
    resetLogTransport();
  });

  it("renders a created submission with fallback data and logs events", async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        mutations: {
          retry: false,
        },
      },
    });

    render(
      <LocaleProvider>
        <QueryClientProvider client={queryClient}>
          <SubmissionsRoute />
        </QueryClientProvider>
      </LocaleProvider>,
    );

    fireEvent.change(screen.getByLabelText("用户 ID"), {
      target: { value: "demo-user" },
    });
    fireEvent.change(screen.getByLabelText("题目 ID"), {
      target: { value: "two-sum" },
    });
    fireEvent.click(screen.getByRole("button", { name: "创建提交" }));

    expect(await screen.findAllByText("等待中")).toHaveLength(2);
    expect(screen.getByText(/tmp\/submissions\/two-sum-/)).toBeInTheDocument();

    await waitFor(() => {
      expect(events.some((entry) => entry.event === "submissions.fallback_used")).toBe(true);
      expect(events.some((entry) => entry.event === "submissions.request_succeeded")).toBe(true);
    });

    expect(events.every((entry) => !JSON.stringify(entry).includes("Hello, OJ"))).toBe(true);
  });

  it("logs validation failures without sending source content", async () => {
    const queryClient = new QueryClient({
      defaultOptions: {
        mutations: {
          retry: false,
        },
      },
    });

    render(
      <LocaleProvider>
        <QueryClientProvider client={queryClient}>
          <SubmissionsRoute />
        </QueryClientProvider>
      </LocaleProvider>,
    );

    fireEvent.change(screen.getByLabelText("源码"), {
      target: { value: "   " },
    });
    fireEvent.click(screen.getByRole("button", { name: "创建提交" }));

    expect(await screen.findByText(/String must contain at least 1 character/)).toBeInTheDocument();
    expect(events.some((entry) => entry.event === "submissions.validation_failed")).toBe(true);
    expect(events.every((entry) => !JSON.stringify(entry).includes("\"source\":"))).toBe(true);
  });
});
