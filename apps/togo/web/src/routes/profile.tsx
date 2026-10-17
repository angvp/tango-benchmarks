import { useCallback, useEffect, useState } from "react";
import { KeyRound, ShieldCheck, User, UserCog } from "lucide-react";
import {
  AccountSettings, ApiKeys, Badge, ChangePasswordForm, Infolist, LoadingState, ProfileForm, SettingsSection, TwoFactorSetup,
  type AccountSettingsItem, type ApiKeyRecord, type ProfileValues,
} from "@fadymondy/nasaq/web";
import { auth, sessionMe, type Me } from "../lib/auth";
import { useLang } from "../lib/i18n";

// Add a language here to offer it across the app (and to NasaqProvider's `locales`).
const LANGUAGES = [
  { value: "en", label: "English" },
  { value: "ar", label: "العربية" },
];

// Abilities a personal access token can carry. Match them to your permissions.
const SCOPES = [
  { id: "read", label: "Read", description: "Read your data through the API." },
  { id: "write", label: "Write", description: "Create and change data through the API." },
];

// The auth plugin stores the email and password only; the rest of the profile
// lives on this device until your app adds a profile endpoint.
const PROFILE_KEY = "app.profile";

function loadProfile(me: Me, locale: string): ProfileValues {
  const saved = JSON.parse(localStorage.getItem(PROFILE_KEY) ?? "{}") as Partial<ProfileValues>;
  return {
    name: saved.name ?? me.email.split("@")[0],
    username: saved.username ?? "",
    email: me.email,
    phone: saved.phone ?? "",
    bio: saved.bio ?? "",
    locale,
    timezone: saved.timezone ?? Intl.DateTimeFormat().resolvedOptions().timeZone,
  };
}

const message = (e: unknown) => (e instanceof Error ? e.message : String(e));

export function Profile() {
  const { locale, setLocale, tx } = useLang();
  const [me, setMe] = useState<Me | null>(null);
  useEffect(() => { sessionMe().then(setMe); }, []);

  if (!me) return <LoadingState />;

  const items: AccountSettingsItem[] = [
    { id: "profile", label: tx("Profile", "الملف الشخصي"), icon: User },
    { id: "security", label: tx("Security", "الأمان"), icon: ShieldCheck },
    { id: "tokens", label: tx("API tokens", "رموز الواجهة"), icon: KeyRound },
    { id: "access", label: tx("Roles & access", "الأدوار والصلاحيات"), icon: UserCog },
  ];

  return (
    <AccountSettings
      title={tx("Account settings", "إعدادات الحساب")}
      description={me.email}
      items={items}
    >
      {(section) =>
        section === "profile" ? (
          <ProfileForm
            values={loadProfile(me, locale)}
            languages={LANGUAGES}
            onSubmit={async ({ email: _email, locale: next, ...rest }) => {
              localStorage.setItem(PROFILE_KEY, JSON.stringify(rest));
              if (next !== locale) setLocale(next);
            }}
          />
        ) : section === "security" ? (
          <Security />
        ) : section === "tokens" ? (
          <Tokens />
        ) : (
          <SettingsSection
            title={tx("Roles & access", "الأدوار والصلاحيات")}
            description={tx("What your account can do in this app. An admin changes it.", "ما يمكن لحسابك فعله في هذا التطبيق. يغيّره المسؤول.")}
          >
            <Infolist
              items={[
                {
                  id: "roles", label: "Roles", labelAr: "الأدوار", value: me.roles ?? ["user"],
                  render: () => <div className="flex flex-wrap gap-1.5">{(me.roles ?? ["user"]).map((r) => <Badge key={r} variant="brand">{r}</Badge>)}</div>,
                },
                { id: "permissions", label: "Permissions", labelAr: "الصلاحيات", type: "list", value: me.permissions ?? [], wide: true },
              ]}
            />
          </SettingsSection>
        )
      }
    </AccountSettings>
  );
}

function Security() {
  const { tx } = useLang();
  // Enrolling tells us the state: a fresh secret when off, "already enabled" when on.
  const [twoFactor, setTwoFactor] = useState<{ enabled: boolean; uri: string; secret?: string } | null>(null);
  const load2fa = useCallback(() => {
    auth.enroll2fa()
      .then((r) => setTwoFactor({ enabled: false, uri: r.otpauth_url, secret: r.secret }))
      .catch(() => setTwoFactor({ enabled: true, uri: "" }));
  }, []);
  useEffect(load2fa, [load2fa]);

  return (
    <div className="flex flex-col gap-6">
      <SettingsSection
        title={tx("Password", "كلمة المرور")}
        description={tx("Use a long password you do not use anywhere else.", "استخدم كلمة مرور طويلة لا تستخدمها في مكان آخر.")}
      >
        <ChangePasswordForm
          showSignOutOthers={false}
          onSubmit={async ({ currentPassword, newPassword }) => {
            try {
              await auth.changePassword(currentPassword, newPassword);
            } catch (e) {
              const m = message(e);
              return /current password/i.test(m) ? { fieldErrors: { currentPassword: m } } : { error: m };
            }
          }}
        />
      </SettingsSection>

      <SettingsSection
        title={tx("Two-factor authentication", "المصادقة الثنائية")}
        description={tx("Ask for a code from an authenticator app at sign-in.", "اطلب رمزًا من تطبيق المصادقة عند تسجيل الدخول.")}
      >
        {!twoFactor ? <LoadingState /> : (
          <TwoFactorSetup
            otpauthUri={twoFactor.uri}
            secret={twoFactor.secret}
            enabled={twoFactor.enabled}
            confirmWith="code"
            onVerify={async (code) => {
              try { await auth.verify2fa(code); } catch (e) { return { error: message(e) }; }
            }}
            onComplete={() => setTwoFactor((t) => t && { ...t, enabled: true })}
            onDisable={async (code) => {
              try { await auth.disable2fa(code); } catch (e) { return { error: message(e) }; }
              load2fa();
            }}
          />
        )}
      </SettingsSection>
    </div>
  );
}

function Tokens() {
  const { tx } = useLang();
  const [keys, setKeys] = useState<ApiKeyRecord[] | null>(null);
  const load = useCallback(() => {
    auth.tokens().then((list) =>
      setKeys(list.map((t) => ({
        id: t.id,
        name: t.name,
        prefix: "togo_pat_",
        scopes: t.abilities,
        createdAt: t.created_at,
        expiresAt: t.expires_at || null,
      }))),
    );
  }, []);
  useEffect(load, [load]);

  return (
    <SettingsSection
      title={tx("API tokens", "رموز الواجهة")}
      description={tx("Personal access tokens call the API as you, limited to the scopes you pick.", "رموز وصول شخصية تستدعي الواجهة باسمك، ضمن النطاقات التي تختارها.")}
    >
      {!keys ? <LoadingState /> : (
        <ApiKeys
          keys={keys}
          scopes={SCOPES}
          onCreate={async ({ name, scopes, expiresInDays }) => {
            try {
              const { token } = await auth.createToken(name, scopes, expiresInDays ? expiresInDays * 24 : 0);
              load();
              return { secret: token };
            } catch (e) {
              return { error: message(e) };
            }
          }}
          onRevoke={async (id) => {
            await auth.revokeToken(id);
            load();
          }}
        />
      )}
    </SettingsSection>
  );
}
