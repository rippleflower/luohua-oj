import { languages } from "@oj/shared";
import type { FormEvent } from "react";

import { useLocale } from "../../lib/locale";
import type { CreateSubmissionRequest, SubmissionLogContext } from "../../features/submissions/schema";

type SubmissionFormProps = {
  form: CreateSubmissionRequest;
  isSubmitting: boolean;
  onFieldChange: (field: NonNullable<SubmissionLogContext["field"]>, value: string) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
};

export function SubmissionForm({ form, isSubmitting, onFieldChange, onSubmit }: SubmissionFormProps) {
  const { locale } = useLocale();
  return (
    <form className="space-y-5 rounded-lg border border-slate-200 bg-white p-6" onSubmit={onSubmit}>
      <div className="grid gap-4 sm:grid-cols-2">
        <label className="grid gap-2 text-sm font-medium text-slate-700">
          {locale === "zh" ? "用户 ID" : "User ID"}
          <input
            className="rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-950 outline-none ring-0 transition focus:border-slate-500"
            name="userId"
            onChange={(event) => onFieldChange("userId", event.target.value)}
            placeholder={locale === "zh" ? "UUID 或本地演示用户 ID" : "UUID or local demo user id"}
            value={form.userId}
          />
        </label>
        <label className="grid gap-2 text-sm font-medium text-slate-700">
          {locale === "zh" ? "题目 ID" : "Problem ID"}
          <input
            className="rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-950 outline-none ring-0 transition focus:border-slate-500"
            name="problemId"
            onChange={(event) => onFieldChange("problemId", event.target.value)}
            placeholder={locale === "zh" ? "题目标识符" : "Problem identifier"}
            value={form.problemId}
          />
        </label>
      </div>

      <label className="grid gap-2 text-sm font-medium text-slate-700">
        {locale === "zh" ? "语言" : "Language"}
        <select
          className="rounded-md border border-slate-300 px-3 py-2 text-sm text-slate-950 outline-none ring-0 transition focus:border-slate-500"
          name="language"
          onChange={(event) => onFieldChange("language", event.target.value)}
          value={form.language}
        >
          {languages.map((language) => (
            <option key={language} value={language}>
              {language}
            </option>
          ))}
        </select>
      </label>

      <label className="grid gap-2 text-sm font-medium text-slate-700">
        {locale === "zh" ? "源码" : "Source"}
        <textarea
          className="min-h-[360px] rounded-md border border-slate-300 px-3 py-2 font-mono text-sm text-slate-950 outline-none ring-0 transition focus:border-slate-500"
          name="source"
          onChange={(event) => onFieldChange("source", event.target.value)}
          spellCheck={false}
          value={form.source}
        />
      </label>

      <div className="flex items-center gap-3">
        <button
          className="rounded-md bg-slate-950 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800 disabled:cursor-not-allowed disabled:bg-slate-400"
          disabled={isSubmitting}
          type="submit"
        >
          {isSubmitting ? (locale === "zh" ? "提交中..." : "Submitting...") : locale === "zh" ? "创建提交" : "Create Submission"}
        </button>
      </div>
    </form>
  );
}
