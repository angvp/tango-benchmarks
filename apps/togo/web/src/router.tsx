import { createRootRoute, createRoute, createRouter, lazyRouteComponent, Outlet, redirect } from "@tanstack/react-router";
import { LogoLoader } from "@fadymondy/nasaq/web";
import { Providers } from "./providers";
import { sessionMe } from "./lib/auth";
import { Welcome } from "./routes/welcome";
import { Login } from "./routes/login";
import { Register } from "./routes/register";
import { Reset } from "./routes/reset";
import { ResetPassword } from "./routes/reset-password";
import { AppLayout } from "./routes/app-layout";

// The authenticated admin surface (dashboard charts/widgets, the
// resource tables/forms/infolists) is the heavy part of the bundle — lazy-load it
// so it splits into its own chunk and the public/auth first paint stays small.
// The router's pending component (the ToGO LogoLoader) shows while the chunk loads.
const Dashboard = lazyRouteComponent(() => import("./routes/dashboard"), "Dashboard");
const AdminHome = lazyRouteComponent(() => import("./routes/admin"), "AdminHome");
const AdminResource = lazyRouteComponent(() => import("./routes/admin-resource"), "AdminResource");
const Users = lazyRouteComponent(() => import("./routes/users"), "Users");
const Profile = lazyRouteComponent(() => import("./routes/profile"), "Profile");

const rootRoute = createRootRoute({ component: () => (<Providers><Outlet /></Providers>) });

// Already signed in → skip the auth pages and go straight to the dashboard.
const redirectIfAuthed = async () => {
  if (await sessionMe()) throw redirect({ to: "/dashboard" });
};

const indexRoute = createRoute({ getParentRoute: () => rootRoute, path: "/", component: Welcome });
const loginRoute = createRoute({ getParentRoute: () => rootRoute, path: "/login", component: Login, beforeLoad: redirectIfAuthed });
const registerRoute = createRoute({ getParentRoute: () => rootRoute, path: "/register", component: Register, beforeLoad: redirectIfAuthed });
const resetRoute = createRoute({ getParentRoute: () => rootRoute, path: "/reset", component: Reset });
// Reset links from auth (AUTH_RESET_PATH, default /reset-password?token=…).
const resetPasswordRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/reset-password",
  component: ResetPassword,
  validateSearch: (s: Record<string, unknown>): { token?: string } => (typeof s.token === "string" ? { token: s.token } : {}),
});

// Protected shell. The guard runs in beforeLoad — BEFORE the layout/children render —
// so unauthenticated visitors are redirected to /login without the private page ever
// painting (the router shows the pending loader while the check runs). The resolved
// user is returned as route context so children don't re-fetch /me.
const appRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: "_app",
  component: AppLayout,
  beforeLoad: async () => {
    const me = await sessionMe();
    if (!me) throw redirect({ to: "/login" });
    return { me };
  },
});
const dashboardRoute = createRoute({ getParentRoute: () => appRoute, path: "/dashboard", component: Dashboard });
const adminRoute = createRoute({ getParentRoute: () => appRoute, path: "/admin", component: AdminHome });
const resourceRoute = createRoute({ getParentRoute: () => appRoute, path: "/admin/$resource", component: AdminResource });
const usersRoute = createRoute({ getParentRoute: () => appRoute, path: "/users", component: Users });
const profileRoute = createRoute({ getParentRoute: () => appRoute, path: "/profile", component: Profile });

const routeTree = rootRoute.addChildren([
  indexRoute, loginRoute, registerRoute, resetRoute, resetPasswordRoute,
  appRoute.addChildren([dashboardRoute, adminRoute, resourceRoute, usersRoute, profileRoute]),
]);

export const router = createRouter({
  routeTree,
  defaultPreload: "intent",
  // Branded full-screen loader while a route's beforeLoad (e.g. the auth check) runs.
  // 150ms delay so cached/instant navigations don't flash it.
  defaultPendingComponent: () => <div className="grid min-h-screen place-items-center"><LogoLoader /></div>,
  defaultPendingMs: 150,
  defaultPendingMinMs: 300,
});

declare module "@tanstack/react-router" {
  interface Register { router: typeof router }
}
