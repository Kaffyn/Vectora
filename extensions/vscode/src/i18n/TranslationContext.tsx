import React, { createContext, useContext, ReactNode, useEffect, useCallback, useState } from "react";
import { TranslationOptions } from "./types";

let globalTranslations: Record<string, any> = {};

export const TranslationContext = createContext<{
  t: (key: string, options?: TranslationOptions | Record<string, any>) => string;
  i18n: {
    language: string;
    changeLanguage: (lang: string) => Promise<void>;
  };
} | null>(null);

/**
 * Component for inline translations with component interpolation
 * Example: <Trans i18nKey="welcome.message" components={{ bold: <strong /> }} />
 */
export const Trans: React.FC<{
  i18nKey: string;
  components?: Record<string, React.ReactElement>;
  values?: Record<string, any>;
}> = ({ i18nKey, components, values }) => {
  const context = useContext(TranslationContext);
  if (!context) return <>{i18nKey}</>;

  const { t } = context;
  const text = t(i18nKey, values);

  if (!components) return <>{text}</>;

  // Basic component interpolation
  const parts = text.split(/(<[^>]+>[^<]*<\/[^>]+>)/g);
  return (
    <>
      {parts.map((part, i) => {
        const match = part.match(/<([^>]+)>(.*)<\/\1>/);
        if (match) {
          const [_, tagName, content] = match;
          const Component = components[tagName];
          if (Component) {
            return React.cloneElement(Component, { key: i }, content);
          }
        }
        return part;
      })}
    </>
  );
};

interface TranslationProviderProps {
  children: ReactNode;
  defaultLanguage?: string;
  fallbackLanguage?: string;
  debug?: boolean;
}

/**
 * Translation provider component
 * Manages translations loading and language switching
 */
export const TranslationProvider: React.FC<TranslationProviderProps> = ({
  children,
  defaultLanguage = "en-US",
  fallbackLanguage = "en-US",
  debug = false,
}) => {
  const [lang, setLang] = useState(defaultLanguage);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    const handleMessage = (event: MessageEvent) => {
      const data = event.data;
      if (data.type === "translations") {
        globalTranslations = data.translations;
        if (debug) {
          console.log("[i18n] Translations loaded:", Object.keys(globalTranslations).length, "keys");
        }
      }
    };

    window.addEventListener("message", handleMessage);
    return () => window.removeEventListener("message", handleMessage);
  }, [debug]);

  const translate = useCallback(
    (key: string, options?: TranslationOptions | Record<string, any>): string => {
      // Parse namespace and key
      let namespace = "common";
      let keyPath = key;

      if (key.includes(":")) {
        const parts = key.split(":");
        namespace = parts[0];
        keyPath = parts.slice(1).join(":");
      }

      // Try to get translation from global translations
      const translation = globalTranslations[keyPath];
      if (translation && translation[lang]) {
        let text = translation[lang];
        // Handle interpolation
        const variables =
          options && "variables" in options ? options.variables : (options as Record<string, any>);
        if (variables) {
          Object.keys(variables).forEach((k) => {
            text = text.replace(new RegExp(`{{${k}}}`, "g"), String(variables[k]));
          });
        }
        return text;
      }

      // Try fallback language
      if (lang !== fallbackLanguage && translation && translation[fallbackLanguage]) {
        let text = translation[fallbackLanguage];
        const variables =
          options && "variables" in options ? options.variables : (options as Record<string, any>);
        if (variables) {
          Object.keys(variables).forEach((k) => {
            text = text.replace(new RegExp(`{{${k}}}`, "g"), String(variables[k]));
          });
        }
        return text;
      }

      if (debug) {
        console.warn(`[i18n] Missing translation for key: ${key}`);
      }
      return (options && "defaultValue" in options && options.defaultValue) || key;
    },
    [lang, fallbackLanguage, debug],
  );

  const changeLanguage = useCallback(
    async (newLang: string) => {
      setLoading(true);
      try {
        setLang(newLang);
        if (debug) {
          console.log(`[i18n] Language changed to: ${newLang}`);
        }
      } finally {
        setLoading(false);
      }
    },
    [debug],
  );

  const i18n = {
    language: lang,
    changeLanguage,
  };

  return (
    <TranslationContext.Provider value={{ t: translate, i18n }}>
      {children}
    </TranslationContext.Provider>
  );
};

/**
 * @deprecated Use useTranslation from useTranslation.ts instead
 */
export const useAppTranslation = () => {
  const context = useContext(TranslationContext);
  if (!context) {
    return { t: (key: string) => key, i18n: { language: "en", changeLanguage: async () => {} } };
  }
  return context;
};

export const useTranslation = useAppTranslation;

// Global i18n object for non-React usage
export const i18n = {
  get language() {
    return "en-US";
  },
  t: (key: string, options?: TranslationOptions | Record<string, any>) => {
    const cleanKey = key.includes(":") ? key.split(":")[1] : key;
    const translation = globalTranslations[cleanKey];
    if (translation && translation["en-US"]) {
      let text = translation["en-US"];
      const variables =
        options && "variables" in options ? options.variables : (options as Record<string, any>);
      if (variables) {
        Object.keys(variables).forEach((k) => {
          text = text.replace(new RegExp(`{{${k}}}`, "g"), String(variables[k]));
        });
      }
      return text;
    }
    return cleanKey;
  },
  changeLanguage: async () => {
    /* no-op globally for now */
  },
};

export default TranslationProvider;
