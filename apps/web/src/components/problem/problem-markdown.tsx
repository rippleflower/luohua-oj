import { memo } from "react";
import ReactMarkdown from "react-markdown";

export const ProblemMarkdown = memo(function ProblemMarkdown({
  content,
}: {
  content: string;
}) {
  return (
    <div className="problem-markdown text-sm leading-7 text-slate-700">
      <ReactMarkdown
        components={{
          h1: ({ children }) => <h3 className="mt-5 text-xl font-semibold text-slate-950 first:mt-0">{children}</h3>,
          h2: ({ children }) => <h4 className="mt-5 text-lg font-semibold text-slate-950 first:mt-0">{children}</h4>,
          h3: ({ children }) => <h5 className="mt-4 text-base font-semibold text-slate-950 first:mt-0">{children}</h5>,
          p: ({ children }) => <p className="mt-3 first:mt-0">{children}</p>,
          ul: ({ children }) => <ul className="mt-3 list-disc space-y-1 pl-5">{children}</ul>,
          ol: ({ children }) => <ol className="mt-3 list-decimal space-y-1 pl-5">{children}</ol>,
          li: ({ children }) => <li>{children}</li>,
          strong: ({ children }) => <strong className="font-semibold text-slate-900">{children}</strong>,
          em: ({ children }) => <em className="italic">{children}</em>,
          code: ({ children }) => (
            <code className="rounded bg-slate-100 px-1.5 py-0.5 font-mono text-[0.9em] text-slate-900">
              {children}
            </code>
          ),
          pre: ({ children }) => <pre className="mt-3">{children}</pre>,
          blockquote: ({ children }) => (
            <blockquote className="mt-3 rounded-r-2xl border-l-4 border-slate-300 bg-slate-50 px-4 py-3 text-slate-600">
              {children}
            </blockquote>
          ),
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
});
