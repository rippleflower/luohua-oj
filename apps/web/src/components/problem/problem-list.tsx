import type { ProblemSummary } from "@oj/shared";

const difficultyClassName: Record<ProblemSummary["difficulty"], string> = {
  EASY: "bg-emerald-50 text-emerald-700 ring-emerald-200",
  MEDIUM: "bg-amber-50 text-amber-700 ring-amber-200",
  HARD: "bg-rose-50 text-rose-700 ring-rose-200",
};

type ProblemListProps = {
  problems: ProblemSummary[];
};

export function ProblemList({ problems }: ProblemListProps) {
  return (
    <div className="overflow-hidden rounded-lg border border-slate-200 bg-white">
      <table className="min-w-full divide-y divide-slate-200 text-sm">
        <thead className="bg-slate-100 text-left text-xs font-semibold uppercase tracking-wide text-slate-600">
          <tr>
            <th className="px-4 py-3">Title</th>
            <th className="px-4 py-3">Difficulty</th>
            <th className="px-4 py-3">Tags</th>
            <th className="px-4 py-3 text-right">Accepted</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-100">
          {problems.map((problem) => (
            <tr key={problem.id} className="hover:bg-slate-50">
              <td className="px-4 py-4 font-medium text-slate-950">
                <a className="hover:underline" href={`/problems/${problem.slug}`}>
                  {problem.title}
                </a>
              </td>
              <td className="px-4 py-4">
                <span
                  className={`inline-flex rounded-full px-2 py-1 text-xs font-semibold ring-1 ring-inset ${
                    difficultyClassName[problem.difficulty]
                  }`}
                >
                  {problem.difficulty}
                </span>
              </td>
              <td className="px-4 py-4">
                <div className="flex flex-wrap gap-2">
                  {problem.tags.map((tag) => (
                    <span key={tag} className="rounded-md bg-slate-100 px-2 py-1 text-xs text-slate-600">
                      {tag}
                    </span>
                  ))}
                </div>
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
