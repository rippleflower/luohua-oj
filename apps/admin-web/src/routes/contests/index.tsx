import type { Dispatch, FormEvent, SetStateAction } from "react";
import { useEffect, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import type {
  AdminContestCreateInput,
  AdminContestProblemBindingInput,
  AdminContestUpdateInput,
  AuthUser,
} from "@oj/shared";
import { AdminShell } from "../../components/layout/admin-shell";
import {
  createContest,
  freezeContest,
  listContests,
  listProblems,
  replaceContestProblems,
  updateContest,
} from "../../features/admin/api";
import { getAuthMe } from "../../features/auth/api";

type ContestDraftForm = {
  slug: string;
  title: string;
  description: string;
  status: AdminContestCreateInput["status"];
  startsAt: string;
  endsAt: string;
  reason: string;
};

const defaultCreateDraft: ContestDraftForm = {
  slug: "",
  title: "",
  description: "",
  status: "UPCOMING",
  startsAt: defaultLocalDateTime(1),
  endsAt: defaultLocalDateTime(3),
  reason: "",
};

export function ContestsRoute() {
  const queryClient = useQueryClient();
  const [createDraft, setCreateDraft] =
    useState<ContestDraftForm>(defaultCreateDraft);
  const [feedback, setFeedback] = useState("");
  const [errorMessage, setErrorMessage] = useState("");
  const [activeContestId, setActiveContestId] = useState<string | null>(null);
  const [problemDrafts, setProblemDrafts] = useState<
    Record<string, AdminContestProblemBindingInput[]>
  >({});

  const { data: contests = [] } = useQuery({
    queryKey: ["admin", "contests"],
    queryFn: listContests,
    retry: false,
  });
  const { data: problems = [] } = useQuery({
    queryKey: ["admin", "problems", "options"],
    queryFn: listProblems,
    retry: false,
  });
  const { data: viewer } = useQuery({
    queryKey: ["admin", "viewer"],
    queryFn: getAuthMe,
    retry: false,
  });

  useEffect(() => {
    setProblemDrafts((current) => {
      const next = { ...current };
      let changed = false;
      for (const contest of contests) {
        if (!next[contest.id]) {
          changed = true;
          next[contest.id] = contest.problems.map((problem) => ({
            problemId: problem.problemId,
            code: problem.code,
            position: problem.position,
          }));
        }
      }
      return changed ? next : current;
    });
  }, [contests]);

  const createMutation = useMutation({
    mutationFn: createContest,
    onSuccess: async () => {
      setFeedback("比赛草稿已创建。");
      setErrorMessage("");
      setCreateDraft(defaultCreateDraft);
      await queryClient.invalidateQueries({ queryKey: ["admin", "contests"] });
    },
    onError: (error) => {
      setFeedback("");
      setErrorMessage(error instanceof Error ? error.message : "创建比赛失败");
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({
      contestId,
      input,
    }: {
      contestId: string;
      input: AdminContestUpdateInput;
    }) => updateContest(contestId, input),
    onSuccess: async () => {
      setFeedback("比赛基础信息已更新。");
      setErrorMessage("");
      setActiveContestId(null);
      await queryClient.invalidateQueries({ queryKey: ["admin", "contests"] });
    },
    onError: (error) => {
      setFeedback("");
      setErrorMessage(error instanceof Error ? error.message : "更新比赛失败");
      setActiveContestId(null);
    },
  });

  const replaceProblemsMutation = useMutation({
    mutationFn: ({
      contestId,
      draft,
    }: {
      contestId: string;
      draft: AdminContestProblemBindingInput[];
    }) =>
      replaceContestProblems(contestId, {
        problems: draft,
        reason: "refresh contest problem set",
      }),
    onSuccess: async () => {
      setFeedback("比赛题目编排已更新。");
      setErrorMessage("");
      setActiveContestId(null);
      await queryClient.invalidateQueries({ queryKey: ["admin", "contests"] });
    },
    onError: (error) => {
      setFeedback("");
      setErrorMessage(
        error instanceof Error ? error.message : "更新题目编排失败",
      );
      setActiveContestId(null);
    },
  });

  const freezeMutation = useMutation({
    mutationFn: ({
      contestId,
      reason,
    }: {
      contestId: string;
      reason: string;
    }) => freezeContest(contestId, { reason }),
    onSuccess: async () => {
      setFeedback("比赛快照已冻结并刷新主站读模型。");
      setErrorMessage("");
      setActiveContestId(null);
      await queryClient.invalidateQueries({ queryKey: ["admin", "contests"] });
    },
    onError: (error) => {
      setFeedback("");
      setErrorMessage(error instanceof Error ? error.message : "冻结比赛失败");
      setActiveContestId(null);
    },
  });

  const canEdit = hasPermission(viewer, "contests.edit");
  const canPublish = hasPermission(viewer, "contests.publish");

  return (
    <AdminShell title="比赛管理">
      <div className="grid gap-4">
        <div className="grid gap-3 md:grid-cols-4">
          <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">
            编辑只修改草稿配置，公开站不会立即变化。
          </div>
          <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">
            冻结会生成新的不可变 snapshot，并刷新 `/contests` 读模型。
          </div>
          <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">
            题目编排采用整表替换，避免局部补丁把顺序打乱。
          </div>
          <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">
            题面展示继续保持纯文本边界，本轮不引入富文本。
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
        {canEdit ? (
          <form
            className="grid gap-4 rounded-[2rem] border border-slate-200 bg-white p-5"
            onSubmit={(event) => {
              event.preventDefault();
              createMutation.mutate(toContestInput(createDraft));
            }}
          >
            <div className="flex items-end justify-between gap-4">
              <div>
                <p className="text-sm uppercase tracking-[0.2em] text-slate-500">
                  contest bootstrap
                </p>
                <h2 className="mt-2 text-2xl font-semibold text-slate-950">
                  新建比赛草稿
                </h2>
              </div>
              <button
                type="submit"
                className="rounded-full bg-slate-950 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
                disabled={createMutation.isPending}
              >
                {createMutation.isPending ? "创建中..." : "创建比赛"}
              </button>
            </div>
            <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
              <label className="grid gap-1.5 text-sm text-slate-700">
                <span>Slug</span>
                <input
                  aria-label="新建比赛 Slug"
                  className="rounded-2xl border border-slate-200 px-4 py-3"
                  value={createDraft.slug}
                  onChange={(event) =>
                    setCreateDraft({ ...createDraft, slug: event.target.value })
                  }
                />
              </label>
              <label className="grid gap-1.5 text-sm text-slate-700">
                <span>标题</span>
                <input
                  aria-label="新建比赛 标题"
                  className="rounded-2xl border border-slate-200 px-4 py-3"
                  value={createDraft.title}
                  onChange={(event) =>
                    setCreateDraft({
                      ...createDraft,
                      title: event.target.value,
                    })
                  }
                />
              </label>
              <label className="grid gap-1.5 text-sm text-slate-700">
                <span>状态</span>
                <select
                  aria-label="新建比赛 状态"
                  className="rounded-2xl border border-slate-200 px-4 py-3"
                  value={createDraft.status}
                  onChange={(event) =>
                    setCreateDraft({
                      ...createDraft,
                      status: event.target.value as ContestDraftForm["status"],
                    })
                  }
                >
                  <option value="UPCOMING">UPCOMING</option>
                  <option value="RUNNING">RUNNING</option>
                  <option value="ENDED">ENDED</option>
                </select>
              </label>
              <label className="grid gap-1.5 text-sm text-slate-700">
                <span>开始时间</span>
                <input
                  aria-label="新建比赛 开始时间"
                  className="rounded-2xl border border-slate-200 px-4 py-3"
                  type="datetime-local"
                  value={createDraft.startsAt}
                  onChange={(event) =>
                    setCreateDraft({
                      ...createDraft,
                      startsAt: event.target.value,
                    })
                  }
                />
              </label>
              <label className="grid gap-1.5 text-sm text-slate-700">
                <span>结束时间</span>
                <input
                  aria-label="新建比赛 结束时间"
                  className="rounded-2xl border border-slate-200 px-4 py-3"
                  type="datetime-local"
                  value={createDraft.endsAt}
                  onChange={(event) =>
                    setCreateDraft({
                      ...createDraft,
                      endsAt: event.target.value,
                    })
                  }
                />
              </label>
              <label className="grid gap-1.5 text-sm text-slate-700 xl:col-span-3">
                <span>简介</span>
                <textarea
                  aria-label="新建比赛 简介"
                  className="min-h-28 rounded-2xl border border-slate-200 px-4 py-3"
                  value={createDraft.description}
                  onChange={(event) =>
                    setCreateDraft({
                      ...createDraft,
                      description: event.target.value,
                    })
                  }
                />
              </label>
              <label className="grid gap-1.5 text-sm text-slate-700 xl:col-span-3">
                <span>变更原因</span>
                <input
                  aria-label="新建比赛 变更原因"
                  className="rounded-2xl border border-slate-200 px-4 py-3"
                  value={createDraft.reason}
                  onChange={(event) =>
                    setCreateDraft({
                      ...createDraft,
                      reason: event.target.value,
                    })
                  }
                />
              </label>
            </div>
          </form>
        ) : null}
        <div className="grid gap-3">
          {contests.map((contest) => {
            const draft = problemDrafts[contest.id] ?? [];
            return (
              <article
                key={contest.id}
                className="rounded-3xl border border-slate-200 bg-white p-5"
              >
                <div className="flex items-start justify-between gap-4">
                  <div>
                    <p className="text-sm text-slate-500">{contest.slug}</p>
                    <h2 className="mt-1 text-lg font-semibold text-slate-950">
                      {contest.title}
                    </h2>
                  </div>
                  <span className="rounded-full bg-orange-100 px-3 py-1.5 text-xs font-semibold text-orange-700">
                    {contest.status}
                  </span>
                </div>
                <div className="mt-4 grid gap-2 text-sm text-slate-600 md:grid-cols-5">
                  <p>题目数：{contest.problemCount}</p>
                  <p>参赛人数：{contest.participantCount}</p>
                  <p>最新快照：#{contest.latestSnapshotNo}</p>
                  <p>{new Date(contest.startsAt).toLocaleString()}</p>
                  <p>{new Date(contest.endsAt).toLocaleString()}</p>
                </div>
                <p className="mt-3 text-sm leading-6 text-slate-600">
                  {contest.description || "当前没有填写比赛简介。"}
                </p>

                {canEdit ? (
                  <form
                    className="mt-5 grid gap-3 rounded-[1.75rem] border border-slate-200 bg-slate-50 p-4"
                    onSubmit={(event) => {
                      event.preventDefault();
                      setActiveContestId(contest.id);
                      updateMutation.mutate({
                        contestId: contest.id,
                        input: formToContestInput(event),
                      });
                    }}
                  >
                    <div className="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
                      <label className="grid gap-1.5 text-sm text-slate-700">
                        <span>Slug</span>
                        <input
                          aria-label={`${contest.slug} slug`}
                          name="slug"
                          className="rounded-2xl border border-slate-200 bg-white px-4 py-3"
                          defaultValue={contest.slug}
                        />
                      </label>
                      <label className="grid gap-1.5 text-sm text-slate-700">
                        <span>标题</span>
                        <input
                          aria-label={`${contest.slug} title`}
                          name="title"
                          className="rounded-2xl border border-slate-200 bg-white px-4 py-3"
                          defaultValue={contest.title}
                        />
                      </label>
                      <label className="grid gap-1.5 text-sm text-slate-700">
                        <span>状态</span>
                        <select
                          aria-label={`${contest.slug} status`}
                          name="status"
                          className="rounded-2xl border border-slate-200 bg-white px-4 py-3"
                          defaultValue={contest.status}
                        >
                          <option value="UPCOMING">UPCOMING</option>
                          <option value="RUNNING">RUNNING</option>
                          <option value="ENDED">ENDED</option>
                        </select>
                      </label>
                      <label className="grid gap-1.5 text-sm text-slate-700">
                        <span>开始时间</span>
                        <input
                          aria-label={`${contest.slug} startsAt`}
                          name="startsAt"
                          type="datetime-local"
                          className="rounded-2xl border border-slate-200 bg-white px-4 py-3"
                          defaultValue={toLocalDateTimeInput(contest.startsAt)}
                        />
                      </label>
                      <label className="grid gap-1.5 text-sm text-slate-700">
                        <span>结束时间</span>
                        <input
                          aria-label={`${contest.slug} endsAt`}
                          name="endsAt"
                          type="datetime-local"
                          className="rounded-2xl border border-slate-200 bg-white px-4 py-3"
                          defaultValue={toLocalDateTimeInput(contest.endsAt)}
                        />
                      </label>
                      <label className="grid gap-1.5 text-sm text-slate-700 xl:col-span-3">
                        <span>简介</span>
                        <textarea
                          aria-label={`${contest.slug} description`}
                          name="description"
                          className="min-h-28 rounded-2xl border border-slate-200 bg-white px-4 py-3"
                          defaultValue={contest.description}
                        />
                      </label>
                      <label className="grid gap-1.5 text-sm text-slate-700 xl:col-span-3">
                        <span>变更原因</span>
                        <input
                          aria-label={`${contest.slug} reason`}
                          name="reason"
                          className="rounded-2xl border border-slate-200 bg-white px-4 py-3"
                          defaultValue=""
                        />
                      </label>
                    </div>
                    <div className="flex flex-wrap items-center gap-3">
                      <button
                        type="submit"
                        className="rounded-full bg-slate-950 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
                        disabled={
                          updateMutation.isPending &&
                          activeContestId === contest.id
                        }
                      >
                        {updateMutation.isPending &&
                        activeContestId === contest.id
                          ? "保存中..."
                          : "保存比赛信息"}
                      </button>
                      {canPublish ? (
                        <button
                          type="button"
                          className="rounded-full border border-orange-300 bg-orange-50 px-4 py-2 text-sm font-semibold text-orange-700 disabled:opacity-60"
                          disabled={
                            freezeMutation.isPending &&
                            activeContestId === contest.id
                          }
                          onClick={(event) => {
                            const form = event.currentTarget.closest("form");
                            if (!(form instanceof HTMLFormElement)) {
                              return;
                            }
                            setActiveContestId(contest.id);
                            freezeMutation.mutate({
                              contestId: contest.id,
                              reason: String(
                                new FormData(form).get("reason") ?? "",
                              ).trim(),
                            });
                          }}
                        >
                          {freezeMutation.isPending &&
                          activeContestId === contest.id
                            ? "冻结中..."
                            : "冻结并发布"}
                        </button>
                      ) : null}
                    </div>
                  </form>
                ) : null}

                {canEdit ? (
                  <section className="mt-5 rounded-[1.75rem] border border-slate-200 bg-slate-50 p-4">
                    <div className="flex items-end justify-between gap-4">
                      <div>
                        <p className="text-sm uppercase tracking-[0.2em] text-slate-500">
                          problem bindings
                        </p>
                        <h3 className="mt-2 text-xl font-semibold text-slate-950">
                          题目编排
                        </h3>
                      </div>
                      <button
                        type="button"
                        className="rounded-full border border-slate-300 px-4 py-2 text-sm font-semibold text-slate-700"
                        onClick={() =>
                          setProblemDrafts((current) => ({
                            ...current,
                            [contest.id]: [
                              ...draft,
                              {
                                problemId: problems[0]?.id ?? "",
                                code: nextProblemCode(draft.length),
                                position: draft.length + 1,
                              },
                            ],
                          }))
                        }
                      >
                        添加题目
                      </button>
                    </div>
                    <div className="mt-4 grid gap-3">
                      {draft.length === 0 ? (
                        <div className="rounded-2xl border border-dashed border-slate-200 bg-white px-4 py-3 text-sm text-slate-500">
                          当前还没有绑定题目。
                        </div>
                      ) : null}
                      {draft.map((item, index) => (
                        <div
                          key={`${contest.id}-${index}`}
                          className="grid gap-3 rounded-2xl border border-slate-200 bg-white p-3 md:grid-cols-[minmax(0,1fr)_120px_120px_96px]"
                        >
                          <label className="grid gap-1 text-sm text-slate-700">
                            <span>题目</span>
                            <select
                              aria-label={`${contest.slug} problem ${index + 1}`}
                              className="rounded-2xl border border-slate-200 px-4 py-3"
                              value={item.problemId}
                              onChange={(event) =>
                                updateProblemDraft(
                                  contest.id,
                                  index,
                                  "problemId",
                                  event.target.value,
                                  setProblemDrafts,
                                )
                              }
                            >
                              <option value="">选择题目</option>
                              {problems.map((problem) => (
                                <option key={problem.id} value={problem.id}>
                                  {problem.slug} · {problem.title}
                                </option>
                              ))}
                            </select>
                          </label>
                          <label className="grid gap-1 text-sm text-slate-700">
                            <span>题号</span>
                            <input
                              aria-label={`${contest.slug} code ${index + 1}`}
                              className="rounded-2xl border border-slate-200 px-4 py-3"
                              value={item.code}
                              onChange={(event) =>
                                updateProblemDraft(
                                  contest.id,
                                  index,
                                  "code",
                                  event.target.value,
                                  setProblemDrafts,
                                )
                              }
                            />
                          </label>
                          <label className="grid gap-1 text-sm text-slate-700">
                            <span>顺序</span>
                            <input
                              aria-label={`${contest.slug} position ${index + 1}`}
                              className="rounded-2xl border border-slate-200 px-4 py-3"
                              type="number"
                              min={1}
                              value={item.position}
                              onChange={(event) =>
                                updateProblemDraft(
                                  contest.id,
                                  index,
                                  "position",
                                  Number(event.target.value),
                                  setProblemDrafts,
                                )
                              }
                            />
                          </label>
                          <div className="flex items-end">
                            <button
                              type="button"
                              className="w-full rounded-full border border-rose-200 px-4 py-3 text-sm font-semibold text-rose-700"
                              onClick={() =>
                                setProblemDrafts((current) => ({
                                  ...current,
                                  [contest.id]: current[contest.id].filter(
                                    (_, rowIndex) => rowIndex !== index,
                                  ),
                                }))
                              }
                            >
                              删除
                            </button>
                          </div>
                        </div>
                      ))}
                    </div>
                    <div className="mt-4">
                      <button
                        type="button"
                        className="rounded-full bg-slate-950 px-4 py-2 text-sm font-semibold text-white disabled:opacity-60"
                        disabled={
                          replaceProblemsMutation.isPending &&
                          activeContestId === contest.id
                        }
                        onClick={() => {
                          setActiveContestId(contest.id);
                          replaceProblemsMutation.mutate({
                            contestId: contest.id,
                            draft,
                          });
                        }}
                      >
                        {replaceProblemsMutation.isPending &&
                        activeContestId === contest.id
                          ? "保存中..."
                          : "保存题目编排"}
                      </button>
                    </div>
                  </section>
                ) : null}

                <section className="mt-5 rounded-[1.75rem] border border-slate-200 bg-slate-50 p-4">
                  <p className="text-sm uppercase tracking-[0.2em] text-slate-500">
                    snapshots
                  </p>
                  <div className="mt-3 grid gap-3 md:grid-cols-3">
                    {contest.snapshots.length === 0 ? (
                      <div className="rounded-2xl border border-dashed border-slate-200 bg-white px-4 py-3 text-sm text-slate-500">
                        尚未冻结任何快照。
                      </div>
                    ) : (
                      contest.snapshots.map((snapshot) => (
                        <div
                          key={snapshot.id}
                          className="rounded-2xl border border-slate-200 bg-white p-4 text-sm text-slate-700"
                        >
                          <p className="font-semibold text-slate-950">
                            Snapshot #{snapshot.snapshotNo}
                          </p>
                          <p className="mt-1">
                            题目数：{snapshot.problemCount}
                          </p>
                          <p className="mt-1 text-slate-500">
                            {new Date(snapshot.frozenAt).toLocaleString()}
                          </p>
                        </div>
                      ))
                    )}
                  </div>
                </section>
              </article>
            );
          })}
        </div>
      </div>
    </AdminShell>
  );
}

function formToContestInput(
  event: FormEvent<HTMLFormElement>,
): AdminContestUpdateInput {
  const formData = new FormData(event.currentTarget);
  return {
    slug: String(formData.get("slug") ?? "").trim(),
    title: String(formData.get("title") ?? "").trim(),
    description: String(formData.get("description") ?? "").trim(),
    status: String(
      formData.get("status") ?? "UPCOMING",
    ) as AdminContestUpdateInput["status"],
    startsAt: localDateTimeToISOString(String(formData.get("startsAt") ?? "")),
    endsAt: localDateTimeToISOString(String(formData.get("endsAt") ?? "")),
    reason: String(formData.get("reason") ?? "").trim(),
  };
}

function toContestInput(draft: ContestDraftForm): AdminContestCreateInput {
  return {
    slug: draft.slug.trim(),
    title: draft.title.trim(),
    description: draft.description.trim(),
    status: draft.status,
    startsAt: localDateTimeToISOString(draft.startsAt),
    endsAt: localDateTimeToISOString(draft.endsAt),
    reason: draft.reason.trim(),
  };
}

function defaultLocalDateTime(offsetHours: number): string {
  const date = new Date();
  date.setMinutes(0, 0, 0);
  date.setHours(date.getHours() + offsetHours);
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}T${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

function toLocalDateTimeInput(value: string): string {
  const date = new Date(value);
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")}T${String(date.getHours()).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

function localDateTimeToISOString(value: string): string {
  return new Date(value).toISOString();
}

function updateProblemDraft(
  contestId: string,
  index: number,
  key: keyof AdminContestProblemBindingInput,
  value: string | number,
  setProblemDrafts: Dispatch<
    SetStateAction<Record<string, AdminContestProblemBindingInput[]>>
  >,
) {
  setProblemDrafts((current) => ({
    ...current,
    [contestId]: current[contestId].map((item, rowIndex) =>
      rowIndex === index ? { ...item, [key]: value } : item,
    ),
  }));
}

function nextProblemCode(index: number): string {
  return String.fromCharCode("A".charCodeAt(0) + index);
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
