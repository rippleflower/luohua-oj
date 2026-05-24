type SubmissionStatusBadgeProps = {
  status: string;
};

const statusClassNames: Record<string, string> = {
  PENDING: "bg-slate-100 text-slate-700 ring-slate-200",
  RUNNING: "bg-sky-50 text-sky-700 ring-sky-200",
  ACCEPTED: "bg-emerald-50 text-emerald-700 ring-emerald-200",
  WRONG_ANSWER: "bg-amber-50 text-amber-700 ring-amber-200",
  TIME_LIMIT_EXCEEDED: "bg-rose-50 text-rose-700 ring-rose-200",
  MEMORY_LIMIT_EXCEEDED: "bg-fuchsia-50 text-fuchsia-700 ring-fuchsia-200",
  COMPILE_ERROR: "bg-orange-50 text-orange-700 ring-orange-200",
  RUNTIME_ERROR: "bg-rose-50 text-rose-700 ring-rose-200",
  SYSTEM_ERROR: "bg-rose-50 text-rose-700 ring-rose-200",
};

const labelMap: Record<string, string> = {
  PENDING: "等待中",
  RUNNING: "运行中",
  ACCEPTED: "已通过",
  WRONG_ANSWER: "答案错误",
  TIME_LIMIT_EXCEEDED: "超出时间限制",
  MEMORY_LIMIT_EXCEEDED: "超出内存限制",
  COMPILE_ERROR: "编译错误",
  RUNTIME_ERROR: "运行错误",
  SYSTEM_ERROR: "系统错误",
};

export function formatSubmissionStatus(status: string) {
  return labelMap[status] ?? status;
}

export function SubmissionStatusBadge({ status }: SubmissionStatusBadgeProps) {
  const className = statusClassNames[status] ?? "bg-slate-100 text-slate-700 ring-slate-200";

  return (
    <span className={`inline-flex rounded-full px-2 py-1 text-xs font-semibold ring-1 ring-inset ${className}`}>
      {formatSubmissionStatus(status)}
    </span>
  );
}
