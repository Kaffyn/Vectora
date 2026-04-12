/**
 * i18n Framework Types
 * Defines types for translation system with namespace support and interpolation
 */

export type SupportedLanguage = "en-US" | "pt-BR" | "es-ES";

/**
 * Translation value can be a string or object with language variants
 */
export type TranslationValue = string | Record<SupportedLanguage, string>;

/**
 * Full translation namespace (e.g., { common: { cancel: "Cancel", confirm: "Confirm" } })
 */
export type TranslationNamespace = Record<string, any>;

/**
 * Options for translation function
 */
export interface TranslationOptions {
  /** Variables to interpolate into the translation */
  variables?: Record<string, string | number | boolean>;
  /** Fallback namespace to look in if key not found */
  defaultValue?: string;
  /** Plural count for plural translations */
  count?: number;
}

/**
 * i18n initialization configuration
 */
export interface I18nConfig {
  /** Default language */
  defaultLanguage: SupportedLanguage;
  /** Fallback language for missing translations */
  fallbackLanguage?: SupportedLanguage;
  /** Debug mode for logging */
  debug?: boolean;
  /** Namespace separator (default: ".") */
  nsSeparator?: string;
  /** Key separator for nested keys (default: ".") */
  keySeparator?: string;
  /** Interpolation prefix (default: "{{") */
  interpolationPrefix?: string;
  /** Interpolation suffix (default: "}}") */
  interpolationSuffix?: string;
}

/**
 * Translation function signature
 */
export type TranslateFunction = (
  key: string,
  options?: TranslationOptions,
) => string;

/**
 * i18n API
 */
export interface I18nAPI {
  /** Current language */
  language: SupportedLanguage;
  /** Change current language */
  changeLanguage: (lang: SupportedLanguage) => Promise<void>;
  /** Get translation function */
  getTranslator: () => TranslateFunction;
  /** Load translations for namespace */
  loadNamespace: (namespace: string, language: SupportedLanguage) => Promise<void>;
  /** Check if language is supported */
  isLanguageSupported: (lang: string) => boolean;
  /** Get available languages */
  getAvailableLanguages: () => SupportedLanguage[];
}

/**
 * Translation namespace loader
 */
export type NamespaceLoader = (language: SupportedLanguage) => Promise<TranslationNamespace>;
