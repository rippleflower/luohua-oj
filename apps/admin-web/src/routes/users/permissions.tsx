import { useEffect, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import { permissionKeys } from "@oj/shared";

import { AdminShell } from "../../components/layout/admin-shell";
import { getUserPermissions, updateUserPermissions } from "../../features/admin/api";

export function UserPermissionsRoute({ userId }: { userId: string }) {
  const queryClient = useQueryClient();
  const { data = [] } = useQuery({
    queryKey: ["admin", "users", userId, "permissions"],
    queryFn: () => getUserPermissions(userId),
    retry: false,
  });
  const [selected, setSelected] = useState<string[]>(data);
  const [message, setMessage] = useState("");

  useEffect(() => {
    setSelected(data);
  }, [data]);

  async function handleSave() {
    await updateUserPermissions(userId, selected as typeof data, "permissions updated in admin console");
    await queryClient.invalidateQueries({ queryKey: ["admin", "users", userId, "permissions"] });
    await queryClient.invalidateQueries({ queryKey: ["admin", "users", userId] });
    setMessage("Permissions updated.");
  }

  return (
    <AdminShell title="权限配置">
      <div className="grid gap-4 rounded-3xl border border-slate-200 bg-white p-5">
        <p className="text-sm text-slate-500">前端菜单和后端接口都按这一组 permission key 裁剪。</p>
        <div className="grid gap-2 md:grid-cols-2">
          {permissionKeys.map((permission) => {
            const checked = selected.includes(permission);
            return (
              <label key={permission} className="flex items-center gap-3 rounded-2xl border border-slate-200 p-3 text-sm">
                <input
                  checked={checked}
                  type="checkbox"
                  onChange={(event) =>
                    setSelected((current) =>
                      event.target.checked ? [...new Set([...current, permission])] : current.filter((item) => item !== permission),
                    )
                  }
                />
                <span>{permission}</span>
              </label>
            );
          })}
        </div>
        <button className="w-fit rounded-full bg-slate-950 px-4 py-3 text-sm font-semibold text-white" type="button" onClick={handleSave}>
          保存权限
        </button>
        {message ? <p className="text-sm text-emerald-700">{message}</p> : null}
      </div>
    </AdminShell>
  );
}
