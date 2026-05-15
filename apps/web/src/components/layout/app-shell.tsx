import type { ReactNode } from "react";

type AppShellProps = {
  title: string;
  subtitle: string;
  action?: ReactNode;
  children: ReactNode;
};

export function AppShell({ title, subtitle, action, children }: AppShellProps) {
  return (
    <main className="min-h-screen bg-slate-50 text-slate-950">
      <section className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-5">
          <div>
            <p className="text-sm font-medium text-slate-500">{subtitle}</p>
            <h1 className="text-2xl font-semibold">{title}</h1>
          </div>
          {action}
        </div>
      </section>
      <section className="mx-auto max-w-6xl px-6 py-6">{children}</section>
    </main>
  );
}
