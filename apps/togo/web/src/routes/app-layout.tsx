import { useEffect, useState, type MouseEvent } from "react";
import { Outlet, useNavigate, useRouterState, Link } from "@tanstack/react-router";
import { LayoutGrid, Table2, UserRound, Users as UsersIcon } from "lucide-react";
import {
  AppShell, AppHeader, AppMain, Sidebar, SidebarHeader, SidebarContent, SidebarFooter,
  SidebarGroup, SidebarItem, SidebarTrigger, SidebarExpandedOnly,
  DropdownMenuItem, UserMenu, ImpersonationBanner, ProductMark, WsStatus,
  type WsState,
} from "@fadymondy/nasaq/web";
import { auth, sessionMe, clearSession, type Me } from "../lib/auth";
import { metaResources, adminList, type ResourceMeta } from "../lib/admin";
import { API, APP_NAME, IMPERSONATION_EVENT, getImpersonation, setImpersonation, type Impersonation } from "../lib/api";
import { useLang } from "../lib/i18n";

/** Group a flat resource list by the optional `group` field.
 * Resources with no group fall into the "Resources" default. */
function groupResources(resources: ResourceMeta[]): Map<string, ResourceMeta[]> {
  const map = new Map<string, ResourceMeta[]>();
  for (const r of resources) {
    const g = r.group ?? "Resources";
    if (!map.has(g)) map.set(g, []);
    map.get(g)!.push(r);
  }
  return map;
}

export function AppLayout() {
  const nav = useNavigate();
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const { tx } = useLang();
  const [me, setMe] = useState<Me | null>(null);
  const [resources, setResources] = useState<ResourceMeta[]>([]);
  const [counts, setCounts] = useState<Record<string, number>>({});
  const [live, setLive] = useState<WsState>("connecting");
  const [imp, setImp] = useState<Impersonation | null>(getImpersonation());

  useEffect(() => {
    const on = () => setImp(getImpersonation());
    window.addEventListener(IMPERSONATION_EVENT, on);
    return () => window.removeEventListener(IMPERSONATION_EVENT, on);
  }, []);

  useEffect(() => {
    // Auth is already guaranteed by the route's beforeLoad guard — just read the cached user.
    sessionMe().then(setMe);
    metaResources().then((rs) => {
      setResources(rs);
      // Sidebar count badges — one fetch per resource (best-effort).
      rs.forEach((r) => adminList(r.table).then((rows) => setCounts((c) => ({ ...c, [r.table]: rows.length }))).catch(() => {}));
    });
    const es = new EventSource(`${API}/events`);
    es.onopen = () => setLive("connected");
    es.onerror = () => setLive(es.readyState === EventSource.CLOSED ? "offline" : "reconnecting");
    return () => es.close();
  }, []);

  // SidebarItem is a real <a>: keep the href (open in new tab still works) and route in-app on a plain click.
  const link = (to: string) => ({
    href: to,
    active: pathname === to,
    onClick: (e: MouseEvent) => {
      if (e.metaKey || e.ctrlKey || e.shiftKey || e.button !== 0) return;
      e.preventDefault();
      nav({ to });
    },
  });
  // Stop acting as the user (auth revokes the token and audits it), then return to the
  // admin's own session. A failed stop rejects, so the banner stays and reports it.
  const stopImpersonating = async () => {
    await auth.stopImpersonation();
    setImpersonation(null);
    window.location.assign("/users");
  };
  const signOut = async () => {
    if (imp) await auth.stopImpersonation();
    setImpersonation(null);
    await auth.logout();
    clearSession();
    nav({ to: "/login" });
  };
  const grouped = groupResources(resources);
  const name = me?.email?.split("@")[0] ?? "…";

  const sidebar = (
    <Sidebar>
      <SidebarHeader>
        <Link to="/dashboard" className="flex items-center gap-2 px-2 py-1.5">
          <ProductMark size={24} />
          <SidebarExpandedOnly><span className="truncate font-semibold">{APP_NAME}</span></SidebarExpandedOnly>
        </Link>
      </SidebarHeader>
      <SidebarContent>
        {/* Core nav — always visible */}
        <SidebarGroup>
          <SidebarItem {...link("/dashboard")} icon={<LayoutGrid />}>{tx("Dashboard", "لوحة التحكم")}</SidebarItem>
          <SidebarItem {...link("/admin")} icon={<Table2 />}>{tx("Admin", "الإدارة")}</SidebarItem>
          <SidebarItem {...link("/users")} icon={<UsersIcon />}>{tx("Users", "المستخدمون")}</SidebarItem>
        </SidebarGroup>

        {/* Resource groups — each `group` value becomes its own sidebar section */}
        {Array.from(grouped.entries()).map(([groupName, items]) => (
          <SidebarGroup key={groupName} label={groupName} collapsible>
            {items.map((r) => (
              <SidebarItem
                key={r.table}
                {...link(`/admin/${r.table}`)}
                icon={<Table2 />}
                trailing={counts[r.table] !== undefined ? <span className="text-caption text-muted-foreground tabular-nums">{counts[r.table]}</span> : undefined}
                className="capitalize"
              >
                {r.name || r.table}
              </SidebarItem>
            ))}
          </SidebarGroup>
        ))}
      </SidebarContent>
      <SidebarFooter>
        <UserMenu user={{ name, email: me?.email ?? "" }} onSignOut={signOut} labels={{ theme: tx("Theme", "المظهر"), language: tx("Language", "اللغة"), signOut: tx("Sign out", "تسجيل الخروج") }}>
          <DropdownMenuItem onClick={() => nav({ to: "/profile" })}><UserRound />{tx("Profile", "الملف الشخصي")}</DropdownMenuItem>
        </UserMenu>
      </SidebarFooter>
    </Sidebar>
  );

  return (
    <AppShell sidebar={sidebar}>
      {imp ? (
        <ImpersonationBanner as={{ name: imp.email, email: imp.email }} startedAt={imp.startedAt} onExit={stopImpersonating} />
      ) : null}
      <AppHeader>
        <SidebarTrigger />
        <WsStatus state={live} showLatency={false} />
      </AppHeader>
      <AppMain><Outlet /></AppMain>
    </AppShell>
  );
}
