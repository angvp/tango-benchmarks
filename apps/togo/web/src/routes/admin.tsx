import { useEffect, useState } from "react";
import { Link } from "@tanstack/react-router";
import { Table2 } from "lucide-react";
import { PageHeader, Card, EmptyState, Skeleton } from "@fadymondy/nasaq/web";
import { metaResources } from "../lib/admin";
import { useLang } from "../lib/i18n";

export function AdminHome() {
  const { tx } = useLang();
  const [list, setList] = useState<{ name: string; table: string }[]>([]);
  const [loading, setLoading] = useState(true);
  useEffect(() => { metaResources().then(setList).finally(() => setLoading(false)); }, []);
  return (
    <div className="flex flex-col gap-6">
      <PageHeader title={tx("Admin", "الإدارة")} description={tx(`Manage your resources · ${list.length}`, `إدارة الموارد · ${list.length}`)} />
      {loading ? (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">{[0, 1, 2].map((i) => <Skeleton key={i} className="h-20" />)}</div>
      ) : list.length === 0 ? (
        <EmptyState
          icon={Table2}
          title={tx("No resources yet", "لا توجد موارد بعد")}
          description={tx("Run `togo make:resource Post title:string` and they'll appear here.", "نفّذ `togo make:resource Post title:string` وستظهر هنا.")}
        />
      ) : (
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          {list.map((r) => (
            <Link key={r.table} to="/admin/$resource" params={{ resource: r.table }} className="block">
              <Card className="flex-row items-center gap-3 p-4 transition-colors hover:border-primary/50">
                <span className="flex size-9 items-center justify-center bg-primary/15 text-primary"><Table2 className="size-4" /></span>
                <span>
                  <span className="block font-medium capitalize">{r.name || r.table}</span>
                  <span className="block text-caption text-muted-foreground" dir="ltr">/api/{r.table}</span>
                </span>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
