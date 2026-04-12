/**
 * RTL Language Detector
 * Phase 11: RTL Languages Support & i18n Enhancement
 * Automatically detects RTL languages and provides direction information
 */

import type { SupportedLanguage } from "../types";

/**
 * Map of RTL languages
 */
const RTL_LANGUAGES: Set<SupportedLanguage> = new Set([
  "ar-SA", // Arabic (Saudi Arabia)
  "he-IL", // Hebrew (Israel)
  "fa-IR", // Persian (Iran)
  "ur-PK", // Urdu (Pakistan)
]);

/**
 * Map of LTR languages
 */
const LTR_LANGUAGES: Set<SupportedLanguage> = new Set([
  "en-US", // English (United States)
  "pt-BR", // Portuguese (Brazil)
  "pt-PT", // Portuguese (Portugal)
  "es-ES", // Spanish (Spain)
  "fr-FR", // French (France)
  "de-DE", // German (Germany)
  "it-IT", // Italian (Italy)
  "ja-JP", // Japanese (Japan)
  "zh-CN", // Chinese Simplified (China)
  "zh-TW", // Chinese Traditional (Taiwan)
  "ko-KR", // Korean (South Korea)
]);

/**
 * Detect if a language is RTL
 */
export function isRTLLanguage(language: SupportedLanguage): boolean {
  return RTL_LANGUAGES.has(language);
}

/**
 * Detect if a language is LTR
 */
export function isLTRLanguage(language: SupportedLanguage): boolean {
  return LTR_LANGUAGES.has(language);
}

/**
 * Get text direction for a language
 */
export function getTextDirection(language: SupportedLanguage): "ltr" | "rtl" {
  return isRTLLanguage(language) ? "rtl" : "ltr";
}

/**
 * Get all RTL languages
 */
export function getRTLLanguages(): SupportedLanguage[] {
  return Array.from(RTL_LANGUAGES);
}

/**
 * Get all LTR languages
 */
export function getLTRLanguages(): SupportedLanguage[] {
  return Array.from(LTR_LANGUAGES);
}

/**
 * Check if a string matches any RTL language code pattern
 * Useful for detecting RTL from browser locale
 */
export function detectRTLFromLocale(locale: string): boolean {
  // Normalize locale string
  const normalized = locale.toLowerCase().replace(/_/g, "-");

  // Check direct language code
  const languageCode = normalized.split("-")[0];
  const rtlCodes = ["ar", "he", "fa", "ur"];

  return rtlCodes.includes(languageCode);
}

/**
 * Get preferred language from browser locale
 * Returns RTL or LTR language based on browser setting
 */
export function getPreferredLanguageFromBrowser(): SupportedLanguage | null {
  if (typeof navigator === "undefined") {
    return null;
  }

  const browserLanguages = navigator.languages || [navigator.language];

  for (const lang of browserLanguages) {
    const normalized = lang.toLowerCase().replace(/_/g, "-");

    // Try exact match first
    if (isValidSupportedLanguage(normalized)) {
      return normalized as SupportedLanguage;
    }

    // Try language code match
    const languageCode = normalized.split("-")[0];
    const match = Array.from([...RTL_LANGUAGES, ...LTR_LANGUAGES]).find(
      (l) => l.startsWith(languageCode + "-"),
    );

    if (match) {
      return match;
    }
  }

  return null;
}

/**
 * Type guard to check if a string is a valid SupportedLanguage
 */
export function isValidSupportedLanguage(lang: string): lang is SupportedLanguage {
  const validLanguages = [
    "en-US",
    "pt-BR",
    "pt-PT",
    "es-ES",
    "fr-FR",
    "de-DE",
    "it-IT",
    "ja-JP",
    "zh-CN",
    "zh-TW",
    "ko-KR",
    "ar-SA",
    "he-IL",
    "fa-IR",
    "ur-PK",
  ];
  return validLanguages.includes(lang);
}
