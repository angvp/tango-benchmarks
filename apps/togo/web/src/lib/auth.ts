// togo auth client — talks to the auth plugin's /api/auth/* endpoints.
// Session is an HttpOnly cookie; CSRF uses the double-submit token. While an admin
// impersonates someone, account calls go through apiFetch with that user's bearer.
import { API, apiFetch } from "./api";

async function csrf(): Promise<string> {
  const res = await fetch(`${API}/api/auth/csrf`, { credentials: "include" });
  const data = await res.json().catch(() => ({}));
  return data.csrf_token ?? "";
}

async function post<T = any>(path: string, body?: unknown): Promise<T> {
  const token = await csrf();
  const res = await apiFetch(`/api/auth/${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json", "X-CSRF-Token": token },
    body: body ? JSON.stringify(body) : undefined,
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || data.detail || `request failed (${res.status})`);
  return data as T;
}

export interface AuthMethod { name: string; label: string; type: string; url: string }

export interface Me { email: string; roles?: string[]; permissions?: string[]; [k: string]: unknown }

export const auth = {
  login: (email: string, password: string) => post("login", { email, password }),
  register: (email: string, password: string) => post("register", { email, password }),
  // Always ends the signed-in (cookie) session, never an impersonation token.
  logout: async () => {
    const token = await csrf();
    const res = await fetch(`${API}/api/auth/logout`, { method: "POST", credentials: "include", headers: { "X-CSRF-Token": token } });
    if (!res.ok) throw new Error(`sign out failed (${res.status})`);
  },
  me: async (): Promise<Me | null> => {
    const res = await apiFetch("/api/auth/me");
    if (!res.ok) return null;
    return res.json();
  },
  methods: async (): Promise<AuthMethod[]> => {
    const res = await fetch(`${API}/api/auth/methods`, { credentials: "include" }).catch(() => null);
    if (!res || !res.ok) return [];
    const d = await res.json().catch(() => ({ methods: [] }));
    return d.methods ?? [];
  },
  requestOtp: (email: string, purpose = "reset") => post("otp", { email, purpose }),
  verifyOtp: (email: string, code: string, purpose = "reset") => post("otp/verify", { email, code, purpose }),
  // Completes a reset link (AUTH_RESET_PATH, default /reset-password?token=…).
  resetPassword: (token: string, password: string) => post("password/reset", { token, password }),
  // Ends an impersonation server-side (revokes the token, writes the audit event).
  stopImpersonation: () => post("impersonation/stop"),

  // Account security (signed in).
  changePassword: (oldPassword: string, newPassword: string) =>
    post("change-password", { old_password: oldPassword, new_password: newPassword }),
  // Starts TOTP enrolment. Rejects with "2fa already enabled…" when it is on.
  enroll2fa: () => post<{ secret: string; otpauth_url: string }>("2fa/enroll"),
  verify2fa: (code: string) => post("2fa/verify", { code }),
  disable2fa: (code: string) => post("2fa/disable", { code }),

  // Personal access tokens: the plaintext token is returned once, on create.
  tokens: async (): Promise<AccessToken[]> => {
    const res = await apiFetch("/api/auth/tokens");
    return res.ok ? res.json() : [];
  },
  createToken: (name: string, abilities: string[], expiresInHours = 0) =>
    post<{ token: string }>("tokens", { name, abilities, expires_in_hours: expiresInHours }),
  revokeToken: async (id: string) => {
    const token = await csrf();
    const res = await apiFetch(`/api/auth/tokens/${encodeURIComponent(id)}`, {
      method: "DELETE",
      headers: { "X-CSRF-Token": token },
    });
    if (!res.ok) throw new Error(`revoke failed (${res.status})`);
  },
};

export interface AccessToken { id: string; name: string; abilities: string[]; created_at: string; expires_at: string }

// Session cache so the router's beforeLoad guards resolve /me once per navigation
// pass instead of re-fetching on every route. Clear it after login/logout/register.
let _meCache: Promise<Me | null> | null = null;
export function sessionMe(force = false): Promise<Me | null> {
  if (force || !_meCache) _meCache = auth.me();
  return _meCache;
}
export function clearSession() { _meCache = null; }
