import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";

import { AppShell } from "../../components/layout/app-shell";
import { login } from "../../features/auth/api";
import { useLocale } from "../../lib/locale";

export function LoginRoute() {
  const { locale } = useLocale();
  const queryClient = useQueryClient();
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    try {
      await login(identifier, password);
      await queryClient.invalidateQueries({ queryKey: ["auth"] });
      window.location.href = "/me";
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "login failed");
    }
  }

  return (
    <AppShell title={locale === "zh" ? "登录" : "Sign In"} subtitle="account">
      <form className="mx-auto grid max-w-xl gap-4 rounded-3xl border border-slate-200 bg-white/85 p-6" onSubmit={handleSubmit}>
        <label className="grid gap-2 text-sm">
          <span>{locale === "zh" ? "邮箱或用户名" : "Email or username"}</span>
          <input className="rounded-2xl border border-slate-200 px-4 py-3" value={identifier} onChange={(event) => setIdentifier(event.target.value)} />
        </label>
        <label className="grid gap-2 text-sm">
          <span>{locale === "zh" ? "密码" : "Password"}</span>
          <input className="rounded-2xl border border-slate-200 px-4 py-3" type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
        </label>
        {error ? <p className="rounded-2xl bg-rose-50 px-4 py-3 text-sm text-rose-700">{error}</p> : null}
        <button className="rounded-full bg-slate-950 px-4 py-3 text-sm font-semibold text-white" type="submit">
          {locale === "zh" ? "登录并进入个人中心" : "Sign in"}
        </button>
        <p className="text-sm text-slate-500">
          {locale === "zh" ? "还没有账号？" : "Need an account? "}
          <a className="font-semibold text-slate-900" href="/register">
            {locale === "zh" ? "去注册" : "Register"}
          </a>
        </p>
      </form>
    </AppShell>
  );
}
