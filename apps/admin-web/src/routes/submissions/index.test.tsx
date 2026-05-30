import type { ReactElement } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { SubmissionsRoute } from "./index";

describe("SubmissionsRoute", () => {
  beforeEach(() => {
    window.history.pushState({}, "", "/submissions");
    document.cookie = "oj_csrf=csrf-token";
  });

  afterEach(() => {
    vi.restoreAllMocks();
    document.cookie = "oj_csrf=; expires=Thu, 01 Jan 1970 00:00:00 GMT";
  });

  it("loads queue summary and rejudges a submission", async () => {
    const viewer = {
      id: "user-1",
      email: "admin@example.com",
      username: "admin",
      role: "SUPER_ADMIN",
      permissions: [],
      displayName: "Admin",
    };
    const submissions = [
      {
        id: "submission-1",
        username: "demo",
        problemTitle: "Two Sum",
        language: "CPP17",
        status: "WRONG_ANSWER",
        createdAt: "2026-05-18T12:00:00Z",
      },
    ];
    const queueSummary = {
      queue: "judge",
      paused: false,
      latencySeconds: 3,
      pending: 2,
      active: 1,
      scheduled: 0,
      retry: 1,
      archived: 0,
      completed: 4,
      processedToday: 5,
      failedToday: 1,
      recentTasks: [
        {
          id: "task-1",
          type: "judge:submission",
          submissionId: "submission-1",
          state: "retry",
          createdAt: null,
          nextProcessAt: "2026-05-18T12:01:00Z",
          completedAt: null,
          lastErr: "wa",
        },
      ],
      updatedAt: "2026-05-18T12:02:00Z",
    };

    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
        const url =
          typeof input === "string"
            ? input
            : input instanceof URL
              ? input.toString()
              : input.url;
        const method = init?.method ?? "GET";

        if (url.endsWith("/auth/me")) {
          return jsonResponse(200, viewer);
        }
        if (url.endsWith("/admin/submissions") && method === "GET") {
          return jsonResponse(200, submissions);
        }
        if (url.endsWith("/admin/judge/queue") && method === "GET") {
          return jsonResponse(200, queueSummary);
        }
        if (
          url.endsWith("/admin/submissions/submission-1/rejudge") &&
          method === "POST"
        ) {
          return jsonResponse(200, {
            submissionId: "submission-1",
            queue: "judge",
            resultSnapshotVersion: 0,
            problemVersionId: null,
          });
        }
        return jsonResponse(404, { error: "not found" });
      }),
    );

    renderWithQueryClient(<SubmissionsRoute />);

    expect(await screen.findByText("Two Sum")).toBeInTheDocument();
    expect(await screen.findByText("judge")).toBeInTheDocument();
    expect(await screen.findByText("wa")).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("submission-1 rejudge reason"), {
      target: { value: "rerun" },
    });
    fireEvent.click(screen.getByRole("button", { name: "重判" }));

    expect(await screen.findByText("重判任务已入队。")).toBeInTheDocument();
  });
});

function renderWithQueryClient(element: ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>{element}</QueryClientProvider>,
  );
}

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
