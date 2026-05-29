import { lazy, Suspense } from "react";
import { normalizePathname } from "@oj/shared";

import { matchAdminRoute } from "./lib/routes";

const AdminLoginRoute = lazy(async () => ({
  default: (await import("./routes/login")).AdminLoginRoute,
}));
const DashboardRoute = lazy(async () => ({
  default: (await import("./routes/dashboard")).DashboardRoute,
}));
const UsersRoute = lazy(async () => ({
  default: (await import("./routes/users")).UsersRoute,
}));
const UserDetailRoute = lazy(async () => ({
  default: (await import("./routes/users/detail")).UserDetailRoute,
}));
const UserPermissionsRoute = lazy(async () => ({
  default: (await import("./routes/users/permissions")).UserPermissionsRoute,
}));
const ProblemsRoute = lazy(async () => ({
  default: (await import("./routes/problems")).ProblemsRoute,
}));
const ContestsRoute = lazy(async () => ({
  default: (await import("./routes/contests")).ContestsRoute,
}));
const SubmissionsRoute = lazy(async () => ({
  default: (await import("./routes/submissions")).SubmissionsRoute,
}));
const AnnouncementsRoute = lazy(async () => ({
  default: (await import("./routes/announcements")).AnnouncementsRoute,
}));
const SystemSettingsRoute = lazy(async () => ({
  default: (await import("./routes/system/settings")).SystemSettingsRoute,
}));
const AuditRoute = lazy(async () => ({
  default: (await import("./routes/audit")).AuditRoute,
}));

function RouteFallback() {
  return (
    <main className="flex min-h-screen items-center justify-center px-6 text-sm text-slate-600">
      正在加载后台页面...
    </main>
  );
}

export function App() {
  const pathname = normalizePathname(window.location.pathname);

  return <Suspense fallback={<RouteFallback />}>{resolveRoute(pathname)}</Suspense>;
}

function resolveRoute(pathname: string) {
  const match = matchAdminRoute(pathname);

  switch (match.kind) {
    case "login":
      return <AdminLoginRoute />;
    case "user-permissions":
      return <UserPermissionsRoute userId={match.userId} />;
    case "user-detail":
      return <UserDetailRoute userId={match.userId} />;
    case "users":
      return <UsersRoute />;
    case "problems":
      return <ProblemsRoute />;
    case "contests":
      return <ContestsRoute />;
    case "submissions":
      return <SubmissionsRoute />;
    case "announcements":
      return <AnnouncementsRoute />;
    case "system":
      return <SystemSettingsRoute section={match.section} />;
    case "audit":
      return <AuditRoute />;
    case "dashboard":
    default:
      return <DashboardRoute />;
  }
}
