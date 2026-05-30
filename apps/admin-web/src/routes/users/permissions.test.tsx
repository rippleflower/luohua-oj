import type { ReactElement } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { permissionKeys } from "@oj/shared";
import { UserPermissionsRoute } from "./permissions";

describe("UserPermissionsRoute", () => {
  beforeEach(() => {
    window.history.pushState({}, "", "/users/user-1/permissions");
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("syncs selected permissions from async query result", async () => {
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
          return jsonResponse(200, {
            id: "admin-1",
            email: "admin@example.com",
            username: "admin",
            role: "SUPER_ADMIN",
            permissions: [],
            displayName: "Admin",
          });
        }

        if (url.endsWith("/admin/users/user-1/permissions")) {
          return jsonResponse(200, [permissionKeys[0]]);
        }

        return jsonResponse(404, { error: "not found" });
      }),
    );

    renderWithQueryClient(<UserPermissionsRoute userId="user-1" />);

    const checkbox = await screen.findByRole("checkbox", {
      name: permissionKeys[0],
    });
    await waitFor(() => {
      expect(checkbox).toBeChecked();
    });
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
