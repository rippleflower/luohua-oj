import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { LocaleProvider } from "../../lib/locale";
import { AppShell } from "./app-shell";

vi.mock("../../features/auth/hooks", () => ({
  useAuthUser: vi.fn(),
}));

import { useAuthUser } from "../../features/auth/hooks";

describe("AppShell", () => {
  it("shows admin console button for admin viewers only", () => {
    vi.mocked(useAuthUser).mockReturnValue({
      data: {
        id: "admin-1",
        email: "admin@example.com",
        username: "admin",
        role: "ADMIN",
        permissions: [],
        displayName: "Admin",
      },
    } as unknown as ReturnType<typeof useAuthUser>);

    render(
      <LocaleProvider>
        <QueryClientProvider client={new QueryClient()}>
          <AppShell subtitle="luooj" title="题库">
            <div>content</div>
          </AppShell>
        </QueryClientProvider>
      </LocaleProvider>,
    );

    expect(screen.getByText("管理端")).toBeInTheDocument();
  });

  it("hides admin console button for regular viewers", () => {
    vi.mocked(useAuthUser).mockReturnValue({
      data: {
        id: "user-1",
        email: "user@example.com",
        username: "user",
        role: "USER",
        permissions: [],
        displayName: "User",
      },
    } as unknown as ReturnType<typeof useAuthUser>);

    render(
      <LocaleProvider>
        <QueryClientProvider client={new QueryClient()}>
          <AppShell subtitle="luooj" title="题库">
            <div>content</div>
          </AppShell>
        </QueryClientProvider>
      </LocaleProvider>,
    );

    expect(screen.queryByText("管理端")).not.toBeInTheDocument();
  });

  it("renders the page action only once", () => {
    vi.mocked(useAuthUser).mockReturnValue({
      data: {
        id: "admin-1",
        email: "admin@example.com",
        username: "admin",
        role: "ADMIN",
        permissions: [],
        displayName: "Admin",
      },
    } as unknown as ReturnType<typeof useAuthUser>);

    render(
      <LocaleProvider>
        <QueryClientProvider client={new QueryClient()}>
          <AppShell
            action={<a href="/problems">返回题库</a>}
            subtitle="luooj"
            title="题目详情"
          >
            <div>content</div>
          </AppShell>
        </QueryClientProvider>
      </LocaleProvider>,
    );

    expect(screen.getAllByText("返回题库")).toHaveLength(1);
  });
});
