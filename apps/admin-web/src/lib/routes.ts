import { adminRoutes, normalizePathname } from "@oj/shared";

export type AdminRouteMatch =
  | { kind: "login" }
  | { kind: "dashboard" }
  | { kind: "users" }
  | { kind: "user-detail"; userId: string }
  | { kind: "user-permissions"; userId: string }
  | { kind: "problems" }
  | { kind: "contests" }
  | { kind: "submissions" }
  | { kind: "announcements" }
  | { kind: "system"; section: "settings" | "storage" | "judge" }
  | { kind: "audit" };

export function matchAdminRoute(pathname: string): AdminRouteMatch {
  const current = normalizePathname(pathname);

  if (current === adminRoutes.login) {
    return { kind: "login" };
  }
  if (current.startsWith("/users/") && current.endsWith("/permissions")) {
    return {
      kind: "user-permissions",
      userId: decodeURIComponent(current.slice("/users/".length, -"/permissions".length)),
    };
  }
  if (current.startsWith("/users/")) {
    return { kind: "user-detail", userId: decodeURIComponent(current.slice("/users/".length)) };
  }
  if (current === adminRoutes.users) {
    return { kind: "users" };
  }
  if (current === adminRoutes.problems) {
    return { kind: "problems" };
  }
  if (current === adminRoutes.contests) {
    return { kind: "contests" };
  }
  if (current === adminRoutes.submissions || current === adminRoutes.judgeQueue) {
    return { kind: "submissions" };
  }
  if (current === adminRoutes.announcements || current === "/announcements/new") {
    return { kind: "announcements" };
  }
  if (current === adminRoutes.systemStorage) {
    return { kind: "system", section: "storage" };
  }
  if (current === adminRoutes.systemJudge) {
    return { kind: "system", section: "judge" };
  }
  if (current === adminRoutes.systemSettings) {
    return { kind: "system", section: "settings" };
  }
  if (current === adminRoutes.audit) {
    return { kind: "audit" };
  }
  return { kind: "dashboard" };
}
