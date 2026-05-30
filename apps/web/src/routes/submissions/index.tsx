import { useDeferredValue, useEffect, useRef, useState, type FormEvent } from "react";

import { AppShell } from "../../components/layout/app-shell";
import { SubmissionForm } from "../../components/submission/submission-form";
import { SubmissionHistoryList } from "../../components/submission/submission-history-list";
import { SubmissionResponsePanel } from "../../components/submission/submission-response-panel";
import { listSubmissionHistory } from "../../features/submissions/history";
import { useCreateSubmission } from "../../features/submissions/hooks";
import {
  createSubmissionRequestId,
  logSubmissionFieldChanged,
  logSubmissionPageViewed,
  logSubmissionRequested,
  logSubmissionValidationFailed,
} from "../../features/submissions/logging";
import { listSubmissions } from "../../features/submissions/read";
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
  const [history, setHistory] = useState<SubmissionList>(() => {
    const items = listSubmissionHistory();
    return {
      items,
      total: items.length,
      page: 1,
      pageSize: items.length === 0 ? historyPageSize : items.length,
    };
  });
  const [historyUsername, setHistoryUsername] = useState(env.submissionsUsername);
  const deferredHistoryUsername = useDeferredValue(historyUsername);
  const [historyPage, setHistoryPage] = useState(1);
  const mutation = useCreateSubmission();
  const apiMode = env.apiBaseUrl !== "" ? "remote" : "fallback";
  const fieldLogTimerRef = useRef<Partial<Record<NonNullable<SubmissionLogContext["field"]>, number>>>({});
  const historyUsernameQuery = deferredHistoryUsername.trim();

  useEffect(() => {
    logSubmissionPageViewed(requestId, apiMode, form);
  }, [requestId]);

  useEffect(
    () => () => {
      for (const timer of Object.values(fieldLogTimerRef.current)) {
        if (timer !== undefined) {
          window.clearTimeout(timer);
        }
      }
    },
    [],
  );

  useEffect(() => {
    const syncHistory = () => {
      if (apiMode === "fallback" || historyUsernameQuery === "") {
        const items = listSubmissionHistory();
        setHistory({
          items,
          total: items.length,
          page: 1,
          pageSize: items.length === 0 ? historyPageSize : items.length,
        });
      }
    };

    window.addEventListener("submissions:history-updated", syncHistory);
    return () => {
      window.removeEventListener("submissions:history-updated", syncHistory);
    };
  }, [apiMode, historyUsernameQuery]);

  useEffect(() => {
    let cancelled = false;

    void listSubmissions(historyUsernameQuery, { page: historyPage, pageSize: historyPageSize })
      .then((next) => {
        if (!cancelled) {
          setHistory(next);
        }
      })
      .catch(() => {
        if (!cancelled && (apiMode === "fallback" || historyUsernameQuery === "")) {
          const items = listSubmissionHistory();
          setHistory({
            items,
            total: items.length,
            page: 1,
            pageSize: items.length === 0 ? historyPageSize : items.length,
          });
        }
      });

    return () => {
      cancelled = true;
    };
  }, [apiMode, historyPage, historyUsernameQuery, mutation.data]);

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
    if (field === "source") {
      const existingTimer = fieldLogTimerRef.current[field];
      if (existingTimer !== undefined) {
        window.clearTimeout(existingTimer);
      }
      fieldLogTimerRef.current[field] = window.setTimeout(() => {
        logSubmissionFieldChanged(requestId, apiMode, nextForm, field);
      }, 250);
      return;
    }
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
          href="/problems"
        >
          {locale === "zh" ? "返回题库" : "Back to Problems"}
        </a>
      }
    >
      <div className="grid gap-6 lg:grid-cols-[minmax(0,2fr)_minmax(320px,1fr)]">
        <div className="space-y-6">
          <SubmissionForm form={form} isSubmitting={mutation.isPending} onFieldChange={handleFieldChange} onSubmit={handleSubmit} />
          <SubmissionHistoryList
            onPageChange={apiMode === "remote" && historyUsernameQuery !== "" ? setHistoryPage : undefined}
            onUsernameChange={handleHistoryUsernameChange}
            page={history.page}
            pageSize={history.pageSize}
            sourceLabel={
              apiMode === "remote" && historyUsernameQuery !== ""
                ? locale === "zh"
                  ? `显示用户 ${historyUsernameQuery} 的远端提交`
                  : `Showing remote submissions for ${historyUsernameQuery}`
                : locale === "zh"
                  ? "显示本地提交记录"
                  : "Showing local submission history"
            }
            submissions={history.items}
            total={history.total}
            username={historyUsername}
          />
        </div>

        <aside className="space-y-4">
          <SubmissionResponsePanel
            data={mutation.data}
            errorMessage={validationError ?? (mutation.isError ? mutation.error.message : undefined)}
          />

          <section className="rounded-lg border border-slate-200 bg-white p-6">
            <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500">{locale === "zh" ? "说明" : "Notes"}</h2>
            <ul className="mt-4 space-y-2 text-sm text-slate-600">
              <li>{locale === "zh" ? "当 `VITE_API_BASE_URL` 为空时，这一页会返回本地的等待中结果。" : "When `VITE_API_BASE_URL` is empty, this page returns a local pending result."}</li>
              <li>{locale === "zh" ? "当前 API 仍要求 `userId` 和 `problemId` 使用后端可解析的值。" : "The current API expects `userId` and `problemId` values that the backend can parse."}</li>
            </ul>
          </section>
        </aside>
      </div>
    </AppShell>
  );
}
