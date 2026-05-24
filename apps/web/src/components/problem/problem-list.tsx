import type { ProblemSummary } from "@oj/shared";
import { useLocale } from "../../lib/locale";

const difficultyClassName: Record<ProblemSummary["difficulty"], string> = {
  EASY: "bg-emerald-50 text-emerald-700 ring-emerald-200",
  MEDIUM: "bg-amber-50 text-amber-700 ring-amber-200",
  HARD: "bg-rose-50 text-rose-700 ring-rose-200",
};

type ProblemRow = ProblemSummary & {
  status?: "UNSOLVED" | "ATTEMPTED" | "SOLVED";
};

type ProblemListProps = {
  problems: ProblemRow[];
};

const statusClassName: Record<NonNullable<ProblemRow["status"]>, string> = {
  UNSOLVED: "bg-slate-100 text-slate-600",
  ATTEMPTED: "bg-amber-100 text-amber-700",
  SOLVED: "bg-teal-100 text-teal-700",
};

export function ProblemList({ problems }: ProblemListProps) {
  const { locale } = useLocale();
  const statusLabel: Record<NonNullable<ProblemRow["status"]>, string> = {
    UNSOLVED: locale === "zh" ? "未完成" : "UNSOLVED",
    ATTEMPTED: locale === "zh" ? "尝试过" : "ATTEMPTED",
    SOLVED: locale === "zh" ? "已通过" : "SOLVED",
  };
  const difficultyLabel: Record<ProblemSummary["difficulty"], string> = {
    EASY: locale === "zh" ? "简单" : "EASY",
    MEDIUM: locale === "zh" ? "中等" : "MEDIUM",
    HARD: locale === "zh" ? "困难" : "HARD",
  };

  return (
    <div className="overflow-hidden rounded-3xl border border-slate-200/80 bg-white/85 shadow-[0_16px_50px_rgba(15,23,42,0.04)]">
      <table className="min-w-full divide-y divide-slate-200 text-sm">
        <thead className="bg-slate-100/90 text-left text-[11px] font-semibold uppercase tracking-[0.22em] text-slate-500">
          <tr>
            <th className="px-4 py-3">{locale === "zh" ? "标题" : "Title"}</th>
            <th className="px-4 py-3">{locale === "zh" ? "难度" : "Difficulty"}</th>
            <th className="px-4 py-3">{locale === "zh" ? "标签" : "Tags"}</th>
            <th className="px-4 py-3">{locale === "zh" ? "状态" : "State"}</th>
            <th className="px-4 py-3 text-right">{locale === "zh" ? "通过率" : "Accepted"}</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-100">
          {problems.map((problem) => (
            <tr key={problem.id} className="hover:bg-slate-50/80">
              <td className="px-4 py-4 text-slate-950">
                <a className="block" href={`/problems/${problem.slug}`}>
                  <div className="font-semibold">{problem.title}</div>
                  <div className="mt-1 font-mono text-[11px] uppercase tracking-[0.18em] text-slate-400">{problem.slug}</div>
                </a>
              </td>
              <td className="px-4 py-4">
                <span
                  className={`inline-flex rounded-full px-2 py-1 text-xs font-semibold ring-1 ring-inset ${
                    difficultyClassName[problem.difficulty]
                  }`}
                >
                  {difficultyLabel[problem.difficulty]}
                </span>
              </td>
              <td className="px-4 py-4">
                <div className="flex flex-wrap gap-2">
                  {problem.tags.map((tag) => (
                    <span key={tag} className="rounded-full bg-slate-100 px-3 py-1 text-xs text-slate-600">
                      {tag}
                    </span>
                  ))}
                </div>
              </td>
              <td className="px-4 py-4">
                {problem.status ? (
                  <span className={`inline-flex rounded-full px-3 py-1 text-xs font-semibold ${statusClassName[problem.status]}`}>
                    {statusLabel[problem.status]}
                  </span>
                ) : (
                  <span className="text-xs text-slate-400">{locale === "zh" ? "未知" : "UNKNOWN"}</span>
                )}
              </td>
              <td className="px-4 py-4 text-right tabular-nums text-slate-700">
                {problem.acceptedRate.toFixed(1)}%
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
