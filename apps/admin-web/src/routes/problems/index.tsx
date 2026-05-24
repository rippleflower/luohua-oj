import type { FormEvent } from "react";
import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import type { AdminProblemCreateInput, AdminProblemUpdateInput, AuthUser } from "@oj/shared";
import { AdminShell } from "../../components/layout/admin-shell";
import { createProblem, listProblems, publishProblem, updateProblem } from "../../features/admin/api";
import { getAuthMe } from "../../features/auth/api";

const defaultCreateDraft: AdminProblemCreateInput = {
  slug: "",
  title: "",
  difficulty: "EASY",
  timeLimitMs: 1000,
  memoryLimitKb: 262144,
  reason: "",
};

export function ProblemsRoute() {
  const queryClient = useQueryClient();
  const [createDraft, setCreateDraft] = useState<AdminProblemCreateInput>(defaultCreateDraft);
  const [feedback, setFeedback] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [activeProblemId, setActiveProblemId] = useState<string | null>(null);

  const { data = [] } = useQuery({
    queryKey: ["admin", "problems"],
    queryFn: listProblems,
    retry: false,
  });
  const { data: viewer } = useQuery({
    queryKey: ["admin", "viewer"],
    queryFn: getAuthMe,
    retry: false,
  });

  const createMutation = useMutation({
    mutationFn: createProblem,
    onSuccess: async () => {
      setFeedback("草稿题目已创建。");
      setErrorMessage("");
      setCreateDraft(defaultCreateDraft);
      await queryClient.invalidateQueries({ queryKey: ["admin", "problems"] });
    },
    onError: (error) => {
      setFeedback("");
      setErrorMessage(error instanceof Error ? error.message : "创建题目失败");
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ problemId, input }: { problemId: string; input: AdminProblemUpdateInput }) => updateProblem(problemId, input),
    onSuccess: async () => {
      setFeedback("题目信息已更新，主站内容会在再次发布后刷新。");
      setErrorMessage("");
      setActiveProblemId(null);
      await queryClient.invalidateQueries({ queryKey: ["admin", "problems"] });
    },
    onError: (error) => {
      setFeedback("");
      setErrorMessage(error instanceof Error ? error.message : "更新题目失败");
      setActiveProblemId(null);
    },
  });

  const publishMutation = useMutation({
    mutationFn: ({ problemId, reason }: { problemId: string; reason: string }) => publishProblem(problemId, { reason }),
    onSuccess: async () => {
      setFeedback("题目已发布，公开读模型已刷新。");
      setErrorMessage("");
      setActiveProblemId(null);
      await queryClient.invalidateQueries({ queryKey: ["admin", "problems"] });
    },
    onError: (error) => {
      setFeedback("");
      setErrorMessage(error instanceof Error ? error.message : "发布题目失败");
      setActiveProblemId(null);
    },
  });

  const canEdit = hasPermission(viewer, "problems.edit");
  const canPublish = hasPermission(viewer, "problems.publish");

  return (
    <AdminShell title="题库管理">
      <div className="grid gap-4">
        <div className="grid gap-3 md:grid-cols-4">
          <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">本轮只开放基础字段：slug、标题、难度、时限、内存。</div>
          <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">题面正文先由后端写入占位文案，后续版本编辑器再补。</div>
          <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">更新数据库后不会自动刷新主站，重新发布才会刷新 public read model。</div>
          <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">发布动作会写审计日志，建议总是填写变更原因。</div>
        </div>
        {feedback ? <div className="rounded-3xl border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800">{feedback}</div> : null}
        {errorMessage ? <div className="rounded-3xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-800">{errorMessage}</div> : null}
        {canEdit ? (
          <form
            className="grid gap-4 rounded-[2rem] border border-slate-200 bg-white p-5"
            onSubmit={(event) => {
              event.preventDefault();
              createMutation.mutate(createDraft);
            }}
          >
            <div className="flex items-end justify-between gap-4">
              <div>
                <p className="text-sm uppercase tracking-[0.2em] text-slate-500">draft bootstrap</p>
                <h2 className="mt-2 text-2xl font-semibold text-slate-950">新建草稿题目</h2>
              </div>
              <button
                type="submit"
                className="rounded-full bg-slate-950 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
                disabled={createMutation.isPending}
              >
                {createMutation.isPending ? "创建中..." : "创建草稿"}
              </button>
            </div>
            <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
              <label className="grid gap-1.5 text-sm text-slate-700">
                <span>Slug</span>
                <input
                  aria-label="新建题目 Slug"
                  className="rounded-2xl border border-slate-200 px-4 py-3"
                  value={createDraft.slug}
                  onChange={(event) => setCreateDraft({ ...createDraft, slug: event.target.value })}
                />
              </label>
              <label className="grid gap-1.5 text-sm text-slate-700">
                <span>标题</span>
                <input
                  aria-label="新建题目 标题"
                  className="rounded-2xl border border-slate-200 px-4 py-3"
                  value={createDraft.title}
                  onChange={(event) => setCreateDraft({ ...createDraft, title: event.target.value })}
                />
              </label>
              <label className="grid gap-1.5 text-sm text-slate-700">
                <span>难度</span>
                <select
                  aria-label="新建题目 难度"
                  className="rounded-2xl border border-slate-200 px-4 py-3"
                  value={createDraft.difficulty}
                  onChange={(event) => setCreateDraft({ ...createDraft, difficulty: event.target.value as AdminProblemCreateInput["difficulty"] })}
                >
                  <option value="EASY">EASY</option>
                  <option value="MEDIUM">MEDIUM</option>
                  <option value="HARD">HARD</option>
                </select>
              </label>
              <label className="grid gap-1.5 text-sm text-slate-700">
                <span>时限 (ms)</span>
                <input
                  aria-label="新建题目 时限"
                  className="rounded-2xl border border-slate-200 px-4 py-3"
                  type="number"
                  min={1}
                  value={createDraft.timeLimitMs}
                  onChange={(event) => setCreateDraft({ ...createDraft, timeLimitMs: Number(event.target.value) })}
                />
              </label>
              <label className="grid gap-1.5 text-sm text-slate-700">
                <span>内存 (KB)</span>
                <input
                  aria-label="新建题目 内存"
                  className="rounded-2xl border border-slate-200 px-4 py-3"
                  type="number"
                  min={1}
                  value={createDraft.memoryLimitKb}
                  onChange={(event) => setCreateDraft({ ...createDraft, memoryLimitKb: Number(event.target.value) })}
                />
              </label>
              <label className="grid gap-1.5 text-sm text-slate-700 xl:col-span-2">
                <span>变更原因</span>
                <input
                  aria-label="新建题目 变更原因"
                  className="rounded-2xl border border-slate-200 px-4 py-3"
                  value={createDraft.reason}
                  onChange={(event) => setCreateDraft({ ...createDraft, reason: event.target.value })}
                />
              </label>
            </div>
          </form>
        ) : null}
        <div className="grid gap-3">
          {data.map((problem) => (
            <article key={problem.id} className="rounded-3xl border border-slate-200 bg-white p-5">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="text-sm text-slate-500">{problem.slug}</p>
                  <h2 className="mt-1 text-lg font-semibold text-slate-950">{problem.title}</h2>
                </div>
                <span className="rounded-full bg-slate-950 px-3 py-1.5 text-xs font-semibold text-white">{problem.status}</span>
              </div>
              <div className="mt-4 grid gap-2 text-sm text-slate-600 md:grid-cols-5">
                <p>难度：{problem.difficulty}</p>
                <p>版本：v{problem.currentVersionNo}</p>
                <p>时限：{problem.timeLimitMs} ms</p>
                <p>内存：{problem.memoryLimitKb} KB</p>
                <p>提交：{problem.submissionCount}</p>
                <p>通过率：{problem.acceptedRate}%</p>
              </div>
              {canEdit ? (
                <form
                  className="mt-5 grid gap-3 rounded-[1.75rem] border border-slate-200 bg-slate-50 p-4"
                  onSubmit={(event) => {
                    event.preventDefault();
                    setActiveProblemId(problem.id);
                    updateMutation.mutate({
                      problemId: problem.id,
                      input: formToProblemInput(event),
                    });
                  }}
                >
                  <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                    <label className="grid gap-1.5 text-sm text-slate-700">
                      <span>Slug</span>
                      <input aria-label={`${problem.slug} slug`} name="slug" className="rounded-2xl border border-slate-200 bg-white px-4 py-3" defaultValue={problem.slug} />
                    </label>
                    <label className="grid gap-1.5 text-sm text-slate-700">
                      <span>标题</span>
                      <input aria-label={`${problem.slug} title`} name="title" className="rounded-2xl border border-slate-200 bg-white px-4 py-3" defaultValue={problem.title} />
                    </label>
                    <label className="grid gap-1.5 text-sm text-slate-700">
                      <span>难度</span>
                      <select aria-label={`${problem.slug} difficulty`} name="difficulty" className="rounded-2xl border border-slate-200 bg-white px-4 py-3" defaultValue={problem.difficulty}>
                        <option value="EASY">EASY</option>
                        <option value="MEDIUM">MEDIUM</option>
                        <option value="HARD">HARD</option>
                      </select>
                    </label>
                    <label className="grid gap-1.5 text-sm text-slate-700">
                      <span>时限 (ms)</span>
                      <input aria-label={`${problem.slug} time limit`} name="timeLimitMs" type="number" min={1} className="rounded-2xl border border-slate-200 bg-white px-4 py-3" defaultValue={problem.timeLimitMs} />
                    </label>
                    <label className="grid gap-1.5 text-sm text-slate-700">
                      <span>内存 (KB)</span>
                      <input aria-label={`${problem.slug} memory limit`} name="memoryLimitKb" type="number" min={1} className="rounded-2xl border border-slate-200 bg-white px-4 py-3" defaultValue={problem.memoryLimitKb} />
                    </label>
                    <label className="grid gap-1.5 text-sm text-slate-700 xl:col-span-2">
                      <span>变更原因</span>
                      <input aria-label={`${problem.slug} reason`} name="reason" className="rounded-2xl border border-slate-200 bg-white px-4 py-3" defaultValue="" />
                    </label>
                  </div>
                  <div className="flex flex-wrap items-center gap-3">
                    <button
                      type="submit"
                      className="rounded-full bg-slate-950 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
                      disabled={updateMutation.isPending && activeProblemId === problem.id}
                    >
                      {updateMutation.isPending && activeProblemId === problem.id ? "保存中..." : "保存修改"}
                    </button>
                    {canPublish ? (
                      <button
                        type="button"
                        className="rounded-full border border-orange-300 bg-orange-50 px-4 py-2 text-sm font-semibold text-orange-700 disabled:opacity-60"
                        disabled={publishMutation.isPending && activeProblemId === problem.id}
                        onClick={(event) => {
                          const form = event.currentTarget.closest("form");
                          if (!(form instanceof HTMLFormElement)) {
                            return;
                          }
                          setActiveProblemId(problem.id);
                          publishMutation.mutate({
                            problemId: problem.id,
                            reason: String(new FormData(form).get("reason") ?? "").trim(),
                          });
                        }}
                      >
                        {publishMutation.isPending && activeProblemId === problem.id ? "发布中..." : problem.isPublished ? "重新发布" : "发布到主站"}
                      </button>
                    ) : null}
                  </div>
                </form>
              ) : null}
            </article>
          ))}
        </div>
      </div>
    </AdminShell>
  );
}

function formToProblemInput(event: FormEvent<HTMLFormElement>): AdminProblemUpdateInput {
  const formData = new FormData(event.currentTarget);
  return {
    slug: String(formData.get("slug") ?? "").trim(),
    title: String(formData.get("title") ?? "").trim(),
    difficulty: String(formData.get("difficulty") ?? "EASY") as AdminProblemUpdateInput["difficulty"],
    timeLimitMs: Number(formData.get("timeLimitMs") ?? 0),
    memoryLimitKb: Number(formData.get("memoryLimitKb") ?? 0),
    reason: String(formData.get("reason") ?? "").trim(),
  };
}

function hasPermission(viewer: AuthUser | undefined, permission: string): boolean {
  if (!viewer) {
    return false;
  }
  if (viewer.role === "SUPER_ADMIN") {
    return true;
  }
  return viewer.permissions.includes(permission as AuthUser["permissions"][number]);
}
