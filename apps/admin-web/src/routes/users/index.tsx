import { useQuery } from "@tanstack/react-query";

import { AdminShell } from "../../components/layout/admin-shell";
import { listUsers } from "../../features/admin/api";

export function UsersRoute() {
  const { data = [] } = useQuery({
    queryKey: ["admin", "users"],
    queryFn: listUsers,
    retry: false,
  });

  return (
    <AdminShell title="用户与权限">
      <div className="grid gap-4">
        <div className="grid gap-3 md:grid-cols-3">
          <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">筛选栏：用户状态、角色、用户名、邮箱。</div>
          <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">批量动作：停用、恢复、导出、会话清理。</div>
          <div className="rounded-3xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">高风险操作必须填写原因，后端会写审计日志。</div>
        </div>
        <div className="overflow-hidden rounded-3xl border border-slate-200 bg-white">
          <table className="min-w-full border-collapse text-sm">
            <thead className="bg-slate-50 text-left text-slate-500">
              <tr>
                {["用户", "邮箱", "角色", "状态", "更新时间", "动作"].map((label) => (
                  <th key={label} className="px-4 py-3 font-medium">{label}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {data.map((user) => (
                <tr key={user.id} className="border-t border-slate-200">
                  <td className="px-4 py-3">
                    <p className="font-semibold text-slate-950">{user.displayName}</p>
                    <p className="text-xs text-slate-500">{user.username}</p>
                  </td>
                  <td className="px-4 py-3">{user.email}</td>
                  <td className="px-4 py-3">{user.role}</td>
                  <td className="px-4 py-3">{user.status}</td>
                  <td className="px-4 py-3">{new Date(user.updatedAt).toLocaleString()}</td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <a className="rounded-full border border-slate-200 px-3 py-1.5 text-xs font-semibold" href={`/users/${user.id}`}>详情</a>
                      <a className="rounded-full border border-slate-200 px-3 py-1.5 text-xs font-semibold" href={`/users/${user.id}/permissions`}>权限</a>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </AdminShell>
  );
}
