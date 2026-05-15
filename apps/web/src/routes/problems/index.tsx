import { AppShell } from "../../components/layout/app-shell";
import { ProblemList } from "../../components/problem/problem-list";
import { useProblems } from "../../features/problems/hooks";

export function ProblemsRoute() {
  const { data: problems } = useProblems();

  return (
    <AppShell
      title="Problems"
      subtitle="OJ3"
      action={
        <button className="rounded-md bg-slate-950 px-4 py-2 text-sm font-medium text-white hover:bg-slate-800">
          Submit Code
        </button>
      }
    >
      <ProblemList problems={problems} />
    </AppShell>
  );
}
