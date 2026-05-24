import { z } from "zod";

import { languages, permissionKeys, userRoles } from "./enums";

export const userRoleSchema = z.enum(userRoles);
export const permissionKeySchema = z.enum(permissionKeys);

export const authUserSchema = z.object({
  id: z.string().min(1),
  email: z.string().email(),
  username: z.string().min(1),
  role: userRoleSchema,
  permissions: z.array(permissionKeySchema),
  displayName: z.string().min(1),
});

export const sessionSummarySchema = z.object({
  id: z.string().min(1),
  current: z.boolean(),
  ip: z.string(),
  userAgent: z.string(),
  expiresAt: z.string().datetime(),
  lastSeenAt: z.string().datetime().nullable(),
  createdAt: z.string().datetime(),
});

export const loginInputSchema = z.object({
  identifier: z.string().min(1),
  password: z.string().min(8),
});

export const registerInputSchema = z.object({
  email: z.string().email(),
  username: z.string().min(3).max(32),
  password: z.string().min(8),
  confirmPassword: z.string().min(8),
  displayName: z.string().min(1).max(64),
});

export const changePasswordInputSchema = z.object({
  currentPassword: z.string().min(8),
  newPassword: z.string().min(8),
  confirmPassword: z.string().min(8),
});

export const profileSettingsSchema = z.object({
  displayName: z.string().min(1).max(64),
  bio: z.string().max(280),
  avatarUrl: z.string().max(2048),
});

export const preferenceSettingsSchema = z.object({
  preferredLocale: z.string().min(2).max(16),
  preferredLanguage: z.enum(languages),
});

export const meSummarySchema = z.object({
  user: authUserSchema,
  stats: z.object({
    solvedCount: z.number().int().nonnegative(),
    submissionCount: z.number().int().nonnegative(),
    acceptedCount: z.number().int().nonnegative(),
    lastActiveAt: z.string().datetime().nullable(),
  }),
  recentSubmissions: z.array(
    z.object({
      id: z.string().min(1),
      status: z.string().min(1),
      createdAt: z.string().datetime(),
      problem: z.object({
        id: z.string().min(1),
        slug: z.string().min(1),
        title: z.string().min(1),
      }),
    }),
  ),
  contests: z.array(
    z.object({
      id: z.string().min(1),
      slug: z.string().min(1),
      title: z.string().min(1),
      status: z.string().min(1),
      startsAt: z.string().datetime(),
      endsAt: z.string().datetime(),
    }),
  ),
});

export const meSettingsSchema = z.object({
  user: authUserSchema,
  profile: profileSettingsSchema.extend({
    email: z.string().email(),
    username: z.string().min(1),
  }),
  preferences: preferenceSettingsSchema,
  sessions: z.array(sessionSummarySchema),
});

export type AuthUser = z.infer<typeof authUserSchema>;
export type SessionSummary = z.infer<typeof sessionSummarySchema>;
export type LoginInput = z.infer<typeof loginInputSchema>;
export type RegisterInput = z.infer<typeof registerInputSchema>;
export type ChangePasswordInput = z.infer<typeof changePasswordInputSchema>;
export type ProfileSettings = z.infer<typeof profileSettingsSchema>;
export type PreferenceSettings = z.infer<typeof preferenceSettingsSchema>;
export type MeSummary = z.infer<typeof meSummarySchema>;
export type MeSettings = z.infer<typeof meSettingsSchema>;

