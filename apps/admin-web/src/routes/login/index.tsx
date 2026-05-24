import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";

import { login } from "../../features/auth/api";

export function AdminLoginRoute() {
  const queryClient = useQueryClient();
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    try {
      await login(identifier, password);
      await queryClient.invalidateQueries({ queryKey: ["admin"] });
      window.location.href = "/";
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "login failed");
    }
  }

  return (
    <main className="relative min-h-screen px-3 py-6 sm:px-4 lg:px-5">
      <div className="grid w-full gap-5 lg:grid-cols-[minmax(0,1.05fr)_minmax(0,0.95fr)]">
        <section className="rounded-[2rem] border border-orange-950/10 bg-white/70 p-8 shadow-[0_20px_80px_rgba(120,53,15,0.08)] backdrop-blur-sm sm:p-10">
          <p className="font-mono text-[11px] uppercase tracking-[0.34em] text-slate-500">luooj admin</p>
          <h1 className="mt-4 font-['Newsreader'] text-4xl leading-[0.95] text-slate-950 sm:text-6xl">管理后台登录</h1>
          <p className="mt-6 max-w-md text-sm leading-6 text-slate-600">
            面向题库、比赛与判题运营的统一控制台。登录后可继续进行题目发布、比赛冻结和审计追踪。
          </p>
        </section>

        <form className="grid content-start gap-4 rounded-[2rem] border border-orange-950/10 bg-white/90 p-7 shadow-[0_20px_80px_rgba(120,53,15,0.08)] sm:p-8" onSubmit={handleSubmit}>
          <div>
            <p className="font-mono text-[11px] uppercase tracking-[0.3em] text-slate-500">access</p>
            <h2 className="mt-3 text-3xl font-semibold text-slate-950">进入后台</h2>
          </div>

          <label className="grid gap-2 text-sm text-slate-700">
            <span className="font-medium">账号</span>
            <input
              className="cursor-text rounded-xl border-2 border-slate-300 bg-slate-50 px-4 py-3.5 text-base text-slate-900 shadow-[inset_0_1px_2px_rgba(15,23,42,0.06)] outline-none transition placeholder:text-slate-400 focus:border-orange-400 focus:bg-white focus:ring-2 focus:ring-orange-100"
              placeholder="邮箱或用户名"
              autoComplete="username"
              value={identifier}
              onChange={(event) => setIdentifier(event.target.value)}
            />
          </label>

          <label className="grid gap-2 text-sm text-slate-700">
            <span className="font-medium">密码</span>
            <input
              className="cursor-text rounded-xl border-2 border-slate-300 bg-slate-50 px-4 py-3.5 text-base text-slate-900 shadow-[inset_0_1px_2px_rgba(15,23,42,0.06)] outline-none transition placeholder:text-slate-400 focus:border-orange-400 focus:bg-white focus:ring-2 focus:ring-orange-100"
              placeholder="输入密码"
              type="password"
              autoComplete="current-password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
            />
          </label>

          {error ? <p className="rounded-2xl bg-rose-50 px-4 py-3 text-sm text-rose-700">{error}</p> : null}

          <button className="mt-1 rounded-full bg-slate-950 px-4 py-3 text-sm font-semibold text-white transition hover:bg-slate-800" type="submit">
            进入后台
          </button>
        </form>
      </div>
    </main>
  );
}
