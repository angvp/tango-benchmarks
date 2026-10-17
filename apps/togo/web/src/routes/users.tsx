// Users admin — the out-of-the-box account-management page built on Nasaq's
// AdminUsers. Wired to the auth plugin's admin API (/api/auth/admin/*). Clicking a
// row opens a dialog for the actions AdminUsers has no slot for: magic link + delete.
import { useCallback, useEffect, useMemo, useState } from "react";
import { Link2, Trash2 } from "lucide-react";
import {
  PageHeader, AdminUsers, ErrorState, Badge, Button, ConfirmButton, CopyButton, Input,
  Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription, DialogFooter, toast,
  type ManagedUser, type ManagedRole,
} from "@fadymondy/nasaq/web";
import { adminUsers, AdminError, type AdminUser, type AdminLinkResult } from "../lib/admin-users";
import { sessionMe } from "../lib/auth";
import { useLang } from "../lib/i18n";

const errorOf = (e: unknown) => ({ error: e instanceof Error ? e.message : String(e) });

function toManaged(u: AdminUser): ManagedUser {
  return {
    id: String(u.id),
    name: u.email.split("@")[0],
    email: u.email,
    roles: u.roles?.length ? u.roles : ["user"],
    status: "active",
    verified: true,
    createdAt: u.created_at ?? new Date(0).toISOString(),
  };
}

export function Users() {
  const { tx } = useLang();
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [available, setAvailable] = useState(true);
  const [meEmail, setMeEmail] = useState<string>();
  const [open, setOpen] = useState<ManagedUser | null>(null);
  const [link, setLink] = useState<{ title: string; url: string } | null>(null);

  const reload = useCallback(async () => {
    setLoading(true);
    try {
      setUsers(await adminUsers.list());
      setError(null);
    } catch (e) {
      if (e instanceof AdminError && (e.status === 404 || e.status === 501)) setAvailable(false);
      else setError(e instanceof Error ? e.message : String(e));
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { reload(); sessionMe().then((m) => setMeEmail(m?.email)); }, [reload]);

  const managed = useMemo(() => users.map(toManaged), [users]);
  const roles = useMemo<ManagedRole[]>(() => {
    const ids = new Set(["admin", "user", ...users.flatMap((u) => u.roles ?? [])]);
    return [...ids].map((id) => ({ id, label: id.charAt(0).toUpperCase() + id.slice(1) }));
  }, [users]);
  const currentUserId = managed.find((u) => u.email === meEmail)?.id;

  // A link the backend could not email is shown so the admin can pass it on.
  function deliver(r: AdminLinkResult, title: string) {
    if (r.emailed) toast.success(tx("Email sent", "تم إرسال البريد"));
    else if (r.link) setLink({ title, url: r.link });
  }

  if (!available) {
    return (
      <div className="flex flex-col gap-6">
        <PageHeader title={tx("Users", "المستخدمون")} />
        <ErrorState
          title={tx("Admin API unavailable", "واجهة الإدارة غير متاحة")}
          description={tx("This app's auth plugin has no admin API. Upgrade it: `go get github.com/togo-framework/auth@v0.10.0`.", "مكوّن المصادقة لا يوفّر واجهة الإدارة. حدّثه: go get github.com/togo-framework/auth@v0.10.0")}
        />
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={tx("Users", "المستخدمون")}
        description={tx("Accounts managed by the ToGO auth plugin", "حسابات يديرها مكوّن ToGO للمصادقة")}
      />

      <AdminUsers
        users={managed}
        roles={roles}
        currentUserId={currentUserId}
        loading={loading}
        error={error}
        onRetry={reload}
        onAddUser={async (v) => {
          try {
            const { user } = await adminUsers.create({ email: v.email, roles: v.roles });
            toast.success(tx("User created", "تم إنشاء المستخدم"));
            if (v.sendInvite) deliver(await adminUsers.magicLink(String(user.id)), tx("Invite link", "رابط الدعوة"));
            await reload();
          } catch (e) { return errorOf(e); }
        }}
        onResetPassword={async (u) => {
          try { deliver(await adminUsers.resetPassword(u.id), tx("Password reset link", "رابط إعادة تعيين كلمة المرور")); }
          catch (e) { return errorOf(e); }
        }}
        onImpersonate={async (u) => {
          try {
            await adminUsers.impersonate(u.id);
            // Full reload so every cached /me and resource read re-runs as the user.
            window.location.assign("/dashboard");
          } catch (e) { return errorOf(e); }
        }}
        onUpdateRoles={async (u, next) => {
          try {
            await adminUsers.update(u.id, { roles: next });
            toast.success(tx("Roles updated", "تم تحديث الأدوار"));
            await reload();
          } catch (e) { return errorOf(e); }
        }}
        onOpenUser={setOpen}
      />

      <Dialog open={!!open} onOpenChange={(o) => { if (!o) setOpen(null); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{open?.name}</DialogTitle>
            <DialogDescription dir="ltr">{open?.email}</DialogDescription>
          </DialogHeader>
          <div className="flex flex-wrap gap-1.5">
            {open?.roles.map((r) => <Badge key={r} variant="outline">{r}</Badge>)}
          </div>
          <DialogFooter>
            <Button
              variant="secondary"
              onClick={async () => {
                if (!open) return;
                try { deliver(await adminUsers.magicLink(open.id), tx("Magic link", "رابط الدخول السحري")); setOpen(null); }
                catch (e) { toast.error(errorOf(e).error); }
              }}
            >
              <Link2 /> {tx("Send magic link", "إرسال رابط دخول")}
            </Button>
            {open && open.id !== currentUserId && (
              <ConfirmButton
                title={tx(`Delete ${open.email}?`, `حذف ${open.email}؟`)}
                description={tx("The account and its sessions are removed. This cannot be undone.", "سيُحذف الحساب وجلساته. لا يمكن التراجع.")}
                confirmLabel={tx("Delete", "حذف")}
                onConfirm={async () => {
                  try {
                    await adminUsers.remove(open.id);
                    toast.success(tx("User deleted", "تم حذف المستخدم"));
                    setOpen(null);
                    await reload();
                  } catch (e) { toast.error(errorOf(e).error); }
                }}
              >
                <Trash2 /> {tx("Delete", "حذف")}
              </ConfirmButton>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={!!link} onOpenChange={(o) => { if (!o) setLink(null); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{link?.title}</DialogTitle>
            <DialogDescription>{tx("Mail is not configured, so share this link yourself.", "البريد غير مُعدّ، شارك هذا الرابط بنفسك.")}</DialogDescription>
          </DialogHeader>
          <div className="flex gap-2">
            <Input readOnly value={link?.url ?? ""} dir="ltr" onFocus={(e) => e.currentTarget.select()} />
            <CopyButton value={link?.url ?? ""} />
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
