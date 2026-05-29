import Editor from "@monaco-editor/react";
import { languages, type Language, webRoutes } from "@oj/shared";
import { useEffect, useState } from "react";

import { AppShell } from "../../components/layout/app-shell";
import { ProblemMarkdown } from "../../components/problem/problem-markdown";
import { SubmissionResponsePanel } from "../../components/submission/submission-response-panel";
import { useAuthUser } from "../../features/auth/hooks";
import { useProblem, useProblemByRouteCode } from "../../features/problems/hooks";
import {
  defaultSourceForLanguage,
  deriveProblemTags,
  loadEditorPreferences,
  monacoLanguageFor,
  saveEditorPreferences,
  type EditorPreferences,
} from "../../features/problems/workspace";
import { useCreateSubmission } from "../../features/submissions/hooks";
import { env } from "../../lib/env";
import { useLocale } from "../../lib/locale";

const sectionLabel: Record<string, { zh: string; en: string }> = {
  statement: { zh: "题意", en: "Statement" },
  input: { zh: "输入格式", en: "Input" },
  output: { zh: "输出格式", en: "Output" },
  constraints: { zh: "说明与约束", en: "Notes & Constraints" },
};

export function ProblemDetailRoute({ routeCode, slug }: { routeCode?: string; slug?: string }) {
  const { locale } = useLocale();
  const problemByRouteCode = useProblemByRouteCode(routeCode ?? "");
  const problemBySlug = useProblem(slug ?? "");
  const problemQuery = routeCode ? problemByRouteCode : problemBySlug;
  const problem = problemQuery.data;
  const { data: viewer } = useAuthUser();
  const mutation = useCreateSubmission();
  const [language, setLanguage] = useState<Language>("CPP17");
  const [source, setSource] = useState(() => defaultSourceForLanguage("CPP17"));
  const [compatUserId, setCompatUserId] = useState("");
  const [validationError, setValidationError] = useState<string>();
  const [preferences, setPreferences] = useState<EditorPreferences>(() =>
    loadEditorPreferences(),
  );
  const [settingsOpen, setSettingsOpen] = useState(false);

  useEffect(() => {
    saveEditorPreferences(preferences);
  }, [preferences]);

  useEffect(() => {
    if (viewer?.id && compatUserId === "") {
      setCompatUserId(viewer.id);
    }
  }, [compatUserId, viewer?.id]);

  useEffect(() => {
    if (!routeCode && slug && problem?.routeCode) {
      const search = typeof window === "undefined" ? "" : window.location.search;
      window.history.replaceState({}, "", `${webRoutes.problemDetail(problem.routeCode)}${search}`);
    }
  }, [problem?.routeCode, routeCode, slug]);

  const currentUserId = viewer?.id ?? (env.demoMode ? compatUserId.trim() : "");
  const currentTemplate = defaultSourceForLanguage(language);
  const problemTags = problem ? deriveProblemTags(problem) : [];
  const fromMakeupSlug =
    typeof window === "undefined"
      ? ""
      : new URLSearchParams(window.location.search).get("fromMakeup") ?? "";

  function handleLanguageChange(nextLanguage: Language) {
    const nextTemplate = defaultSourceForLanguage(nextLanguage);
    setSource((current) => (current === currentTemplate ? nextTemplate : current));
    setLanguage(nextLanguage);
  }

  function handleSubmit() {
    if (!problem) {
      return;
    }
    if (currentUserId === "") {
      setValidationError(
        locale === "zh"
          ? env.demoMode
            ? "当前环境还没有用户身份，请先登录或填写兼容用户 ID。"
            : "当前环境还没有用户身份，请先登录后再提交。"
          : env.demoMode
            ? "No user identity is available yet. Log in or provide a compatibility user ID."
            : "No user identity is available yet. Log in before submitting.",
      );
      return;
    }

    setValidationError(undefined);
    mutation.mutate({
      form: {
        userId: currentUserId,
        problemId: problem.id,
        language,
        source,
      },
      requestId: crypto.randomUUID(),
    });
  }

  return (
    <AppShell
      title={problem?.title ?? (locale === "zh" ? "题目详情" : "Problem Detail")}
      subtitle="luooj"
      action={
        <div className="flex flex-wrap gap-3">
          {fromMakeupSlug !== "" ? (
            <a
              className="rounded-full border border-slate-200 bg-white px-4 py-2.5 text-sm font-semibold text-slate-700 hover:text-slate-950"
              href={webRoutes.contestMakeup(fromMakeupSlug)}
            >
              {locale === "zh" ? "返回补题清单" : "Back to Makeup List"}
            </a>
          ) : null}
          <a
            className="rounded-full border border-slate-200 bg-white px-4 py-2.5 text-sm font-semibold text-slate-700 hover:text-slate-950"
            href={webRoutes.problems}
          >
            {locale === "zh" ? "返回题库" : "Back to Problems"}
          </a>
          <a
            className="rounded-full bg-slate-950 px-4 py-2.5 text-sm font-semibold text-white hover:bg-slate-800"
            href={webRoutes.submissions}
          >
            {locale === "zh" ? "查看提交记录" : "Open Submission History"}
          </a>
        </div>
      }
    >
      {problemQuery.isError ? (
        <div className="rounded-3xl border border-rose-200 bg-rose-50 p-8 text-sm text-rose-700">
          {problemQuery.error instanceof Error
            ? problemQuery.error.message
            : locale === "zh"
              ? "题目加载失败。"
              : "Failed to load problem."}
        </div>
      ) : !problem ? (
        <div className="rounded-3xl border border-dashed border-slate-200 bg-white/80 p-8 text-sm text-slate-500">
          {locale === "zh" ? "没有找到这道题。" : "Problem not found."}
        </div>
      ) : (
        <div className="grid gap-4 xl:grid-cols-[minmax(0,1.16fr)_minmax(420px,0.94fr)]">
          <section className="space-y-4">
            <div className="rounded-3xl border border-slate-200/80 bg-white/88 p-5">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">
                    {problem.slug}
                  </p>
                  <h2 className="mt-1.5 text-[1.85rem] font-semibold text-slate-950">
                    {problem.title}
                  </h2>
                </div>
                <div className="flex flex-wrap gap-2">
                  <span className="rounded-full bg-teal-50 px-3 py-1.5 text-xs font-semibold text-teal-700">
                    {problem.difficulty}
                  </span>
                  <span className="rounded-full bg-slate-100 px-3 py-1.5 text-xs font-semibold text-slate-700">
                    {locale === "zh" ? "时间限制" : "Time limit"} {problem.limitsJson.timeLimitMs}ms
                  </span>
                  <span className="rounded-full bg-slate-100 px-3 py-1.5 text-xs font-semibold text-slate-700">
                    {locale === "zh" ? "内存限制" : "Memory limit"} {problem.limitsJson.memoryLimitKb}KB
                  </span>
                </div>
              </div>
              {problemTags.length > 0 ? (
                <div className="mt-4 flex flex-wrap gap-2">
                  {problemTags.map((tag) => (
                    <a
                      key={tag}
                      className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1.5 text-xs font-semibold text-slate-600 transition hover:border-slate-950 hover:text-slate-950"
                      href={`${webRoutes.problems}?tag=${encodeURIComponent(tag)}`}
                    >
                      #{tag}
                    </a>
                  ))}
                </div>
              ) : null}
            </div>

            {problem.statementJson.map((section) => (
              <section
                key={section.section}
                className="rounded-3xl border border-slate-200/80 bg-white/85 p-5"
              >
                <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">
                  {locale === "zh"
                    ? (sectionLabel[section.section]?.zh ?? section.section)
                    : (sectionLabel[section.section]?.en ?? section.section)}
                </p>
                <div className="mt-3">
                  <ProblemMarkdown content={section.content} />
                </div>
              </section>
            ))}

            <section className="rounded-3xl border border-slate-200/80 bg-white/85 p-5">
              <div className="flex items-center justify-between gap-3">
                <div>
                  <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">
                    {locale === "zh" ? "样例与数据入口" : "Samples & Data"}
                  </p>
                  <h3 className="mt-1 text-lg font-semibold text-slate-950">
                    {locale === "zh"
                      ? "公开样例"
                      : "Public samples"}
                  </h3>
                </div>
              </div>
              <div className="mt-4 grid gap-3 lg:grid-cols-2">
                {problem.samplesJson.map((sample, index) => (
                  <div
                    key={`${problem.id}-sample-${index}`}
                    className="rounded-2xl border border-slate-200 bg-slate-50 p-4 shadow-sm"
                  >
                    <p className="text-xs font-semibold uppercase tracking-[0.22em] text-slate-500">
                      {locale === "zh" ? `样例 ${index + 1}` : `Sample ${index + 1}`}
                    </p>
                    <p className="mt-3 text-xs text-slate-500">
                      {locale === "zh" ? "输入" : "Input"}
                    </p>
                    <pre className="mt-1 overflow-x-auto whitespace-pre-wrap break-all rounded-xl bg-slate-950 px-3 py-2 text-xs text-slate-100">
                      {sample.input}
                    </pre>
                    <p className="mt-3 text-xs text-slate-500">
                      {locale === "zh" ? "输出" : "Output"}
                    </p>
                    <pre className="mt-1 overflow-x-auto whitespace-pre-wrap break-all rounded-xl bg-slate-100 px-3 py-2 text-xs text-slate-700">
                      {sample.output}
                    </pre>
                    <p className="mt-3 text-xs text-slate-500">
                      {locale === "zh" ? "权重" : "Weight"} {sample.weight}
                    </p>
                  </div>
                ))}
              </div>
            </section>
          </section>

          <aside className="space-y-4 xl:sticky xl:top-6 xl:self-start">
            <section className="overflow-hidden rounded-3xl border border-slate-200/80 bg-white/92 shadow-[0_16px_50px_rgba(15,23,42,0.06)]">
              <div className="border-b border-slate-200 px-5 py-4">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">
                      {locale === "zh" ? "做题工作区" : "Workspace"}
                    </p>
                    <h3 className="mt-1 text-lg font-semibold text-slate-950">
                      {locale === "zh" ? "代码编辑与提交" : "Code editor & submit"}
                    </h3>
                  </div>
                  <button
                    className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1.5 text-xs font-semibold text-slate-600 hover:bg-slate-100"
                    onClick={() => setSettingsOpen((current) => !current)}
                    type="button"
                  >
                    {locale === "zh" ? "编辑器设置" : "Editor settings"}
                  </button>
                </div>
                {settingsOpen ? (
                  <div className="mt-4 grid gap-3 rounded-2xl border border-slate-200 bg-slate-50 p-3 sm:grid-cols-2">
                    <label className="grid gap-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                      Theme
                      <select
                        className="rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm font-medium normal-case tracking-normal text-slate-700"
                        onChange={(event) =>
                          setPreferences((current) => ({
                            ...current,
                            theme: event.target.value as EditorPreferences["theme"],
                          }))
                        }
                        value={preferences.theme}
                      >
                        <option value="light">Light</option>
                        <option value="vs-dark">Dark</option>
                      </select>
                    </label>
                    <label className="grid gap-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                      Font size
                      <input
                        className="rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm font-medium normal-case tracking-normal text-slate-700"
                        max="24"
                        min="12"
                        onChange={(event) =>
                          setPreferences((current) => ({
                            ...current,
                            fontSize: Number(event.target.value),
                          }))
                        }
                        type="range"
                        value={preferences.fontSize}
                      />
                    </label>
                    <label className="grid gap-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                      Tab size
                      <select
                        className="rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm font-medium normal-case tracking-normal text-slate-700"
                        onChange={(event) =>
                          setPreferences((current) => ({
                            ...current,
                            tabSize: Number(event.target.value),
                          }))
                        }
                        value={preferences.tabSize}
                      >
                        {[2, 4, 8].map((size) => (
                          <option key={size} value={size}>
                            {size}
                          </option>
                        ))}
                      </select>
                    </label>
                    <label className="grid gap-1 text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
                      Word wrap
                      <select
                        className="rounded-xl border border-slate-200 bg-white px-3 py-2 text-sm font-medium normal-case tracking-normal text-slate-700"
                        onChange={(event) =>
                          setPreferences((current) => ({
                            ...current,
                            wordWrap: event.target.value as EditorPreferences["wordWrap"],
                          }))
                        }
                        value={preferences.wordWrap}
                      >
                        <option value="on">{locale === "zh" ? "自动换行" : "Wrap on"}</option>
                        <option value="off">{locale === "zh" ? "关闭换行" : "Wrap off"}</option>
                      </select>
                    </label>
                    <label className="flex items-center gap-2 text-sm text-slate-700">
                      <input
                        checked={preferences.minimap}
                        onChange={(event) =>
                          setPreferences((current) => ({
                            ...current,
                            minimap: event.target.checked,
                          }))
                        }
                        type="checkbox"
                      />
                      {locale === "zh" ? "显示 minimap" : "Show minimap"}
                    </label>
                    <label className="flex items-center gap-2 text-sm text-slate-700">
                      <input
                        checked={preferences.lineNumbers}
                        onChange={(event) =>
                          setPreferences((current) => ({
                            ...current,
                            lineNumbers: event.target.checked,
                          }))
                        }
                        type="checkbox"
                      />
                      {locale === "zh" ? "显示行号" : "Show line numbers"}
                    </label>
                  </div>
                ) : null}
              </div>

              <div className="space-y-4 px-5 py-4">
                <div className="grid gap-3 sm:grid-cols-[minmax(0,1fr)_160px]">
                  <label className="grid gap-2 text-sm font-medium text-slate-700">
                    {locale === "zh" ? "当前用户" : "Current user"}
                    {viewer?.id ? (
                      <div className="rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700">
                        {viewer.displayName} · {viewer.id}
                      </div>
                    ) : env.demoMode ? (
                      <input
                        className="rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700"
                        onChange={(event) => setCompatUserId(event.target.value)}
                        placeholder={locale === "zh" ? "兼容模式用户 ID" : "Compatibility user ID"}
                        value={compatUserId}
                      />
                    ) : (
                      <div className="rounded-2xl border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">
                        {locale === "zh" ? "请先登录后提交。" : "Log in before submitting."}
                      </div>
                    )}
                  </label>
                  <label className="grid gap-2 text-sm font-medium text-slate-700">
                    {locale === "zh" ? "语言" : "Language"}
                    <select
                      className="rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700"
                      onChange={(event) =>
                        handleLanguageChange(event.target.value as Language)
                      }
                      value={language}
                    >
                      {languages.map((item) => (
                        <option key={item} value={item}>
                          {item}
                        </option>
                      ))}
                    </select>
                  </label>
                </div>

                <div className="overflow-hidden rounded-[1.5rem] border border-slate-200">
                  <Editor
                    height="520px"
                    language={monacoLanguageFor(language)}
                    onChange={(value) => setSource(value ?? "")}
                    options={{
                      automaticLayout: true,
                      fontSize: preferences.fontSize,
                      lineNumbers: preferences.lineNumbers ? "on" : "off",
                      minimap: { enabled: preferences.minimap },
                      scrollBeyondLastLine: false,
                      tabSize: preferences.tabSize,
                      wordWrap: preferences.wordWrap,
                    }}
                    theme={preferences.theme}
                    value={source}
                  />
                </div>

                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div className="text-xs text-slate-500">
                    {locale === "zh"
                      ? env.demoMode
                        ? "题目 ID 已自动绑定；demo 模式下仍允许兼容用户 ID。"
                        : "题目 ID 已自动绑定；提交需要真实登录身份。"
                      : env.demoMode
                        ? "Problem ID is injected automatically; demo mode still allows a compatibility user ID."
                        : "Problem ID is injected automatically; submission requires a real signed-in user."}
                  </div>
                  <button
                    className="rounded-full bg-slate-950 px-4 py-2.5 text-sm font-semibold text-white hover:bg-slate-800 disabled:bg-slate-400"
                    disabled={mutation.isPending}
                    onClick={handleSubmit}
                    type="button"
                  >
                    {mutation.isPending
                      ? locale === "zh"
                        ? "提交中..."
                        : "Submitting..."
                      : locale === "zh"
                        ? "提交这道题"
                        : "Submit this problem"}
                  </button>
                </div>
              </div>
            </section>

            <SubmissionResponsePanel
              data={mutation.data}
              errorMessage={
                validationError ??
                (mutation.isError ? mutation.error.message : undefined)
              }
            />

            <section className="rounded-3xl border border-slate-200/80 bg-white/85 p-4">
              <p className="font-mono text-[11px] uppercase tracking-[0.28em] text-slate-500">
                {locale === "zh" ? "元数据" : "Metadata"}
              </p>
              <div className="mt-3 space-y-2">
                {Object.entries(problem.metadataJson).map(([key, value]) => (
                  <div
                    key={key}
                    className="rounded-2xl border border-slate-200 bg-slate-50 px-3 py-2.5"
                  >
                    <p className="text-xs uppercase tracking-[0.22em] text-slate-400">
                      {key}
                    </p>
                    <p className="mt-1 break-all text-sm font-medium text-slate-900">
                      {typeof value === "string"
                        ? value
                        : JSON.stringify(value)}
                    </p>
                  </div>
                ))}
              </div>
            </section>
          </aside>
        </div>
      )}
    </AppShell>
  );
}
