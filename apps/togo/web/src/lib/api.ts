// Same-origin by default: dev uses the vite proxy (vite.config.ts), prod is served by the Go backend.
export const API = import.meta.env.VITE_API_ORIGIN ?? "";
export const APP_NAME = import.meta.env.VITE_APP_NAME ?? "ToGO";

// Impersonation (auth plugin): POST /api/auth/admin/users/{id}/impersonate returns a
// short-lived bearer token that acts as the user. The admin's own cookie session is
// untouched, so the token lives in this tab only (sessionStorage) and is sent on the
// app's acting-as-user calls. Admin API calls never carry it: auth refuses them.
export interface Impersonation { token: string; email: string; expiresAt?: string; startedAt: string }

const IMP_KEY = "togo-impersonation";
export const IMPERSONATION_EVENT = "togo-impersonation";

export function getImpersonation(): Impersonation | null {
  try {
    const raw = sessionStorage.getItem(IMP_KEY);
    return raw ? (JSON.parse(raw) as Impersonation) : null;
  } catch {
    return null;
  }
}

export function setImpersonation(imp: Impersonation | null) {
  if (imp) sessionStorage.setItem(IMP_KEY, JSON.stringify(imp));
  else sessionStorage.removeItem(IMP_KEY);
  window.dispatchEvent(new Event(IMPERSONATION_EVENT));
}

/** fetch for the app's own API as the current user: the impersonation bearer when one
 * is active (no cookies, so the admin session cannot leak into it), else the session
 * cookie. A rejected bearer (expired, stopped, or the admin lost the role) ends the
 * impersonation and reloads back into the admin's own session. */
export async function apiFetch(path: string, init: RequestInit = {}): Promise<Response> {
  const imp = getImpersonation();
  if (!imp) return fetch(`${API}${path}`, { ...init, credentials: "include" });
  const headers = new Headers(init.headers);
  headers.set("Authorization", `Bearer ${imp.token}`);
  const res = await fetch(`${API}${path}`, { ...init, headers, credentials: "omit" });
  if (res.status === 401) {
    setImpersonation(null);
    window.location.assign("/users");
  }
  return res;
}
