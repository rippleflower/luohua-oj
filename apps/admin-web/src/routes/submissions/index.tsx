import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import type { AuthUser } from "@oj/shared";
import { AdminShell } from "../../components/layout/admin-shell";
import {
  getJudgeQueueSummary,
  listSubmissions,
  rejudgeSubmission,
} from "../../features/admin/api";
import { getAuthMe } from "../../features/auth/api";

export function SubmissionsRoute() {
  const queryClient = useQueryClient();
  const [feedback, setFeedback] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [activeSubmissionId, setActiveSubmissionId] = useState<string | null>(
    null,
  );
  const [rejudgeReasons, setRejudgeReasons] = useState<Record<string, string>>(
    {},
  );

  const { data: submissions = [] } = useQuery({
    queryKey: ["admin", "submissions"],
    queryFn: listSubmissions,
    retry: false,
  });
  const { data: queueSummary } = useQuery({
    queryKey: ["admin", "judge", "queue"],
    queryFn: getJudgeQueueSummary,
    retry: false,
  });
  const { data: viewer } = useQuery({
    queryKey: ["admin", "viewer"],
    queryFn: getAuthMe,
    retry: false,
  });

  const rejudgeMutation = useMutation({
    mutationFn: ({
      submissionId,
      reason,
    }: {
      submissionId: string;
      reason: string;
    }) => rejudgeSubmission(submissionId, { reason }),
    onSuccess: async () => {
      setFeedback("重判任务已入队。");
      setErrorMessage("");
      setActiveSubmissionId(null);
      await queryClient.invalidateQueries({
        queryKey: ["admin", "judge", "queue"],
      });
    },
    onError: (error) => {
      setFeedback("");
      setErrorMessage(error instanceof Error ? error.message : "重判失败");
      setActiveSubmissionId(null);
    },
  });

  const canRejudge = hasPermission(viewer, "submissions.rejudge");

  return (
    <AdminShell title="提交与判题">
      <div className="grid gap-4">
        <div className="grid gap-3 md:grid-cols-3">
          <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">
            队列摘要直接从 Asynq inspector 读取，前端不接触 Redis。
          </div>
          <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">
            重判前会清空旧结果并回到 `PENDING`，避免旧测试点结果污染。
          </div>
          <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">
            比赛提交重判优先沿用冻结时记录的 snapshot 版本元数据。
          </div>
        </div>
        {feedback ? (
          <div className="rounded-3xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800">
            {feedback}
          </div>
        ) : null}
        {errorMessage ? (
          <div className="rounded-3xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800">
            {errorMessage}
          </div>
        ) : null}

        <section className="grid gap-3 md:grid-cols-4">
          <QueueMetric label="Pending" value={queueSummary?.pending ?? 0} />
          <QueueMetric label="Active" value={queueSummary?.active ?? 0} />
          <QueueMetric label="Retry" value={queueSummary?.retry ?? 0} />
          <QueueMetric
            label="Processed Today"
            value={queueSummary?.processedToday ?? 0}
          />
        </section>

        <section className="rounded-[2rem] border border-slate-200 bg-white p-5">
          <div className="flex items-end justify-between gap-4">
            <div>
              <p className="text-sm uppercase tracking-[0.2em] text-slate-500">
                judge queue
              </p>
              <h2 className="mt-2 text-2xl font-semibold text-slate-950">
                {queueSummary?.queue ?? "judge"}
              </h2>
            </div>
            <div className="rounded-full border border-slate-200 px-4 py-2 text-sm text-slate-600">
              {queueSummary?.paused ? "Paused" : "Running"} · latency{" "}
              {queueSummary?.latencySeconds ?? 0}s
            </div>
          </div>
          <div className="mt-4 overflow-hidden rounded-3xl border border-slate-200">
            <table className="min-w-full border-collapse text-sm">
              <thead className="bg-slate-50 text-left text-slate-500">
                <tr>
                  {["任务", "提交", "状态", "下次处理", "错误"].map((label) => (
                    <th key={label} className="px-4 py-3 font-medium">
                      {label}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {(queueSummary?.recentTasks ?? []).map((task) => (
                  <tr key={task.id} className="border-t border-slate-200">
                    <td className="px-4 py-3 font-mono text-xs text-slate-500">
                      {task.id.slice(0, 8)}
                    </td>
                    <td className="px-4 py-3 font-mono text-xs text-slate-500">
                      {task.submissionId?.slice(0, 8) ?? "-"}
                    </td>
                    <td className="px-4 py-3">{task.state}</td>
                    <td className="px-4 py-3">
                      {task.nextProcessAt
                        ? new Date(task.nextProcessAt).toLocaleString()
                        : "-"}
                    </td>
                    <td className="px-4 py-3 text-slate-600">
                      {task.lastErr || "-"}
                    </td>
                  </tr>
                ))}
                {(queueSummary?.recentTasks ?? []).length === 0 ? (
                  <tr className="border-t border-slate-200">
                    <td
                      className="px-4 py-6 text-sm text-slate-500"
                      colSpan={5}
                    >
                      当前没有最近任务。
                    </td>
                  </tr>
                ) : null}
              </tbody>
            </table>
          </div>
        </section>

        <div className="overflow-hidden rounded-[2rem] border border-slate-200 bg-white">
          <table className="min-w-full border-collapse text-sm">
            <thead className="bg-slate-50 text-left text-slate-500">
              <tr>
                {[
                  "提交",
                  "用户",
                  "题目",
                  "语言",
                  "状态",
                  "时间",
                  "重判原因",
                  "动作",
                ].map((label) => (
                  <th key={label} className="px-4 py-3 font-medium">
                    {label}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              {submissions.map((submission) => (
                <tr key={submission.id} className="border-t border-slate-200">
                  <td className="px-4 py-3 font-mono text-xs text-slate-500">
                    {submission.id.slice(0, 8)}
                  </td>
                  <td className="px-4 py-3">{submission.username}</td>
                  <td className="px-4 py-3">{submission.problemTitle}</td>
                  <td className="px-4 py-3">{submission.language}</td>
                  <td className="px-4 py-3">{submission.status}</td>
                  <td className="px-4 py-3">
                    {new Date(submission.createdAt).toLocaleString()}
                  </td>
                  <td className="px-4 py-3">
                    <input
                      aria-label={`${submission.id} rejudge reason`}
                      className="w-full rounded-2xl border border-slate-200 px-3 py-2"
                      value={rejudgeReasons[submission.id] ?? ""}
                      onChange={(event) =>
                        setRejudgeReasons((current) => ({
                          ...current,
                          [submission.id]: event.target.value,
                        }))
                      }
                    />
                  </td>
                  <td className="px-4 py-3">
                    {canRejudge ? (
                      <button
                        type="button"
                        className="rounded-full border border-orange-300 bg-orange-50 px-4 py-2 text-sm font-semibold text-orange-700 disabled:opacity-60"
                        disabled={
                          rejudgeMutation.isPending &&
                          activeSubmissionId === submission.id
                        }
                        onClick={() => {
                          setActiveSubmissionId(submission.id);
                          rejudgeMutation.mutate({
                            submissionId: submission.id,
                            reason: rejudgeReasons[submission.id] ?? "",
                          });
                        }}
                      >
                        {rejudgeMutation.isPending &&
                        activeSubmissionId === submission.id
                          ? "入队中..."
                          : "重判"}
                      </button>
                    ) : (
                      <span className="text-slate-400">无权限</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </AdminShell>
  );
}

function QueueMetric({ label, value }: { label: string; value: number }) {
  return (
    <div className="rounded-3xl border border-slate-200 bg-white p-4">
      <p className="text-sm uppercase tracking-[0.2em] text-slate-500">
        {label}
      </p>
      <p className="mt-2 text-3xl font-semibold text-slate-950">{value}</p>
    </div>
  );
}

function hasPermission(
  viewer: AuthUser | undefined,
  permission: string,
): boolean {
  if (!viewer) {
    return false;
  }
  if (viewer.role === "SUPER_ADMIN") {
    return true;
  }
  return viewer.permissions.includes(
    permission as AuthUser["permissions"][number],
  );
}
