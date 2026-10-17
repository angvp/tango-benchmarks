import { useNasaq } from "@fadymondy/nasaq/web";

/** EN/AR helper over the Nasaq locale: `tx("Save", "حفظ")` picks the active language.
 * The locale (and the page direction) lives in NasaqProvider; switch it with `setLocale`. */
export function useLang() {
  const { locale, setLocale, isRtl } = useNasaq();
  const ar = locale.startsWith("ar");
  return { locale, setLocale, ar, isRtl, tx: (en: string, arText: string) => (ar ? arText : en) };
}
