import type {
  AuthUser,
  ChangePasswordInput,
  MeSettings,
  MeSummary,
  PreferenceSettings,
  ProfileSettings,
  RegisterInput,
  SessionSummary,
} from "@oj/shared";
import {
  authUserSchema,
  meSettingsSchema,
  meSummarySchema,
  sessionSummarySchema,
} from "@oj/shared";

import { getJSON, patchJSON, postJSON } from "../../lib/http-client";

export async function getAuthMe(): Promise<AuthUser> {
  const response = await getJSON<unknown>("/auth/me");
  return authUserSchema.parse(response);
}

export async function login(identifier: string, password: string): Promise<AuthUser> {
  const response = await postJSON<{ user: unknown }>("/auth/login", {
    body: { identifier, password },
  });
  return authUserSchema.parse(response.user);
}

export async function register(input: RegisterInput): Promise<AuthUser> {
  const response = await postJSON<{ user: unknown }>("/auth/register", {
    body: input,
  });
  return authUserSchema.parse(response.user);
}

export async function logout(): Promise<void> {
  await postJSON("/auth/logout", { body: {} });
}

export async function getMeSummary(): Promise<MeSummary> {
  const response = await getJSON<unknown>("/me/summary");
  return meSummarySchema.parse(response);
}

export async function getMeSettings(): Promise<MeSettings> {
  const response = await getJSON<unknown>("/me/settings");
  return meSettingsSchema.parse(response);
}

export async function updateProfile(input: ProfileSettings): Promise<MeSettings> {
  const response = await patchJSON<unknown>("/me/profile", { body: input });
  return meSettingsSchema.parse(response);
}

export async function updatePreferences(input: PreferenceSettings): Promise<MeSettings> {
  const response = await patchJSON<unknown>("/me/preferences", { body: input });
  return meSettingsSchema.parse(response);
}

export async function changePassword(input: ChangePasswordInput): Promise<void> {
  await postJSON("/auth/password/change", { body: input });
}

export async function listSessions(): Promise<SessionSummary[]> {
  const response = await getJSON<unknown[]>("/auth/sessions");
  return response.map((item) => sessionSummarySchema.parse(item));
}

export async function revokeSession(sessionId: string): Promise<void> {
  await postJSON("/auth/sessions/revoke", { body: { sessionId } });
}
