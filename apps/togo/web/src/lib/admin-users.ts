// Admin user-management client for the auth plugin's /api/auth/admin/* API (admin
// role required, re-checked on every request). It always runs as the signed-in
// admin's cookie session, never an impersonation token (auth refuses those here),
// and sends the double-submit CSRF token on writes.
import { API, setImpersonation } from "./api";

export class AdminError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

async function csrf(): Promise<string> {
  const res = await fetch(`${API}/api/auth/csrf`, { credentials: "include" });
  return (await res.json().catch(() => ({}))).csrf_token ?? "";
}

async function req<T = any>(method: string, path: string, body?: unknown): Promise<T> {
  const headers: Record<string, string> = {};
  if (body !== undefined) headers["Content-Type"] = "application/json";
  if (method !== "GET") headers["X-CSRF-Token"] = await csrf();
  const res = await fetch(`${API}/api/auth/admin${path}`, {
    method,
    credentials: "include",
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new AdminError(data.error || data.detail || `request failed (${res.status})`, res.status);
  return data as T;
}

/** A user as the backend returns it. */
export interface AdminUser {
  id: string | number;
  email: string;
  roles?: string[];
  permissions?: string[];
  created_at?: string;
}
/** Result of a reset-password / magic-link call: the link, or `emailed` when SMTP delivered it. */
export interface AdminLinkResult { link?: string; emailed?: boolean; expires_at?: string }

export interface CreateUserInput {
  email: string;
  password?: string;
  roles: string[];
  permissions?: string[];
}
export interface EditUserPayload {
  email?: string;
  roles?: string[];
  permissions?: string[];
}
interface ImpersonationGrant { token: string; identity: { id: string; email: string }; expires_at: string }

export const adminUsers = {
  list: (q?: string): Promise<AdminUser[]> => req("GET", `/users${q ? `?q=${encodeURIComponent(q)}` : ""}`),
  create: (input: CreateUserInput): Promise<{ user: AdminUser }> => req("POST", "/users", input),
  update: (id: string, input: EditUserPayload): Promise<AdminUser> => req("PATCH", `/users/${id}`, input),
  remove: (id: string): Promise<{ deleted: boolean; id: string }> => req("DELETE", `/users/${id}`),
  /** Starts acting as the user in this tab; the admin's own session is left as it is. */
  impersonate: async (id: string): Promise<void> => {
    const g = await req<ImpersonationGrant>("POST", `/users/${id}/impersonate`);
    setImpersonation({ token: g.token, email: g.identity.email, expiresAt: g.expires_at, startedAt: new Date().toISOString() });
  },
  resetPassword: (id: string, password?: string): Promise<AdminLinkResult & { reset?: boolean }> =>
    req("POST", `/users/${id}/reset-password`, password ? { password } : {}),
  magicLink: (id: string): Promise<AdminLinkResult> => req("POST", `/users/${id}/magic-link`),
};
