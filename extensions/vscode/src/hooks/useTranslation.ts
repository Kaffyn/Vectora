/**
 * useTranslation - React hook for translations
 * Provides access to translation function and language management
 */

import { useContext, useCallback, useState, useEffect } from "react";
import { TranslationContext } from "../i18n/TranslationContext";
import { TranslationOptions } from "../i18n/types";

/**
 * Hook to access translation function and language controls
 * @returns Object with translation function, current language, and language changer
 */
export function useTranslation() {
  const context = useContext(TranslationContext);

  if (!context) {
    // Fallback for components rendered outside provider
    return {
      t: (key: string) => key,
      i18n: {
        language: "en" as const,
        changeLanguage: async () => {},
      },
      locale: "en" as const,
      setLocale: async () => {},
    };
  }

  const { t, i18n } = context;

  /**
   * Memoized translation function with full options support
   */
  const translate = useCallback(
    (key: string, options?: TranslationOptions | Record<string, any>) => {
      // Support both TranslationOptions and simple object
      if (options && !("variables" in options)) {
        return t(key, { variables: options as Record<string, any> });
      }
      return t(key, options as TranslationOptions);
    },
    [t],
  );

  return {
    /** Translation function: t("key") or t("key", { variables: { name: "João" } }) */
    t: translate,
    /** Alias for language from i18n context */
    locale: i18n.language,
    /** Change language: setLocale("pt-BR") */
    setLocale: i18n.changeLanguage,
    /** Full i18n object for advanced usage */
    i18n,
  };
}

/**
 * Hook to get only the translation function
 * Lighter weight when only translations are needed
 */
export function useTranslator() {
  const context = useContext(TranslationContext);

  if (!context) {
    return (key: string) => key;
  }

  return context.t;
}

/**
 * Hook to get current locale and change it
 */
export function useLocale() {
  const context = useContext(TranslationContext);
  const [locale, setLocale] = useState(context?.i18n.language || "en");

  useEffect(() => {
    if (context) {
      setLocale(context.i18n.language);
    }
  }, [context?.i18n.language]);

  const changeLocale = useCallback(
    async (newLocale: string) => {
      if (context) {
        await context.i18n.changeLanguage(newLocale);
        setLocale(newLocale);
      }
    },
    [context],
  );

  return {
    locale,
    setLocale: changeLocale,
    availableLocales: ["en-US", "pt-BR", "es-ES"] as const,
  };
}

/**
 * Hook for component translations with nested keys support
 * Useful for translations organized by component
 */
export function useComponentTranslation(componentNamespace: string) {
  const { t } = useTranslation();

  const componentT = useCallback(
    (key: string, options?: TranslationOptions | Record<string, any>) => {
      const fullKey = `${componentNamespace}:${key}`;
      if (options && !("variables" in options)) {
        return t(fullKey, { variables: options as Record<string, any> });
      }
      return t(fullKey, options as TranslationOptions);
    },
    [t, componentNamespace],
  );

  return componentT;
}

/**
 * Hook for plural translations
 */
export function usePluralTranslation() {
  const { t } = useTranslation();

  const pluralT = useCallback(
    (key: string, count: number, options?: TranslationOptions) => {
      // Simple plural handling: add "_plural" suffix for count !== 1
      const pluralKey = count !== 1 ? `${key}_plural` : key;
      return t(pluralKey, { ...options, variables: { ...options?.variables, count } });
    },
    [t],
  );

  return pluralT;
}
