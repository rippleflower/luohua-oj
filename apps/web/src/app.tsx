import { lazy, Suspense } from "react";

const LoginRoute = lazy(async () => ({
  default: (await import("./routes/auth/login")).LoginRoute,
}));
const RegisterRoute = lazy(async () => ({
  default: (await import("./routes/auth/register")).RegisterRoute,
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
  const pathname = window.location.pathname.replace(/\/$/, "") || "/";

  return <Suspense fallback={<RouteFallback />}>{resolveRoute(pathname)}</Suspense>;
}

function resolveRoute(pathname: string) {
  if (pathname.startsWith("/contests/")) {
    const slug = pathname.slice("/contests/".length);
    return <ContestDetailRoute slug={slug} />;
  }

  if (pathname.startsWith("/submissions/")) {
    const submissionId = pathname.slice("/submissions/".length);
    return <SubmissionDetailRoute submissionId={submissionId} />;
  }

  if (pathname.startsWith("/problems/")) {
    const slug = pathname.slice("/problems/".length);
    return <ProblemDetailRoute slug={slug} />;
  }

  if (pathname === "/login") {
    return <LoginRoute />;
  }

  if (pathname === "/register") {
    return <RegisterRoute />;
  }

  if (pathname === "/me") {
    return <MeRoute />;
  }

  if (pathname === "/settings/profile") {
    return <ProfileSettingsRoute />;
  }

  if (pathname === "/settings/account") {
    return <AccountSettingsRoute />;
  }

  if (pathname === "/settings/security") {
    return <SecuritySettingsRoute />;
  }

  if (pathname === "/contests") {
    return <ContestsRoute />;
  }

  if (pathname === "/submissions") {
    return <SubmissionsRoute />;
  }

  if (pathname === "/problems") {
    return <ProblemsRoute />;
  }

  return <HomeRoute />;
}
