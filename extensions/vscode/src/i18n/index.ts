/**
 * i18n Framework - Core Engine
 * Lightweight, efficient translation system with namespace support
 */

import {
  SupportedLanguage,
  I18nConfig,
  I18nAPI,
  TranslateFunction,
  TranslationNamespace,
  TranslationOptions,
  NamespaceLoader,
} from "./types";

class I18nEngine implements I18nAPI {
  private currentLanguage: SupportedLanguage;
  private fallbackLanguage: SupportedLanguage;
  private config: Required<I18nConfig>;
  private namespaces: Map<string, Map<SupportedLanguage, TranslationNamespace>>;
  private loaders: Map<string, NamespaceLoader>;
  private loadingPromises: Map<string, Promise<void>>;

  constructor(config: I18nConfig) {
    this.currentLanguage = config.defaultLanguage;
    this.fallbackLanguage = config.fallbackLanguage || config.defaultLanguage;
    this.config = {
      defaultLanguage: config.defaultLanguage,
      fallbackLanguage: this.fallbackLanguage,
      debug: config.debug || false,
      nsSeparator: config.nsSeparator || ":",
      keySeparator: config.keySeparator || ".",
      interpolationPrefix: config.interpolationPrefix || "{{",
      interpolationSuffix: config.interpolationSuffix || "}}",
    };
    this.namespaces = new Map();
    this.loaders = new Map();
    this.loadingPromises = new Map();
  }

  /**
   * Register a namespace loader
   */
  registerNamespaceLoader(namespace: string, loader: NamespaceLoader): void {
    this.loaders.set(namespace, loader);
    if (this.config.debug) {
      console.log(`[i18n] Registered loader for namespace: ${namespace}`);
    }
  }

  /**
   * Get current language
   */
  get language(): SupportedLanguage {
    return this.currentLanguage;
  }

  /**
   * Check if a language is supported
   */
  isLanguageSupported(lang: string): boolean {
    return ["en-US", "pt-BR", "es-ES"].includes(lang);
  }

  /**
   * Get available languages
   */
  getAvailableLanguages(): SupportedLanguage[] {
    return ["en-US", "pt-BR", "es-ES"];
  }

  /**
   * Change current language
   */
  async changeLanguage(lang: SupportedLanguage): Promise<void> {
    if (!this.isLanguageSupported(lang)) {
      console.warn(`[i18n] Language not supported: ${lang}`);
      return;
    }

    this.currentLanguage = lang;

    // Load all registered namespaces for the new language
    const promises = Array.from(this.loaders.entries()).map(([ns]) =>
      this.loadNamespace(ns, lang),
    );

    await Promise.all(promises);

    if (this.config.debug) {
      console.log(`[i18n] Language changed to: ${lang}`);
    }
  }

  /**
   * Load a namespace for a specific language
   */
  async loadNamespace(namespace: string, language: SupportedLanguage): Promise<void> {
    // Check if already loaded
    if (
      this.namespaces.has(namespace) &&
      this.namespaces.get(namespace)!.has(language)
    ) {
      return;
    }

    // Check if already loading
    const cacheKey = `${namespace}:${language}`;
    if (this.loadingPromises.has(cacheKey)) {
      return this.loadingPromises.get(cacheKey)!;
    }

    // Load via registered loader
    const loader = this.loaders.get(namespace);
    if (!loader) {
      if (this.config.debug) {
        console.warn(`[i18n] No loader found for namespace: ${namespace}`);
      }
      return;
    }

    const loadingPromise = (async () => {
      try {
        const data = await loader(language);
        if (!this.namespaces.has(namespace)) {
          this.namespaces.set(namespace, new Map());
        }
        this.namespaces.get(namespace)!.set(language, data);

        if (this.config.debug) {
          console.log(`[i18n] Loaded namespace: ${namespace} (${language})`);
        }
      } catch (error) {
        console.error(`[i18n] Failed to load namespace: ${namespace} (${language})`, error);
      } finally {
        this.loadingPromises.delete(cacheKey);
      }
    })();

    this.loadingPromises.set(cacheKey, loadingPromise);
    return loadingPromise;
  }

  /**
   * Get translation function
   */
  getTranslator(): TranslateFunction {
    return this.translate.bind(this);
  }

  /**
   * Translate a key with optional interpolation
   * Key format: "namespace:key.subkey"
   * or just "key.subkey" for global namespace
   */
  private translate(key: string, options?: TranslationOptions): string {
    // Parse namespace and key
    const [namespace, keyPart] = this.parseKey(key);

    // Attempt to get translation
    const translation = this.getValue(namespace, keyPart, this.currentLanguage);

    if (translation) {
      return this.interpolate(translation, options?.variables);
    }

    // Try fallback language
    if (this.currentLanguage !== this.fallbackLanguage) {
      const fallbackTranslation = this.getValue(namespace, keyPart, this.fallbackLanguage);
      if (fallbackTranslation) {
        return this.interpolate(fallbackTranslation, options?.variables);
      }
    }

    // Return default value or original key
    const result = options?.defaultValue || key;
    if (this.config.debug) {
      console.warn(`[i18n] Missing translation for key: ${key}`);
    }
    return result;
  }

  /**
   * Parse key into namespace and path
   * "namespace:a.b.c" => ["namespace", "a.b.c"]
   * "a.b.c" => ["common", "a.b.c"]
   */
  private parseKey(key: string): [string, string] {
    if (key.includes(this.config.nsSeparator)) {
      const [ns, ...rest] = key.split(this.config.nsSeparator);
      return [ns, rest.join(this.config.nsSeparator)];
    }
    return ["common", key];
  }

  /**
   * Get value from nested object using dot notation
   */
  private getValue(
    namespace: string,
    keyPath: string,
    language: SupportedLanguage,
  ): string | null {
    const ns = this.namespaces.get(namespace);
    if (!ns) {
      return null;
    }

    const data = ns.get(language);
    if (!data) {
      return null;
    }

    const keys = keyPath.split(this.config.keySeparator);
    let value: any = data;

    for (const k of keys) {
      if (value && typeof value === "object" && k in value) {
        value = value[k];
      } else {
        return null;
      }
    }

    return typeof value === "string" ? value : null;
  }

  /**
   * Interpolate variables in translation string
   * "Hello {{name}}" + { name: "John" } => "Hello John"
   */
  private interpolate(
    text: string,
    variables?: Record<string, string | number | boolean>,
  ): string {
    if (!variables) {
      return text;
    }

    let result = text;
    for (const [key, value] of Object.entries(variables)) {
      const pattern = new RegExp(
        `${this.escapeRegex(this.config.interpolationPrefix)}${key}${this.escapeRegex(this.config.interpolationSuffix)}`,
        "g",
      );
      result = result.replace(pattern, String(value));
    }
    return result;
  }

  /**
   * Escape special regex characters
   */
  private escapeRegex(str: string): string {
    return str.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  }
}

// Global instance
let globalI18n: I18nEngine | null = null;

/**
 * Initialize i18n with configuration
 */
export function initI18n(config: I18nConfig): I18nAPI {
  globalI18n = new I18nEngine(config);
  return globalI18n;
}

/**
 * Get the global i18n instance
 */
export function getI18n(): I18nAPI {
  if (!globalI18n) {
    throw new Error(
      "i18n not initialized. Call initI18n() first.",
    );
  }
  return globalI18n;
}

/**
 * Convenience function to translate
 */
export function t(key: string, options?: TranslationOptions): string {
  return getI18n().getTranslator()(key, options);
}

export * from "./types";
