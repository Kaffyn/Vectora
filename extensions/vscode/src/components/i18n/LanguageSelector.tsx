/**
 * Language Selector Component
 * Phase 11: RTL Languages Support & i18n Enhancement
 * Enhanced language selector with RTL support and language metadata
 */

import React, { useCallback, useMemo } from "react";
import type { SupportedLanguage, LanguageMetadata } from "../../i18n/types";
import { useRTL } from "../../hooks/useRTL";
import "../../../styles/rtl/rtl-base.css";

interface LanguageSelectorProps {
  currentLanguage: SupportedLanguage;
  availableLanguages: SupportedLanguage[];
  languageMetadata?: Record<SupportedLanguage, LanguageMetadata>;
  onLanguageChange: (language: SupportedLanguage) => Promise<void>;
  loading?: boolean;
  showRegion?: boolean;
  showDirection?: boolean;
  compact?: boolean;
  className?: string;
}

/**
 * Language metadata database
 */
const DEFAULT_LANGUAGE_METADATA: Record<SupportedLanguage, LanguageMetadata> = {
  "en-US": {
    code: "en-US",
    nativeName: "English",
    englishName: "English",
    region: "United States",
    direction: "ltr",
    dateFormat: "MM/DD/YYYY",
    timeFormat: "12h",
    numberFormat: { decimal: ".", thousands: "," },
  },
  "pt-BR": {
    code: "pt-BR",
    nativeName: "Português",
    englishName: "Portuguese",
    region: "Brazil",
    direction: "ltr",
    dateFormat: "DD/MM/YYYY",
    timeFormat: "24h",
    numberFormat: { decimal: ",", thousands: "." },
  },
  "pt-PT": {
    code: "pt-PT",
    nativeName: "Português",
    englishName: "Portuguese",
    region: "Portugal",
    direction: "ltr",
    dateFormat: "DD/MM/YYYY",
    timeFormat: "24h",
    numberFormat: { decimal: ",", thousands: "." },
  },
  "es-ES": {
    code: "es-ES",
    nativeName: "Español",
    englishName: "Spanish",
    region: "Spain",
    direction: "ltr",
    dateFormat: "DD/MM/YYYY",
    timeFormat: "24h",
    numberFormat: { decimal: ",", thousands: "." },
  },
  "fr-FR": {
    code: "fr-FR",
    nativeName: "Français",
    englishName: "French",
    region: "France",
    direction: "ltr",
    dateFormat: "DD/MM/YYYY",
    timeFormat: "24h",
    numberFormat: { decimal: ",", thousands: " " },
  },
  "de-DE": {
    code: "de-DE",
    nativeName: "Deutsch",
    englishName: "German",
    region: "Germany",
    direction: "ltr",
    dateFormat: "DD.MM.YYYY",
    timeFormat: "24h",
    numberFormat: { decimal: ",", thousands: "." },
  },
  "it-IT": {
    code: "it-IT",
    nativeName: "Italiano",
    englishName: "Italian",
    region: "Italy",
    direction: "ltr",
    dateFormat: "DD/MM/YYYY",
    timeFormat: "24h",
    numberFormat: { decimal: ",", thousands: "." },
  },
  "ja-JP": {
    code: "ja-JP",
    nativeName: "日本語",
    englishName: "Japanese",
    region: "Japan",
    direction: "ltr",
    dateFormat: "YYYY/MM/DD",
    timeFormat: "24h",
    numberFormat: { decimal: ".", thousands: "," },
  },
  "zh-CN": {
    code: "zh-CN",
    nativeName: "简体中文",
    englishName: "Chinese",
    region: "China",
    direction: "ltr",
    dateFormat: "YYYY-MM-DD",
    timeFormat: "24h",
    numberFormat: { decimal: ".", thousands: "," },
  },
  "zh-TW": {
    code: "zh-TW",
    nativeName: "繁體中文",
    englishName: "Chinese",
    region: "Taiwan",
    direction: "ltr",
    dateFormat: "YYYY-MM-DD",
    timeFormat: "24h",
    numberFormat: { decimal: ".", thousands: "," },
  },
  "ko-KR": {
    code: "ko-KR",
    nativeName: "한국어",
    englishName: "Korean",
    region: "South Korea",
    direction: "ltr",
    dateFormat: "YYYY-MM-DD",
    timeFormat: "24h",
    numberFormat: { decimal: ".", thousands: "," },
  },
  "ar-SA": {
    code: "ar-SA",
    nativeName: "العربية",
    englishName: "Arabic",
    region: "Saudi Arabia",
    direction: "rtl",
    dateFormat: "DD/MM/YYYY",
    timeFormat: "24h",
    numberFormat: { decimal: ",", thousands: "." },
  },
  "he-IL": {
    code: "he-IL",
    nativeName: "עברית",
    englishName: "Hebrew",
    region: "Israel",
    direction: "rtl",
    dateFormat: "DD.MM.YYYY",
    timeFormat: "24h",
    numberFormat: { decimal: ".", thousands: "," },
  },
  "fa-IR": {
    code: "fa-IR",
    nativeName: "فارسی",
    englishName: "Persian",
    region: "Iran",
    direction: "rtl",
    dateFormat: "YYYY/MM/DD",
    timeFormat: "24h",
    numberFormat: { decimal: "/", thousands: "," },
  },
  "ur-PK": {
    code: "ur-PK",
    nativeName: "اردو",
    englishName: "Urdu",
    region: "Pakistan",
    direction: "rtl",
    dateFormat: "DD/MM/YYYY",
    timeFormat: "24h",
    numberFormat: { decimal: ".", thousands: "," },
  },
};

/**
 * Language Selector Component
 */
export const LanguageSelector: React.FC<LanguageSelectorProps> = ({
  currentLanguage,
  availableLanguages,
  languageMetadata = DEFAULT_LANGUAGE_METADATA,
  onLanguageChange,
  loading = false,
  showRegion = true,
  showDirection = false,
  compact = false,
  className = "",
}) => {
  const { direction, isRTL } = useRTL();

  const handleLanguageChange = useCallback(
    async (e: React.ChangeEvent<HTMLSelectElement>) => {
      const lang = e.target.value as SupportedLanguage;
      await onLanguageChange(lang);
    },
    [onLanguageChange],
  );

  const sortedLanguages = useMemo(() => {
    return availableLanguages.sort((a, b) => {
      const metaA = languageMetadata[a];
      const metaB = languageMetadata[b];
      return (metaA?.nativeName || a).localeCompare(metaB?.nativeName || b);
    });
  }, [availableLanguages, languageMetadata]);

  const currentMetadata = languageMetadata[currentLanguage];

  return (
    <div
      className={`language-selector ${className} ${direction}`}
      dir={direction}
      style={{
        display: "flex",
        alignItems: "center",
        gap: isRTL ? "var(--spacing-start, 1rem)" : "var(--spacing-end, 1rem)",
        flexDirection: isRTL ? "row-reverse" : "row",
      }}
    >
      {/* Language Label */}
      {!compact && (
        <label
          htmlFor="language-select"
          style={{
            textAlign: isRTL ? "right" : "left",
            marginInlineEnd: isRTL ? "0" : "0.5rem",
            marginInlineStart: isRTL ? "0.5rem" : "0",
          }}
        >
          Language
        </label>
      )}

      {/* Language Select */}
      <select
        id="language-select"
        value={currentLanguage}
        onChange={handleLanguageChange}
        disabled={loading}
        style={{
          direction: currentMetadata?.direction || "ltr",
          textAlign: "center",
          paddingInlineEnd: isRTL ? "2rem" : "0.5rem",
          paddingInlineStart: isRTL ? "0.5rem" : "2rem",
          minWidth: compact ? "auto" : "150px",
        }}
        aria-label="Select language"
        aria-busy={loading}
        title={`Current language: ${currentMetadata?.nativeName}`}
      >
        {sortedLanguages.map((lang) => {
          const meta = languageMetadata[lang];
          const label = compact
            ? meta?.nativeName || lang
            : `${meta?.nativeName || lang}${showRegion && meta?.region ? ` (${meta.region})` : ""}${showDirection && meta?.direction ? ` [${meta.direction.toUpperCase()}]` : ""}`;

          return (
            <option key={lang} value={lang}>
              {label}
            </option>
          );
        })}
      </select>

      {/* Direction Badge */}
      {showDirection && currentMetadata && (
        <span
          style={{
            fontSize: "0.85rem",
            opacity: 0.7,
            marginInlineStart: isRTL ? "0" : "0.5rem",
            marginInlineEnd: isRTL ? "0.5rem" : "0",
            padding: "0.25rem 0.5rem",
            borderRadius: "4px",
            backgroundColor: currentMetadata.direction === "rtl" ? "#ff6b6b" : "#4ecdc4",
            color: "white",
          }}
          title={`Text direction: ${currentMetadata.direction.toUpperCase()}`}
        >
          {currentMetadata.direction.toUpperCase()}
        </span>
      )}

      {/* Loading Indicator */}
      {loading && (
        <span
          style={{
            marginInlineStart: isRTL ? "0" : "0.5rem",
            marginInlineEnd: isRTL ? "0.5rem" : "0",
            animation: "spin 1s linear infinite",
          }}
          aria-label="Loading"
        >
          ⏳
        </span>
      )}
    </div>
  );
};

/**
 * Language Info Component
 * Displays detailed language information
 */
export const LanguageInfo: React.FC<{
  language: SupportedLanguage;
  metadata?: LanguageMetadata;
  showAll?: boolean;
}> = ({ language, metadata = DEFAULT_LANGUAGE_METADATA[language], showAll = false }) => {
  const { direction } = useRTL();

  if (!metadata) {
    return <div>Unknown language: {language}</div>;
  }

  return (
    <div
      className="language-info"
      dir={direction}
      style={{
        padding: "1rem",
        borderRadius: "8px",
        backgroundColor: "var(--bg-secondary, #f5f5f5)",
        textAlign: direction === "rtl" ? "right" : "left",
      }}
    >
      <div style={{ marginBottom: "0.5rem" }}>
        <strong>{metadata.nativeName}</strong> ({metadata.englishName})
      </div>
      <div style={{ fontSize: "0.9rem", opacity: 0.8 }}>
        Region: {metadata.region}
      </div>
      <div style={{ fontSize: "0.9rem", opacity: 0.8 }}>
        Direction: {metadata.direction.toUpperCase()}
      </div>

      {showAll && (
        <>
          <div style={{ fontSize: "0.85rem", marginTop: "0.5rem", opacity: 0.7 }}>
            Date format: {metadata.dateFormat}
          </div>
          <div style={{ fontSize: "0.85rem", opacity: 0.7 }}>
            Time format: {metadata.timeFormat}
          </div>
          <div style={{ fontSize: "0.85rem", opacity: 0.7 }}>
            Number format: Decimal: {metadata.numberFormat.decimal}, Thousands:{" "}
            {metadata.numberFormat.thousands}
          </div>
        </>
      )}
    </div>
  );
};

export default LanguageSelector;
