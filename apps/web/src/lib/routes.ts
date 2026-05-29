import { normalizePathname, webRoutes } from "@oj/shared";

export type WebRouteMatch =
  | { kind: "home" }
  | { kind: "login" }
  | { kind: "register" }
  | { kind: "me" }
  | { kind: "settings-profile" }
  | { kind: "settings-account" }
  | { kind: "settings-security" }
  | { kind: "problems" }
  | { kind: "problem-detail"; routeCode?: string; slug?: string }
  | { kind: "contests" }
  | { kind: "contest-makeup"; slug: string }
  | { kind: "contest-detail"; slug: string }
  | { kind: "submissions" }
  | { kind: "submission-detail"; submissionId: string };

export function matchWebRoute(pathname: string): WebRouteMatch {
  const current = normalizePathname(pathname);

  if (current.startsWith("/problems/p/")) {
    return { kind: "problem-detail", routeCode: decodeURIComponent(current.slice("/problems/p/".length)) };
  }
  if (current.startsWith("/problems/")) {
    return { kind: "problem-detail", slug: decodeURIComponent(current.slice("/problems/".length)) };
  }
  if (current === webRoutes.login) {
    return { kind: "login" };
  }
  if (current === webRoutes.register) {
    return { kind: "register" };
  }
  if (current === webRoutes.me) {
    return { kind: "me" };
  }
  if (current === webRoutes.settingsProfile) {
    return { kind: "settings-profile" };
  }
  if (current === webRoutes.settingsAccount) {
    return { kind: "settings-account" };
  }
  if (current === webRoutes.settingsSecurity) {
    return { kind: "settings-security" };
  }
  if (current.startsWith("/contests/") && current.endsWith("/makeup-list")) {
    return {
      kind: "contest-makeup",
      slug: decodeURIComponent(current.slice("/contests/".length, -"/makeup-list".length)),
    };
  }
  if (current.startsWith("/contests/")) {
    return { kind: "contest-detail", slug: decodeURIComponent(current.slice("/contests/".length)) };
  }
  if (current.startsWith("/submissions/")) {
    return { kind: "submission-detail", submissionId: decodeURIComponent(current.slice("/submissions/".length)) };
  }
  if (current === webRoutes.contests) {
    return { kind: "contests" };
  }
  if (current === webRoutes.submissions) {
    return { kind: "submissions" };
  }
  if (current === webRoutes.problems) {
    return { kind: "problems" };
  }
  return { kind: "home" };
}
