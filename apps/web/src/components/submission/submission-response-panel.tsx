import type {
  SubmissionDetail,
  SubmissionSummary,
} from "../../features/submissions/schema";
import { useLocale } from "../../lib/locale";
import { SubmissionStatusBadge } from "./submission-status-badge";

type SubmissionResponsePanelProps = {
  data?: SubmissionSummary | SubmissionDetail;
  errorMessage?: string;
};

export function SubmissionResponsePanel({
  data,
  errorMessage,
}: SubmissionResponsePanelProps) {
  const { locale } = useLocale();
  const resultCount =
    data && "results" in data && Array.isArray(data.results)
      ? data.results.length
      : undefined;
  return (
    <section className="rounded-lg border border-slate-200 bg-white p-6">
      <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500">
        {locale === "zh" ? "响应" : "Response"}
      </h2>
      {errorMessage ? (
        <p className="mt-4 text-sm text-rose-600">{errorMessage}</p>
      ) : null}
      {data ? (
        <dl className="mt-4 space-y-3 text-sm">
          <div>
            <dt className="text-slate-500">
              {locale === "zh" ? "提交 ID" : "Submission ID"}
            </dt>
            <dd className="break-all font-medium text-slate-950">{data.id}</dd>
          </div>
          <div>
            <dt className="text-slate-500">
              {locale === "zh" ? "状态" : "Status"}
            </dt>
            <dd className="font-medium text-slate-950">
              <SubmissionStatusBadge status={data.status} />
            </dd>
          </div>
          <div>
            <dt className="text-slate-500">
              {locale === "zh" ? "源码对象 Key" : "Source Object Key"}
            </dt>
            <dd className="break-all font-medium text-slate-950">
              {data.sourceObjectKey}
            </dd>
          </div>
          <div>
            <dt className="text-slate-500">
              {locale === "zh" ? "结果数" : "Result Count"}
            </dt>
            <dd className="font-medium text-slate-950">{resultCount ?? "—"}</dd>
          </div>
        </dl>
      ) : (
        <p className="mt-4 text-sm text-slate-500">
          {locale === "zh" ? "还没有提交记录。" : "No submission yet."}
        </p>
      )}
    </section>
  );
}
