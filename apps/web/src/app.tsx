import { lazy, Suspense } from "react";
import { normalizePathname } from "@oj/shared";

import { matchWebRoute } from "./lib/routes";

const LoginRoute = lazy(async () => ({
  default: (await import("./routes/auth/login")).LoginRoute,
}));
const RegisterRoute = lazy(async () => ({
  default: (await import("./routes/auth/register")).RegisterRoute,
}));
const ContestMakeupRoute = lazy(async () => ({
  default: (await import("./routes/contests/makeup")).ContestMakeupRoute,
}));
const ContestDetailRoute = lazy(async () => ({
  default: (await import("./routes/contests/detail")).ContestDetailRoute,
}));
const ContestsRoute = lazy(async () => ({
  default: (await import("./routes/contests")).ContestsRoute,
}));
const HomeRoute = lazy(async () => ({
  default: (await import("./routes/home")).HomeRoute,
}));
const MeRoute = lazy(async () => ({
  default: (await import("./routes/me")).MeRoute,
}));
const ProblemDetailRoute = lazy(async () => ({
  default: (await import("./routes/problems/detail")).ProblemDetailRoute,
}));
const ProblemsRoute = lazy(async () => ({
  default: (await import("./routes/problems")).ProblemsRoute,
}));
const AccountSettingsRoute = lazy(async () => ({
  default: (await import("./routes/settings/account")).AccountSettingsRoute,
}));
const ProfileSettingsRoute = lazy(async () => ({
  default: (await import("./routes/settings/profile")).ProfileSettingsRoute,
}));
const SecuritySettingsRoute = lazy(async () => ({
  default: (await import("./routes/settings/security")).SecuritySettingsRoute,
}));
const SubmissionDetailRoute = lazy(async () => ({
  default: (await import("./routes/submissions/detail")).SubmissionDetailRoute,
}));
const SubmissionsRoute = lazy(async () => ({
  default: (await import("./routes/submissions")).SubmissionsRoute,
}));

function RouteFallback() {
  return (
    <main className="min-h-screen bg-transparent px-6 py-12 text-slate-600">
      <div className="mx-auto max-w-7xl rounded-3xl border border-slate-200/80 bg-white/80 px-5 py-4 text-sm shadow-sm">
        Loading...
      </div>
    </main>
  );
}

export function App() {
  const pathname = normalizePathname(window.location.pathname);

  return <Suspense fallback={<RouteFallback />}>{resolveRoute(pathname)}</Suspense>;
}

function resolveRoute(pathname: string) {
  const match = matchWebRoute(pathname);

  switch (match.kind) {
    case "contest-makeup":
      return <ContestMakeupRoute slug={match.slug} />;
    case "contest-detail":
      return <ContestDetailRoute slug={match.slug} />;
    case "submission-detail":
      return <SubmissionDetailRoute submissionId={match.submissionId} />;
    case "problem-detail":
      return <ProblemDetailRoute routeCode={match.routeCode} slug={match.slug} />;
    case "login":
      return <LoginRoute />;
    case "register":
      return <RegisterRoute />;
    case "me":
      return <MeRoute />;
    case "settings-profile":
      return <ProfileSettingsRoute />;
    case "settings-account":
      return <AccountSettingsRoute />;
    case "settings-security":
      return <SecuritySettingsRoute />;
    case "contests":
      return <ContestsRoute />;
    case "submissions":
      return <SubmissionsRoute />;
    case "problems":
      return <ProblemsRoute />;
    case "home":
    default:
      return <HomeRoute />;
  }
}
