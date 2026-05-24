import type { AdminDashboard, AuthUser } from "@oj/shared";
import { adminDashboardSchema, authUserSchema } from "@oj/shared";

import { getJSON, postJSON } from "../../lib/http-client";

export async function login(identifier: string, password: string): Promise<AuthUser> {
  const response = await postJSON<{ user: unknown }>("/auth/login", {
    body: { identifier, password },
  });
  return authUserSchema.parse(response.user);
}

export async function getAuthMe(): Promise<AuthUser> {
  const response = await getJSON<unknown>("/auth/me");
  return authUserSchema.parse(response);
}

export async function getAdminDashboard(): Promise<AdminDashboard> {
  const response = await getJSON<unknown>("/admin/me");
  return adminDashboardSchema.parse(response);
}
