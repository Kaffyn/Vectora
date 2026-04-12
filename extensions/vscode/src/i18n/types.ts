/**
 * i18n Framework Types - Phase 11: RTL Languages Support & i18n Enhancement
 * Defines types for translation system with namespace support, RTL support, and language metadata
 */

// LTR Languages (8): English, Portuguese, Spanish, French, German, Italian, Japanese, Chinese
// RTL Languages (4): Arabic, Hebrew, Persian, Urdu
// Total: 12 languages
export type SupportedLanguage =
  | "en-US"   // English (United States)
  | "pt-BR"   // Portuguese (Brazil)
  | "pt-PT"   // Portuguese (Portugal)
  | "es-ES"   // Spanish (Spain)
  | "fr-FR"   // French (France)
  | "de-DE"   // German (Germany)
  | "it-IT"   // Italian (Italy)
  | "ja-JP"   // Japanese (Japan)
  | "zh-CN"   // Chinese Simplified (China)
  | "zh-TW"   // Chinese Traditional (Taiwan)
  | "ko-KR"   // Korean (South Korea)
  | "ar-SA"   // Arabic (Saudi Arabia) - RTL
  | "he-IL"   // Hebrew (Israel) - RTL
  | "fa-IR"   // Persian (Iran) - RTL
  | "ur-PK";  // Urdu (Pakistan) - RTL

/**
 * Language metadata for rich language information
 */
export interface LanguageMetadata {
  /** Language code (e.g., "ar-SA") */
  code: SupportedLanguage;
  /** Native name (e.g., "العربية") */
  nativeName: string;
  /** English name (e.g., "Arabic") */
  englishName: string;
  /** Region name (e.g., "Saudi Arabia") */
  region: string;
  /** Text direction: ltr (left-to-right) or rtl (right-to-left) */
  direction: "ltr" | "rtl";
  /** Date format (e.g., "DD/MM/YYYY" or "YYYY-MM-DD") */
  dateFormat: string;
  /** Time format (e.g., "24h" or "12h") */
  timeFormat: "24h" | "12h";
  /** Number format (e.g., "1,234.56" or "1.234,56") */
  numberFormat: {
    decimal: string;
    thousands: string;
    currency?: string;
  };
  /** Alternative languages (e.g., "zh-TW" for Traditional Chinese) */
  alternativeLanguages?: SupportedLanguage[];
}

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
 * RTL configuration
 */
export interface RTLConfig {
  /** Enable automatic RTL detection and direction management */
  enabled: boolean;
  /** Apply CSS class for RTL direction (default: "rtl") */
  rtlClass?: string;
  /** Apply CSS class for LTR direction (default: "ltr") */
  ltrClass?: string;
  /** Use CSS logical properties (start/end instead of left/right) */
  useCSSLogicalProperties?: boolean;
  /** Auto-flip component layouts for RTL */
  autoFlipLayout?: boolean;
  /** Flip icon directions for RTL (e.g., arrow-right becomes arrow-left) */
  flipIcons?: boolean;
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
  /** RTL configuration */
  rtl?: RTLConfig;
}

/**
 * Translation function signature
 */
export type TranslateFunction = (
  key: string,
  options?: TranslationOptions,
) => string;

/**
 * RTL API
 */
export interface RTLAPI {
  /** Current text direction */
  direction: "ltr" | "rtl";
  /** Check if current language is RTL */
  isRTL: (language?: SupportedLanguage) => boolean;
  /** Get direction class (rtl or ltr) */
  getDirectionClass: (language?: SupportedLanguage) => string;
  /** Apply RTL styles to element */
  applyRTLStyles: (element: HTMLElement) => void;
  /** Remove RTL styles from element */
  removeRTLStyles: (element: HTMLElement) => void;
}

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
  /** Get language metadata */
  getLanguageMetadata: (language: SupportedLanguage) => LanguageMetadata | null;
  /** Get all available language metadata */
  getAllLanguagesMetadata: () => LanguageMetadata[];
  /** Get RTL API */
  rtl: RTLAPI;
}

/**
 * Translation namespace loader
 */
export type NamespaceLoader = (language: SupportedLanguage) => Promise<TranslationNamespace>;
