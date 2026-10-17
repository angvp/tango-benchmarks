import { Link, useSearch } from "@tanstack/react-router";
import { AuthLayout, ResetPasswordForm } from "@fadymondy/nasaq/web";
import { auth } from "../lib/auth";
import { useLang } from "../lib/i18n";
import { AppAuthFooter } from "../components/auth-footer";

// The page a password-reset link opens: auth sends `AUTH_RESET_PATH?token=…`
// (default /reset-password) from the admin "reset password" action and the
// forgot-password email. The token is single-use and expires.
export function ResetPassword() {
  const { tx } = useLang();
  const { token } = useSearch({ from: "/reset-password" });
  return (
    <AuthLayout
      title={tx("Choose a new password", "اختر كلمة مرور جديدة")}
      prompt={<><Link to="/login" className="font-medium text-primary hover:underline">{tx("Back to sign in", "العودة لتسجيل الدخول")}</Link></>}
      footer={<AppAuthFooter />}
    >
      <ResetPasswordForm
        defaultState={token ? "idle" : "expired"}
        signIn="/login"
        requestLink="/reset"
        onSubmit={async ({ password }) => {
          try {
            await auth.resetPassword(token ?? "", password);
          } catch (e) {
            const message = (e as Error).message;
            if (/invalid or expired/i.test(message)) return { expired: true };
            return { error: message };
          }
        }}
      />
    </AuthLayout>
  );
}
