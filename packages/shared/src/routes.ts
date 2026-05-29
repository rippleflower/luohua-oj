type AppSection = "home" | "problems" | "contests" | "submissions" | "me";

export const webRoutes = {
  home: "/",
  problems: "/problems",
  contests: "/contests",
  submissions: "/submissions",
  me: "/me",
  login: "/login",
  register: "/register",
  settingsProfile: "/settings/profile",
  settingsAccount: "/settings/account",
  settingsSecurity: "/settings/security",
  problemDetail(routeCode: string): string {
    return `/problems/p/${encodeURIComponent(routeCode)}`;
  },
  legacyProblemDetail(slug: string): string {
    return `/problems/${encodeURIComponent(slug)}`;
  },
  contestDetail(slug: string): string {
    return `/contests/${encodeURIComponent(slug)}`;
  },
  contestMakeup(slug: string): string {
    return `/contests/${encodeURIComponent(slug)}/makeup-list`;
  },
  submissionDetail(submissionId: string): string {
    return `/submissions/${encodeURIComponent(submissionId)}`;
  },
} as const;

export const adminRoutes = {
  dashboard: "/",
  login: "/login",
  users: "/users",
  problems: "/problems",
  contests: "/contests",
  submissions: "/submissions",
  judgeQueue: "/judge/queue",
  announcements: "/announcements",
  systemSettings: "/system/settings",
  systemStorage: "/system/storage",
  systemJudge: "/system/judge",
  audit: "/audit",
  userDetail(userId: string): string {
    return `/users/${encodeURIComponent(userId)}`;
  },
  userPermissions(userId: string): string {
    return `/users/${encodeURIComponent(userId)}/permissions`;
  },
} as const;

export function normalizePathname(pathname: string): string {
  return pathname.replace(/\/+$/, "") || "/";
}

export function isRouteActive(pathname: string, href: string): boolean {
  const current = normalizePathname(pathname);
  const target = normalizePathname(href);

  if (target === "/") {
    return current === "/";
  }

  return current === target || current.startsWith(`${target}/`);
}

export function webSectionForPath(pathname: string): AppSection {
  const current = normalizePathname(pathname);
  if (isRouteActive(current, webRoutes.problems)) {
    return "problems";
  }
  if (isRouteActive(current, webRoutes.contests)) {
    return "contests";
  }
  if (isRouteActive(current, webRoutes.submissions)) {
    return "submissions";
  }
  if (isRouteActive(current, webRoutes.me) || current.startsWith("/settings/")) {
    return "me";
  }
  return "home";
}
