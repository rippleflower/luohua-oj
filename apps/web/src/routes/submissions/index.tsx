import { useEffect, useState, type FormEvent } from "react";
import { webRoutes } from "@oj/shared";

import { AppShell } from "../../components/layout/app-shell";
import { SubmissionForm } from "../../components/submission/submission-form";
import { SubmissionHistoryList } from "../../components/submission/submission-history-list";
import { SubmissionResponsePanel } from "../../components/submission/submission-response-panel";
import { useCreateSubmission, useSubmissionHistory } from "../../features/submissions/hooks";
import {
  createSubmissionRequestId,
  logSubmissionFieldChanged,
  logSubmissionPageViewed,
  logSubmissionRequested,
  logSubmissionValidationFailed,
} from "../../features/submissions/logging";
import { createSubmissionRequestSchema, type CreateSubmissionRequest, type SubmissionList, type SubmissionLogContext } from "../../features/submissions/schema";
import { env } from "../../lib/env";
import { useLocale } from "../../lib/locale";

const defaultForm: CreateSubmissionRequest = {
  userId: "",
  problemId: "",
  language: "CPP17",
  source: "#include <bits/stdc++.h>\nusing namespace std;\n\nint main() {\n  cout << \"Hello, OJ\" << '\\n';\n  return 0;\n}\n",
};

export function SubmissionsRoute() {
  const { locale } = useLocale();
  const historyPageSize = 20;
  const [form, setForm] = useState<CreateSubmissionRequest>(defaultForm);
  const [requestId, setRequestId] = useState(() => createSubmissionRequestId());
  const [validationError, setValidationError] = useState<string>();
  const [historyUsername, setHistoryUsername] = useState(env.submissionsUsername);
  const [historyPage, setHistoryPage] = useState(1);
  const mutation = useCreateSubmission();
  const apiMode = env.demoMode && env.apiBaseUrl === "" ? "fallback" : "remote";
  const historyQuery = useSubmissionHistory(historyUsername, historyPage, historyPageSize);
  const history = historyQuery.data;

  useEffect(() => {
    logSubmissionPageViewed(requestId, apiMode, form);
  }, [requestId]);

  function handleHistoryUsernameChange(value: string) {
    setHistoryUsername(value);
    setHistoryPage(1);
  }

  function handleFieldChange(field: NonNullable<SubmissionLogContext["field"]>, value: string) {
    const nextForm =
      field === "language"
        ? { ...form, language: value as CreateSubmissionRequest["language"] }
        : { ...form, [field]: value };

    setForm(nextForm);
    logSubmissionFieldChanged(requestId, apiMode, nextForm, field);
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setValidationError(undefined);

    const validation = createSubmissionRequestSchema.safeParse(form);
    if (!validation.success) {
      const message = validation.error.issues[0]?.message ?? "invalid submission request";
      setValidationError(message);
      logSubmissionValidationFailed(requestId, apiMode, form, new Error(message));
      return;
    }

    logSubmissionRequested(requestId, apiMode, form);
    mutation.mutate(
      { form, requestId },
      {
        onSettled: () => {
          setRequestId(createSubmissionRequestId());
        },
      },
    );
  }

  return (
    <AppShell
      title={locale === "zh" ? "提交" : "Submissions"}
      subtitle="luooj"
      action={
        <a
          className="rounded-md border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700 hover:bg-slate-50"
          href={webRoutes.problems}
        >
          {locale === "zh" ? "返回题库" : "Back to Problems"}
        </a>
      }
    >
      <div className="grid gap-6 lg:grid-cols-[minmax(0,2fr)_minmax(320px,1fr)]">
        <div className="space-y-6">
          <SubmissionForm form={form} isSubmitting={mutation.isPending} onFieldChange={handleFieldChange} onSubmit={handleSubmit} />
          <SubmissionHistoryList
            onPageChange={apiMode === "remote" && historyUsername.trim() !== "" ? setHistoryPage : undefined}
            onUsernameChange={handleHistoryUsernameChange}
            page={history?.page ?? 1}
            pageSize={history?.pageSize ?? historyPageSize}
            sourceLabel={
              apiMode === "remote" && historyUsername.trim() !== ""
                ? locale === "zh"
                  ? `显示用户 ${historyUsername.trim()} 的远端提交`
                  : `Showing remote submissions for ${historyUsername.trim()}`
                : locale === "zh"
                  ? "显示 demo 本地提交记录"
                  : "Showing demo-mode local submission history"
            }
            submissions={history?.items ?? []}
            total={history?.total ?? 0}
            username={historyUsername}
          />
          {historyQuery.isError ? (
            <div className="rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
              {historyQuery.error instanceof Error
                ? historyQuery.error.message
                : locale === "zh"
                  ? "提交记录加载失败。"
                  : "Failed to load submissions."}
            </div>
          ) : null}
        </div>

        <aside className="space-y-4">
          <SubmissionResponsePanel
            data={mutation.data}
            errorMessage={validationError ?? (mutation.isError ? mutation.error.message : undefined)}
          />

          <section className="rounded-lg border border-slate-200 bg-white p-6">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500">{locale === "zh" ? "说明" : "Notes"}</h2>
            <ul className="mt-4 space-y-2 text-sm text-slate-600">
              <li>{locale === "zh" ? "只有在 `VITE_DEMO_MODE=true` 时，这一页才会回退到本地 demo 数据。" : "This page falls back to local demo data only when `VITE_DEMO_MODE=true`."}</li>
              <li>{locale === "zh" ? "当前 API 仍要求 `userId` 和 `problemId` 使用后端可解析的值。" : "The current API expects `userId` and `problemId` values that the backend can parse."}</li>
            </ul>
          </section>
        </aside>
      </div>
    </AppShell>
  );
}
