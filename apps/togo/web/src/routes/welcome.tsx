import { Link } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { ArrowLeft, ArrowRight, BookOpen, Blocks, Boxes, Database, FileText, GitBranch, LayoutGrid, Rocket, Terminal as TerminalIcon } from "lucide-react";
import {
  AppMockupHero, CtaBanner, FeatureGrid, HowItWorks, LocaleSwitcher, ProductLogo, Status, Terminal, TextFlip, ThemeSwitcher,
  buttonVariants, type FeatureGridItem, type TerminalLine,
} from "@fadymondy/nasaq/web";
import { API, APP_NAME } from "../lib/api";
import { sessionMe, type Me } from "../lib/auth";
import { useLang } from "../lib/i18n";

const DOCS = "https://to-go.dev/en/docs";

// What `togo serve` prints on a fresh app: the hero's mockup.
function bootLines(version: string): TerminalLine[] {
  return [
    { kind: "command", text: "togo serve" },
    { kind: "info", text: `ToGO ${version} · ${APP_NAME}` },
    { text: "  REST     /api        OpenAPI 3.1 at /docs" },
    { text: "  GraphQL  /graphql    playground at /graphql/play" },
    { text: "  Web      /           React + Nasaq" },
    { kind: "success", text: "ready on http://localhost:8080" },
    { kind: "command", text: "togo make:resource Post title:string body:text" },
    { kind: "success", text: "created migration, sqlc queries, REST + GraphQL, admin page" },
  ];
}

export function Welcome() {
  const { ar, tx } = useLang();
  const Arrow = ar ? ArrowLeft : ArrowRight;

  const [health, setHealth] = useState<{ status?: string; togo?: string } | null>(null);
  const [me, setMe] = useState<Me | null | undefined>(undefined);

  useEffect(() => {
    fetch(`${API}/api/health`).then((r) => r.json()).then(setHealth).catch(() => setHealth(null));
    sessionMe().then(setMe).catch(() => setMe(null));
  }, []);

  const online = health?.status === "ok";
  const version = health?.togo ?? "dev";

  const primary = buttonVariants({ variant: "primary", size: "lg" });
  const secondary = buttonVariants({ variant: "secondary", size: "lg" });

  // Where a fresh ToGO app sends you next.
  const next: FeatureGridItem[] = [
    { icon: <LayoutGrid />, title: tx("Dashboard", "لوحة التحكم"), description: tx("Your admin panel: account, roles and every resource.", "لوحة الإدارة: الحساب والأدوار وكل الموارد."), href: "/dashboard" },
    { icon: <BookOpen />, title: tx("REST API", "واجهة REST"), description: tx("Typed OpenAPI 3.1 docs for every resource.", "توثيق OpenAPI 3.1 لكل مورد."), href: `${API}/docs` },
    { icon: <Boxes />, title: "GraphQL", description: tx("Explore the schema in the playground.", "استكشف المخطط في ساحة التجربة."), href: `${API}/graphql/play` },
    { icon: <FileText />, title: tx("Documentation", "التوثيق"), description: tx("Guides, generators and the ToGO CLI.", "أدلة ومولّدات وأداة ToGO."), href: DOCS },
    { icon: <Blocks />, title: tx("Plugins", "الإضافات"), description: tx("Add auth, queues, storage or AI in one command.", "أضف المصادقة والطوابير والتخزين والذكاء الاصطناعي بأمر واحد."), href: "https://to-go.dev/en/plugins" },
    { icon: <GitBranch />, title: "GitHub", description: tx("Source, issues and every framework repo.", "المصدر والبلاغات وكل مستودعات الإطار."), href: "https://github.com/togo-framework" },
  ];

  return (
    <div className="min-h-dvh bg-background text-foreground">
      <header className="sticky top-0 z-20 border-b border-border bg-background/80 backdrop-blur">
        <div className="mx-auto flex h-14 w-full max-w-6xl items-center gap-4 px-4 sm:px-8">
          <Link to="/" aria-label={APP_NAME} className="flex items-center gap-2">
            <ProductLogo size={22} />
          </Link>
          <nav className="hidden items-center gap-5 text-body-sm text-muted-foreground md:flex">
            <a href={DOCS} target="_blank" rel="noreferrer" className="hover:text-foreground">{tx("Docs", "التوثيق")}</a>
            <a href={`${API}/docs`} className="hover:text-foreground">API</a>
            <a href="https://to-go.dev/en/plugins" target="_blank" rel="noreferrer" className="hover:text-foreground">{tx("Plugins", "الإضافات")}</a>
          </nav>
          <div className="ms-auto flex items-center gap-2">
            <LocaleSwitcher />
            <ThemeSwitcher />
            {me ? (
              <Link to="/dashboard" className={buttonVariants({ variant: "primary", size: "sm" })}>{tx("Dashboard", "لوحة التحكم")}</Link>
            ) : (
              <Link to="/login" className={buttonVariants({ variant: "secondary", size: "sm" })}>{tx("Sign in", "تسجيل الدخول")}</Link>
            )}
          </div>
        </div>
      </header>

      <main className="mx-auto flex w-full max-w-6xl flex-col gap-20 px-4 pb-20 sm:px-8">
        <AppMockupHero
          eyebrow={<Status tone={online ? "success" : "neutral"}>{online ? tx("API connected", "الواجهة متصلة") : tx("API offline", "الواجهة غير متصلة")}</Status>}
          title={
            <>
              {APP_NAME}
              <br />
              <TextFlip
                className="text-nq-brand"
                phrases={ar ? ["واجهة Go وتطبيق React", "مولّدات بأمر واحد", "ملف تنفيذي واحد"] : ["Go API + React UI", "one-command generators", "one binary to ship"]}
              />
            </>
          }
          description={tx(
            "Built with ToGO: typed REST and GraphQL from one schema, an admin dashboard, auth and a React front end, shipped as a single binary.",
            "مبنيّ بـ ToGO: واجهات REST وGraphQL مُنمّطة من مخطط واحد، ولوحة إدارة، ومصادقة، وواجهة React، في ملف تنفيذي واحد.",
          )}
          actions={
            me ? (
              <Link to="/dashboard" className={primary}>{tx("Go to dashboard", "اذهب إلى لوحة التحكم")} <Arrow /></Link>
            ) : (
              <>
                <Link to="/register" className={primary}>{tx("Create account", "إنشاء حساب")} <Arrow /></Link>
                <Link to="/login" className={secondary}>{tx("Sign in", "تسجيل الدخول")}</Link>
              </>
            )
          }
          proof={`ToGO ${version} · ${tx("powered by Go", "مدعوم بـ Go")}`}
          frame="window"
          frameTitle="terminal"
          mockupLabel={tx("The ToGO CLI serving this app", "أداة ToGO تشغّل هذا التطبيق")}
          mockup={<Terminal lines={bootLines(version)} title="togo" height={300} />}
        />

        <FeatureGrid
          eyebrow={tx("Start here", "ابدأ من هنا")}
          title={tx("Where to next", "إلى أين بعد ذلك")}
          description={tx("Everything a fresh ToGO app gives you, one click away.", "كل ما يقدّمه تطبيق ToGO جديد، على بُعد نقرة.")}
          features={next}
        />

        <HowItWorks
          eyebrow={tx("How it works", "كيف يعمل")}
          title={tx("From schema to shipped", "من المخطط إلى الإطلاق")}
          steps={[
            { icon: <TerminalIcon />, title: tx("Make a resource", "أنشئ موردًا"), description: <code className="text-caption">togo make:resource Post title:string</code> },
            { icon: <Database />, title: tx("Generate", "ولّد"), description: tx("Migrations, sqlc queries, REST, GraphQL and the admin page.", "الترحيلات واستعلامات sqlc وREST وGraphQL وصفحة الإدارة.") },
            { icon: <Rocket />, title: tx("Ship one binary", "أطلق ملفًا واحدًا"), description: <code className="text-caption">togo build && togo deploy</code> },
          ]}
        />

        <CtaBanner
          title={me ? tx("Your dashboard is ready", "لوحة التحكم جاهزة") : tx(`Start using ${APP_NAME}`, `ابدأ باستخدام ${APP_NAME}`)}
          description={tx("Manage your account, roles and every resource you generate.", "أدِر حسابك وأدوارك وكل مورد تولّده.")}
          action={
            me ? <Link to="/dashboard" className={primary}>{tx("Open dashboard", "افتح لوحة التحكم")}</Link>
               : <Link to="/register" className={primary}>{tx("Create account", "إنشاء حساب")}</Link>
          }
          secondaryAction={<a href={DOCS} target="_blank" rel="noreferrer" className={secondary}>{tx("Read the docs", "اقرأ التوثيق")}</a>}
        />
      </main>

      <footer className="border-t border-border">
        <div className="mx-auto flex w-full max-w-6xl flex-wrap items-center justify-between gap-3 px-4 py-6 text-caption text-muted-foreground sm:px-8">
          <span>{APP_NAME} · ToGO {version}</span>
          <span>{tx("Built with ToGO and Nasaq", "مبنيّ بـ ToGO وNasaq")}</span>
        </div>
      </footer>
    </div>
  );
}
