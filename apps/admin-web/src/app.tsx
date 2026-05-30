import { AdminLoginRoute } from "./routes/login";
import { DashboardRoute } from "./routes/dashboard";
import { UsersRoute } from "./routes/users";
import { UserDetailRoute } from "./routes/users/detail";
import { UserPermissionsRoute } from "./routes/users/permissions";
import { ProblemsRoute } from "./routes/problems";
import { ContestsRoute } from "./routes/contests";
import { SubmissionsRoute } from "./routes/submissions";
import { AnnouncementsRoute } from "./routes/announcements";
import { SystemSettingsRoute } from "./routes/system/settings";
import { AuditRoute } from "./routes/audit";

export function App() {
  const pathname = window.location.pathname.replace(/\/$/, "") || "/";

  if (pathname === "/login") {
    return <AdminLoginRoute />;
  }

  if (pathname.startsWith("/users/") && pathname.endsWith("/permissions")) {
    return <UserPermissionsRoute userId={pathname.slice("/users/".length, -"/permissions".length)} />;
  }

  if (pathname.startsWith("/users/")) {
    return <UserDetailRoute userId={pathname.slice("/users/".length)} />;
  }

  if (pathname === "/users") {
    return <UsersRoute />;
  }

  if (pathname === "/problems") {
    return <ProblemsRoute />;
  }

  if (pathname === "/contests") {
    return <ContestsRoute />;
  }

  if (pathname === "/submissions" || pathname === "/judge/queue") {
    return <SubmissionsRoute />;
  }

  if (pathname === "/announcements" || pathname === "/announcements/new") {
    return <AnnouncementsRoute />;
  }

  if (pathname === "/system/storage") {
    return <SystemSettingsRoute section="storage" />;
  }

  if (pathname === "/system/judge") {
    return <SystemSettingsRoute section="judge" />;
  }

  if (pathname === "/system/settings") {
    return <SystemSettingsRoute section="settings" />;
  }

  if (pathname === "/audit") {
    return <AuditRoute />;
  }

  return <DashboardRoute />;
}
