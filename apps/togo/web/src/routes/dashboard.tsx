import { useEffect, useState } from "react";
import { useNavigate } from "@tanstack/react-router";
import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from "recharts";
import { Table2, ShieldCheck, KeyRound, UserRound } from "lucide-react";
import {
  PageHeader, StatCard, StatGrid, Card, CardHeader, CardTitle, CardDescription, CardContent,
  ChartContainer, ChartTooltip, ChartTooltipContent, EmptyState, useChartAxis, type ChartConfig,
} from "@fadymondy/nasaq/web";
import { sessionMe, type Me } from "../lib/auth";
import { metaResources, adminList, type ResourceMeta } from "../lib/admin";
import { useLang } from "../lib/i18n";

const labelOf = (s: string) => s.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
type Point = { label: string; value: number };

export function Dashboard() {
  const nav = useNavigate();
  const { ar, tx } = useLang();
  const { xAxis, yAxis } = useChartAxis();
  const [me, setMe] = useState<Me | null>(null);
  const [counts, setCounts] = useState<{ meta: ResourceMeta; count: number }[] | null>(null);
  const [trend, setTrend] = useState<Point[]>([]);

  useEffect(() => { sessionMe().then(setMe); }, []);
  useEffect(() => {
    metaResources().then(async (ms) => {
      const all = await Promise.all(ms.map(async (m) => ({ meta: m, rows: await adminList(m.table).catch(() => []) })));
      setCounts(all.map(({ meta, rows }) => ({ meta, count: rows.length })));
      // Records-created-per-day over the last 7 days, across every resource (uses created_at).
      const days: { key: string; label: string }[] = [];
      for (let i = 6; i >= 0; i--) {
        const d = new Date(); d.setDate(d.getDate() - i);
        days.push({ key: d.toISOString().slice(0, 10), label: d.toLocaleDateString(ar ? "ar" : "en", { weekday: "short" }) });
      }
      const bucket: Record<string, number> = Object.fromEntries(days.map((d) => [d.key, 0]));
      for (const { rows } of all) for (const r of rows) {
        const ts = r.created_at ?? r.createdAt;
        if (ts) { const k = new Date(ts).toISOString().slice(0, 10); if (k in bucket) bucket[k]++; }
      }
      setTrend(days.map((d) => ({ label: d.label, value: bucket[d.key] })));
    });
  }, [ar]);

  const byResource: Point[] = (counts ?? []).map(({ meta, count }) => ({ label: labelOf(meta.name || meta.table), value: count }));
  const total = byResource.reduce((s, p) => s + p.value, 0);
  const hasTrend = trend.some((p) => p.value > 0);
  const config: ChartConfig = { value: { label: tx("Records", "السجلات") } };
  const open = (table: string) => nav({ to: "/admin/$resource", params: { resource: table } });

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title={tx("Dashboard", "لوحة التحكم")}
        description={me ? tx(`Welcome back, ${me.email}`, `مرحبًا بعودتك، ${me.email}`) : undefined}
      />

      <StatGrid>
        <StatCard icon={<UserRound />} label={tx("Account", "الحساب")} value={me?.email ?? "…"} loading={!me} />
        <StatCard icon={<ShieldCheck />} label={tx("Roles", "الأدوار")} value={(me?.roles ?? ["user"]).join(", ")} loading={!me} />
        <StatCard icon={<KeyRound />} label={tx("Permissions", "الصلاحيات")} value={(me?.permissions ?? []).length} loading={!me} />
      </StatGrid>

      {/* One stat card per registered resource (record count); click through to its table. */}
      {counts && counts.length > 0 && (
        <StatGrid>
          {counts.map(({ meta, count }) => (
            <StatCard
              key={meta.table}
              icon={<Table2 />}
              label={labelOf(meta.name || meta.table)}
              value={count}
              role="link"
              tabIndex={0}
              className="cursor-pointer transition-colors hover:border-primary/50"
              onClick={() => open(meta.table)}
              onKeyDown={(e) => { if (e.key === "Enter") open(meta.table); }}
            />
          ))}
        </StatGrid>
      )}

      {counts && counts.length === 0 ? (
        <EmptyState
          title={tx("No resources yet", "لا توجد موارد بعد")}
          description={tx("Run `togo make:resource Post title:string` and it shows up here.", "نفّذ `togo make:resource Post title:string` وسيظهر هنا.")}
        />
      ) : (
        <div className="grid gap-4 lg:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>{tx("Records by resource", "السجلات حسب المورد")}</CardTitle>
              <CardDescription>{total} {tx("total", "إجمالي")}</CardDescription>
            </CardHeader>
            <CardContent>
              <ChartContainer config={config} label={tx("Records by resource", "السجلات حسب المورد")} className="aspect-auto h-56">
                <BarChart data={byResource}>
                  <CartesianGrid vertical={false} />
                  <XAxis dataKey="label" tickLine={false} axisLine={false} {...xAxis} />
                  <YAxis tickLine={false} axisLine={false} width={32} allowDecimals={false} {...yAxis} />
                  <ChartTooltip content={<ChartTooltipContent config={config} />} />
                  <Bar dataKey="value" fill="var(--color-value)" />
                </BarChart>
              </ChartContainer>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>{tx("New records (7 days)", "سجلات جديدة (7 أيام)")}</CardTitle>
              <CardDescription>{tx("Across every resource, by created_at", "عبر كل الموارد، حسب created_at")}</CardDescription>
            </CardHeader>
            <CardContent>
              {hasTrend ? (
                <ChartContainer config={config} label={tx("New records over the last 7 days", "السجلات الجديدة خلال 7 أيام")} className="aspect-auto h-56">
                  <BarChart data={trend}>
                    <CartesianGrid vertical={false} />
                    <XAxis dataKey="label" tickLine={false} axisLine={false} {...xAxis} />
                    <YAxis tickLine={false} axisLine={false} width={32} allowDecimals={false} {...yAxis} />
                    <ChartTooltip content={<ChartTooltipContent config={config} />} />
                    <Bar dataKey="value" fill="var(--color-value)" />
                  </BarChart>
                </ChartContainer>
              ) : (
                <p className="py-16 text-center text-body-sm text-muted-foreground">{tx("No timestamped records yet", "لا توجد بيانات بعد")}</p>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
