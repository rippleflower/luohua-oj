import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";

type Locale = "zh" | "en";

const localeStorageKey = "luooj.locale";

const localeContext = createContext<{
  locale: Locale;
  setLocale: (locale: Locale) => void;
} | null>(null);

export function LocaleProvider({ children }: { children: ReactNode }) {
  const [locale, setLocale] = useState<Locale>(() => {
    if (typeof window === "undefined") {
      return "zh";
    }
    const stored = window.localStorage.getItem(localeStorageKey);
    return stored === "en" ? "en" : "zh";
  });

  useEffect(() => {
    window.localStorage.setItem(localeStorageKey, locale);
  }, [locale]);

  const value = useMemo(() => ({ locale, setLocale }), [locale]);
  return <localeContext.Provider value={value}>{children}</localeContext.Provider>;
}

export function useLocale() {
  const value = useContext(localeContext);
  if (!value) {
    throw new Error("useLocale must be used within LocaleProvider");
  }
  return value;
}
