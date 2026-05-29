import type { Dispatch, FormEvent, SetStateAction } from "react";
import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import type {
  AdminProblemContentInput,
  AdminProblemCreateInput,
  AdminProblemDetail,
  AdminProblemUpdateInput,
  AuthUser,
} from "@oj/shared";
import { AdminShell } from "../../components/layout/admin-shell";
import {
  createProblem,
  getProblemDetail,
  listProblems,
  publishProblem,
  updateProblem,
  updateProblemContent,
} from "../../features/admin/api";
import { getAuthMe } from "../../features/auth/api";

type ProblemContentDraft = {
  statement: string;
  input: string;
  output: string;
  constraints: string;
  tags: string;
  samples: Array<{
    input: string;
    output: string;
    weight: number;
  }>;
  reason: string;
};

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
  const [contentProblemId, setContentProblemId] = useState<string | null>(null);
  const [contentDraft, setContentDraft] = useState<ProblemContentDraft | null>(null);

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
  const problemDetailQuery = useQuery({
    queryKey: ["admin", "problem", contentProblemId],
    queryFn: () => getProblemDetail(String(contentProblemId)),
    enabled: contentProblemId !== null,
    retry: false,
  });

  useEffect(() => {
    if (problemDetailQuery.data) {
      setContentDraft(detailToContentDraft(problemDetailQuery.data));
    }
  }, [problemDetailQuery.data]);

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

  const contentMutation = useMutation({
    mutationFn: ({
      problemId,
      input,
    }: {
      problemId: string;
      input: AdminProblemContentInput;
    }) => updateProblemContent(problemId, input),
    onSuccess: async (detail) => {
      setFeedback("题目内容已保存，重新发布后主站会读取这份新内容。");
      setErrorMessage("");
      setActiveProblemId(null);
      setContentDraft(detailToContentDraft(detail));
      await queryClient.invalidateQueries({ queryKey: ["admin", "problem", detail.id] });
      await queryClient.invalidateQueries({ queryKey: ["admin", "problems"] });
    },
    onError: (error) => {
      setFeedback("");
      setErrorMessage(error instanceof Error ? error.message : "保存题目内容失败");
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
      if (contentProblemId) {
        await queryClient.invalidateQueries({ queryKey: ["admin", "problem", contentProblemId] });
      }
    },
    onError: (error) => {
      setFeedback("");
      setErrorMessage(error instanceof Error ? error.message : "发布题目失败");
      setActiveProblemId(null);
    },
  });

  const canEdit = hasPermission(viewer, "problems.edit");
  const canPublish = hasPermission(viewer, "problems.publish");
  const selectedProblem = data.find((item) => item.id === contentProblemId);

  return (
    <AdminShell
      title="题库管理"
      sidebar={
        <div className="grid gap-4">
          <div className="rounded-3xl border border-slate-200 bg-white p-5">
            <h2 className="text-lg font-semibold">操作说明</h2>
            <div className="mt-4 grid gap-3 text-sm text-slate-600">
              <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">基础元数据和题面内容分开编辑，避免一次改动把列表操作拖得过重。</div>
              <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">保存内容只更新草稿；主站公开内容仍以“重新发布”为准。</div>
              <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">公开样例现在直接保存文本内容，不再向用户暴露对象键。</div>
              <div className="rounded-2xl border border-slate-200 bg-slate-50 p-4">标签使用逗号分隔；保存时会自动去重和规范化。</div>
            </div>
          </div>
          <div className="rounded-3xl border border-slate-200 bg-white p-5">
            <h2 className="text-lg font-semibold">题库摘要</h2>
            <div className="mt-4 grid gap-2 text-sm text-slate-600">
              <p>题目总数：{data.length}</p>
              <p>已发布：{data.filter((problem) => problem.isPublished).length}</p>
              <p>草稿：{data.filter((problem) => !problem.isPublished).length}</p>
            </div>
          </div>
        </div>
      }
    >
      <div className="grid gap-4">
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

        {contentProblemId ? (
          <section className="grid gap-4 rounded-[2rem] border border-slate-200 bg-white p-5">
            <div className="flex items-start justify-between gap-4">
              <div>
                <p className="text-sm uppercase tracking-[0.2em] text-slate-500">problem content</p>
                <h2 className="mt-2 text-2xl font-semibold text-slate-950">
                  {selectedProblem?.title ?? "题目内容编辑"}
                </h2>
                <p className="mt-2 text-sm text-slate-600">
                  保存后只更新草稿内容；要让主站读取新题面，请再执行发布。
                </p>
              </div>
              <button
                type="button"
                className="rounded-full border border-slate-200 px-4 py-2 text-sm font-semibold text-slate-700"
                onClick={() => {
                  setContentProblemId(null);
                  setContentDraft(null);
                }}
              >
                收起编辑器
              </button>
            </div>

            {problemDetailQuery.isPending ? (
              <div className="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm text-slate-600">
                正在加载题目内容...
              </div>
            ) : null}

            {problemDetailQuery.isError ? (
              <div className="rounded-2xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
                {problemDetailQuery.error instanceof Error ? problemDetailQuery.error.message : "加载题目内容失败"}
              </div>
            ) : null}

            {contentDraft ? (
              <form
                className="grid gap-4"
                onSubmit={(event) => {
                  event.preventDefault();
                  if (!contentProblemId) {
                    return;
                  }
                  setActiveProblemId(contentProblemId);
                  contentMutation.mutate({
                    problemId: contentProblemId,
                    input: contentDraftToInput(contentDraft),
                  });
                }}
              >
                <div className="grid gap-4 xl:grid-cols-2">
                  {[
                    { key: "statement", label: "题意" },
                    { key: "input", label: "输入格式" },
                    { key: "output", label: "输出格式" },
                    { key: "constraints", label: "说明与约束" },
                  ].map((section) => (
                    <label key={section.key} className="grid gap-1.5 text-sm text-slate-700">
                      <span>{section.label}</span>
                      <textarea
                        aria-label={`题目内容 ${section.label}`}
                        className="min-h-40 rounded-2xl border border-slate-200 px-4 py-3 font-mono text-sm"
                        value={contentDraft[section.key as keyof Omit<ProblemContentDraft, "samples" | "tags" | "reason">]}
                        onChange={(event) =>
                          setContentDraft((current) =>
                            current
                              ? {
                                  ...current,
                                  [section.key]: event.target.value,
                                }
                              : current,
                          )
                        }
                      />
                    </label>
                  ))}
                </div>

                <label className="grid gap-1.5 text-sm text-slate-700">
                  <span>标签（逗号分隔）</span>
                  <input
                    aria-label="题目内容 标签"
                    className="rounded-2xl border border-slate-200 px-4 py-3"
                    value={contentDraft.tags}
                    onChange={(event) =>
                      setContentDraft((current) =>
                        current
                          ? {
                              ...current,
                              tags: event.target.value,
                            }
                          : current,
                      )
                    }
                  />
                </label>

                <div className="grid gap-3">
                  <div className="flex items-center justify-between">
                    <h3 className="text-lg font-semibold text-slate-950">公开样例</h3>
                    <button
                      type="button"
                      className="rounded-full border border-slate-200 px-3 py-1.5 text-sm font-semibold text-slate-700"
                      onClick={() =>
                        setContentDraft((current) =>
                          current
                            ? {
                                ...current,
                                samples: [...current.samples, { input: "", output: "", weight: 1 }],
                              }
                            : current,
                        )
                      }
                    >
                      添加样例
                    </button>
                  </div>
                  {contentDraft.samples.map((sample, index) => (
                    <div key={`sample-${index}`} className="grid gap-3 rounded-3xl border border-slate-200 bg-slate-50 p-4">
                      <div className="flex items-center justify-between">
                        <p className="text-sm font-semibold text-slate-900">样例 {index + 1}</p>
                        <button
                          type="button"
                          className="text-sm font-medium text-rose-700"
                          onClick={() =>
                            setContentDraft((current) =>
                              current
                                ? {
                                    ...current,
                                    samples: current.samples.filter((_, sampleIndex) => sampleIndex !== index),
                                  }
                                : current,
                            )
                          }
                        >
                          删除
                        </button>
                      </div>
                      <div className="grid gap-3 xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_120px]">
                        <label className="grid gap-1.5 text-sm text-slate-700">
                          <span>输入</span>
                          <textarea
                            aria-label={`样例 ${index + 1} 输入`}
                            className="min-h-32 rounded-2xl border border-slate-200 bg-white px-4 py-3 font-mono text-sm"
                            value={sample.input}
                            onChange={(event) => updateSampleDraft(setContentDraft, index, "input", event.target.value)}
                          />
                        </label>
                        <label className="grid gap-1.5 text-sm text-slate-700">
                          <span>输出</span>
                          <textarea
                            aria-label={`样例 ${index + 1} 输出`}
                            className="min-h-32 rounded-2xl border border-slate-200 bg-white px-4 py-3 font-mono text-sm"
                            value={sample.output}
                            onChange={(event) => updateSampleDraft(setContentDraft, index, "output", event.target.value)}
                          />
                        </label>
                        <label className="grid gap-1.5 text-sm text-slate-700">
                          <span>权重</span>
                          <input
                            aria-label={`样例 ${index + 1} 权重`}
                            className="rounded-2xl border border-slate-200 bg-white px-4 py-3"
                            type="number"
                            min={1}
                            value={sample.weight}
                            onChange={(event) => updateSampleDraft(setContentDraft, index, "weight", Number(event.target.value))}
                          />
                        </label>
                      </div>
                    </div>
                  ))}
                </div>

                <label className="grid gap-1.5 text-sm text-slate-700">
                  <span>变更原因</span>
                  <input
                    aria-label="题目内容 变更原因"
                    className="rounded-2xl border border-slate-200 px-4 py-3"
                    value={contentDraft.reason}
                    onChange={(event) =>
                      setContentDraft((current) =>
                        current
                          ? {
                              ...current,
                              reason: event.target.value,
                            }
                          : current,
                      )
                    }
                  />
                </label>

                <div className="flex flex-wrap items-center gap-3">
                  <button
                    type="submit"
                    className="rounded-full bg-slate-950 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
                    disabled={contentMutation.isPending && activeProblemId === contentProblemId}
                  >
                    {contentMutation.isPending && activeProblemId === contentProblemId ? "保存中..." : "保存题面内容"}
                  </button>
                </div>
              </form>
            ) : null}
          </section>
        ) : null}

        <div className="grid gap-3">
          {data.map((problem) => (
            <article key={problem.id} className="rounded-3xl border border-slate-200 bg-white p-5">
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="text-sm text-slate-500">#{problem.problemNo} · {problem.slug}</p>
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
                    <button
                      type="button"
                      className="rounded-full border border-slate-200 bg-white px-4 py-2 text-sm font-semibold text-slate-700"
                      onClick={() => {
                        setContentProblemId(problem.id);
                        setContentDraft(null);
                      }}
                    >
                      编辑题面内容
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

function detailToContentDraft(detail: AdminProblemDetail): ProblemContentDraft {
  return {
    statement: findSection(detail, "statement"),
    input: findSection(detail, "input"),
    output: findSection(detail, "output"),
    constraints: findSection(detail, "constraints"),
    tags: detail.tags.join(", "),
    samples: detail.samples.length > 0 ? detail.samples : [{ input: "", output: "", weight: 1 }],
    reason: "",
  };
}

function findSection(detail: AdminProblemDetail, section: "statement" | "input" | "output" | "constraints") {
  return detail.statementJson.find((item) => item.section === section)?.content ?? "";
}

function contentDraftToInput(draft: ProblemContentDraft): AdminProblemContentInput {
  return {
    statementJson: [
      { kind: "markdown", section: "statement", content: draft.statement },
      { kind: "markdown", section: "input", content: draft.input },
      { kind: "markdown", section: "output", content: draft.output },
      { kind: "markdown", section: "constraints", content: draft.constraints },
    ],
    samples: draft.samples,
    tags: draft.tags
      .split(",")
      .map((tag) => tag.trim())
      .filter((tag) => tag !== ""),
    reason: draft.reason.trim(),
  };
}

function updateSampleDraft(
  setDraft: Dispatch<SetStateAction<ProblemContentDraft | null>>,
  index: number,
  field: "input" | "output" | "weight",
  value: string | number,
) {
  setDraft((current) => {
    if (!current) {
      return current;
    }
    const nextSamples = current.samples.map((sample, sampleIndex) =>
      sampleIndex === index
        ? {
            ...sample,
            [field]: value,
          }
        : sample,
    );
    return {
      ...current,
      samples: nextSamples,
    };
  });
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
