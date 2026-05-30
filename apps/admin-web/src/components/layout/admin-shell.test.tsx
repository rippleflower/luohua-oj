import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { AdminShell } from "./admin-shell";

vi.mock("../../features/auth/api", () => ({
  getAuthMe: vi.fn(async () => ({
    id: "admin-1",
    email: "admin@example.com",
    username: "admin",
    role: "SUPER_ADMIN",
    permissions: [],
    displayName: "Admin",
  })),
}));

describe("AdminShell", () => {
  it("shows user-site shortcut in the top bar", async () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <AdminShell title="总览">
          <div>content</div>
        </AdminShell>
      </QueryClientProvider>,
    );

    expect(await screen.findByText("前台")).toBeInTheDocument();
  });
});
