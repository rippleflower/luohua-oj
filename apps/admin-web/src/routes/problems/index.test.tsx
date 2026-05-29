import type { ReactElement } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { ProblemsRoute } from "./index";

describe("ProblemsRoute", () => {
  beforeEach(() => {
    window.history.pushState({}, "", "/problems");
    document.cookie = "oj_csrf=csrf-token";
  });

  afterEach(() => {
    vi.restoreAllMocks();
    document.cookie = "oj_csrf=; expires=Thu, 01 Jan 1970 00:00:00 GMT";
  });

  it("loads, creates, updates, and publishes problems", async () => {
    let problems = [
      {
        id: "problem-1",
        problemNo: 1,
        routeCode: "CODE1",
        slug: "two-sum",
        title: "Two Sum",
        difficulty: "EASY",
        timeLimitMs: 1000,
        memoryLimitKb: 262144,
        status: "DRAFT",
        currentVersionNo: 1,
        isPublished: false,
        submissionCount: 0,
        acceptedRate: 0,
        updatedAt: "2026-05-18T12:00:00Z",
      },
    ];
    const viewer = {
      id: "user-1",
      email: "admin@example.com",
      username: "admin",
      role: "SUPER_ADMIN",
      permissions: [],
      displayName: "Admin",
    };

    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = typeof input === "string" ? input : input instanceof URL ? input.toString() : input.url;
        const method = init?.method ?? "GET";

        if (url.endsWith("/auth/me")) {
          return jsonResponse(200, viewer);
        }
        if (url.endsWith("/admin/problems") && method === "GET") {
          return jsonResponse(200, problems);
        }
        if (url.endsWith("/admin/problems") && method === "POST") {
          const body = JSON.parse(String(init?.body ?? "{}"));
          const created = {
            id: "problem-2",
            problemNo: 2,
            routeCode: "CODE2",
            slug: body.slug,
            title: body.title,
            difficulty: body.difficulty,
            timeLimitMs: body.timeLimitMs,
            memoryLimitKb: body.memoryLimitKb,
            status: "DRAFT",
            currentVersionNo: 1,
            isPublished: false,
            submissionCount: 0,
            acceptedRate: 0,
            updatedAt: "2026-05-18T12:05:00Z",
          };
          problems = [created, ...problems];
          return jsonResponse(201, created);
        }
        if (url.includes("/admin/problems/problem-1") && method === "PATCH") {
          const body = JSON.parse(String(init?.body ?? "{}"));
          problems = problems.map((item) =>
            item.id === "problem-1"
              ? {
                  ...item,
                  slug: body.slug,
                  title: body.title,
                  difficulty: body.difficulty,
                  timeLimitMs: body.timeLimitMs,
                  memoryLimitKb: body.memoryLimitKb,
                }
              : item,
          );
          return jsonResponse(200, problems.find((item) => item.id === "problem-1"));
        }
        if (url.endsWith("/admin/problems/problem-1/publish") && method === "POST") {
          problems = problems.map((item) =>
            item.id === "problem-1"
              ? {
                  ...item,
                  status: "PUBLISHED",
                  isPublished: true,
                }
              : item,
          );
          return jsonResponse(200, problems.find((item) => item.id === "problem-1"));
        }
        return jsonResponse(404, { error: "not found" });
      }),
    );

    renderWithQueryClient(<ProblemsRoute />);

    expect(await screen.findByRole("heading", { name: "Two Sum", level: 2 })).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("新建题目 Slug"), { target: { value: "sum-ii" } });
    fireEvent.change(screen.getByLabelText("新建题目 标题"), { target: { value: "Sum II" } });
    fireEvent.click(screen.getByRole("button", { name: "创建草稿" }));

    expect(await screen.findByText("草稿题目已创建。")).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "Sum II", level: 2 })).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("two-sum title"), { target: { value: "Two Sum Plus" } });
    fireEvent.change(screen.getByLabelText("two-sum difficulty"), { target: { value: "MEDIUM" } });
    fireEvent.click(screen.getAllByRole("button", { name: "保存修改" })[1]);

    expect(await screen.findByText("题目信息已更新，主站内容会在再次发布后刷新。")).toBeInTheDocument();
    expect(await screen.findByRole("heading", { name: "Two Sum Plus", level: 2 })).toBeInTheDocument();

    fireEvent.change(screen.getByLabelText("two-sum reason"), { target: { value: "ship it" } });
    fireEvent.click(screen.getAllByRole("button", { name: "发布到主站" })[1]);

    expect(await screen.findByText("题目已发布，公开读模型已刷新。")).toBeInTheDocument();
    await waitFor(() => expect(screen.getAllByText("PUBLISHED")[0]).toBeInTheDocument());
  });

  it("shows backend validation errors", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
        const url = typeof input === "string" ? input : input instanceof URL ? input.toString() : input.url;
        const method = init?.method ?? "GET";

        if (url.endsWith("/auth/me")) {
          return jsonResponse(200, {
            id: "user-1",
            email: "admin@example.com",
            username: "admin",
            role: "SUPER_ADMIN",
            permissions: [],
            displayName: "Admin",
          });
        }
        if (url.endsWith("/admin/problems") && method === "GET") {
          return jsonResponse(200, []);
        }
        if (url.endsWith("/admin/problems") && method === "POST") {
          return jsonResponse(400, { error: "slug already exists" });
        }
        return jsonResponse(404, { error: "not found" });
      }),
    );

    renderWithQueryClient(<ProblemsRoute />);

    await screen.findByText("新建草稿题目");
    fireEvent.change(screen.getByLabelText("新建题目 Slug"), { target: { value: "two-sum" } });
    fireEvent.change(screen.getByLabelText("新建题目 标题"), { target: { value: "Two Sum" } });
    fireEvent.click(screen.getByRole("button", { name: "创建草稿" }));

    expect(await screen.findByText("slug already exists")).toBeInTheDocument();
  });
});

function renderWithQueryClient(element: ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  });

  return render(<QueryClientProvider client={queryClient}>{element}</QueryClientProvider>);
}

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
