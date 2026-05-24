import { LoginRoute } from "./routes/auth/login";
import { RegisterRoute } from "./routes/auth/register";
import { ContestDetailRoute } from "./routes/contests/detail";
import { ContestsRoute } from "./routes/contests";
import { HomeRoute } from "./routes/home";
import { MeRoute } from "./routes/me";
import { ProblemDetailRoute } from "./routes/problems/detail";
import { ProblemsRoute } from "./routes/problems";
import { AccountSettingsRoute } from "./routes/settings/account";
import { ProfileSettingsRoute } from "./routes/settings/profile";
import { SecuritySettingsRoute } from "./routes/settings/security";
import { SubmissionDetailRoute } from "./routes/submissions/detail";
import { SubmissionsRoute } from "./routes/submissions";

export function App() {
  const pathname = window.location.pathname.replace(/\/$/, "") || "/";

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
