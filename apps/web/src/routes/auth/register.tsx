import { useState } from "react";
import { useQueryClient } from "@tanstack/react-query";

import { AppShell } from "../../components/layout/app-shell";
import { register } from "../../features/auth/api";
import { useLocale } from "../../lib/locale";

export function RegisterRoute() {
  const { locale } = useLocale();
  const queryClient = useQueryClient();
  const [form, setForm] = useState({
    email: "",
    username: "",
    displayName: "",
    password: "",
    confirmPassword: "",
  });
  const [error, setError] = useState("");

  async function handleSubmit(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setError("");
    try {
      await register(form);
      await queryClient.invalidateQueries({ queryKey: ["auth"] });
      window.location.href = "/me";
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : "register failed");
    }
  }

  return (
    <AppShell title={locale === "zh" ? "注册账号" : "Create Account"} subtitle="account">
      <form className="mx-auto grid max-w-xl gap-4 rounded-3xl border border-slate-200 bg-white/85 p-6" onSubmit={handleSubmit}>
        {[
          ["email", locale === "zh" ? "邮箱" : "Email", "email"],
          ["username", locale === "zh" ? "用户名" : "Username", "text"],
          ["displayName", locale === "zh" ? "显示名称" : "Display name", "text"],
          ["password", locale === "zh" ? "密码" : "Password", "password"],
          ["confirmPassword", locale === "zh" ? "确认密码" : "Confirm password", "password"],
        ].map(([key, label, type]) => (
          <label key={key} className="grid gap-2 text-sm">
            <span>{label}</span>
            <input
              className="rounded-2xl border border-slate-200 px-4 py-3"
              type={type}
              value={form[key as keyof typeof form]}
              onChange={(event) => setForm((current) => ({ ...current, [key]: event.target.value }))}
            />
          </label>
        ))}
        {error ? <p className="rounded-2xl bg-rose-50 px-4 py-3 text-sm text-rose-700">{error}</p> : null}
        <button className="rounded-full bg-slate-950 px-4 py-3 text-sm font-semibold text-white" type="submit">
          {locale === "zh" ? "注册并登录" : "Register"}
        </button>
      </form>
    </AppShell>
  );
}
