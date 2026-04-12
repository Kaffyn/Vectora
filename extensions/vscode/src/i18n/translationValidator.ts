/**
 * Translation Validator
 * Validates translation files for completeness and consistency
 */

import { translationSchema, languageMetadata, LanguageCode } from "./translations-schema";

export interface ValidationError {
  type: "missing" | "extra" | "invalid" | "plural";
  key: string;
  language: string;
  message: string;
}

export interface ValidationResult {
  valid: boolean;
  errors: ValidationError[];
  warnings: ValidationError[];
  coverage: Record<string, number>; // Percentage coverage per language
  summary: {
    totalKeys: number;
    languages: LanguageCode[];
    missing: number;
    extra: number;
  };
}

/**
 * Flatten nested translation object
 */
function flattenKeys(obj: any, prefix = ""): Record<string, string> {
  const result: Record<string, string> = {};

  for (const key in obj) {
    const value = obj[key];
    const newKey = prefix ? `${prefix}.${key}` : key;

    if (typeof value === "string") {
      result[newKey] = value;
    } else if (typeof value === "object" && value !== null) {
      Object.assign(result, flattenKeys(value, newKey));
    }
  }

  return result;
}

/**
 * Get all keys from schema
 */
function getSchemaKeys(): string[] {
  const flattened = flattenKeys(translationSchema);
  return Object.keys(flattened);
}

/**
 * Validate a single translation object
 */
export function validateTranslation(
  translation: Record<string, any>,
  language: string,
): ValidationError[] {
  const errors: ValidationError[] = [];
  const schemaKeys = getSchemaKeys();
  const translationKeys = new Set(Object.keys(flattenKeys(translation)));

  // Check for missing keys
  for (const key of schemaKeys) {
    if (!translationKeys.has(key)) {
      errors.push({
        type: "missing",
        key,
        language,
        message: `Missing translation key: ${key}`,
      });
    }
  }

  // Check for extra keys
  for (const key of translationKeys) {
    if (!schemaKeys.includes(key)) {
      errors.push({
        type: "extra",
        key,
        language,
        message: `Extra translation key: ${key}`,
      });
    }
  }

  // Check for placeholder consistency
  for (const key of schemaKeys) {
    const schemaValue = getValueByKey(translationSchema, key);
    const translationValue = getValueByKey(translation, key);

    if (schemaValue && translationValue && typeof schemaValue === "string" && typeof translationValue === "string") {
      const schemaPlaceholders = extractPlaceholders(schemaValue);
      const translationPlaceholders = extractPlaceholders(translationValue);

      if (JSON.stringify(schemaPlaceholders) !== JSON.stringify(translationPlaceholders)) {
        errors.push({
          type: "invalid",
          key,
          language,
          message: `Placeholder mismatch in ${key}: expected {{${schemaPlaceholders.join(", ")}}}, got {{${translationPlaceholders.join(", ")}}}`,
        });
      }
    }
  }

  return errors;
}

/**
 * Extract placeholder variables from translation string
 */
function extractPlaceholders(text: string): string[] {
  const matches = text.match(/\{\{(\w+)\}\}/g) || [];
  return matches.map((m) => m.slice(2, -2)).sort();
}

/**
 * Get value from nested object using dot notation
 */
function getValueByKey(obj: any, path: string): any {
  return path.split(".").reduce((current, key) => current?.[key], obj);
}

/**
 * Validate all translations
 */
export function validateAllTranslations(
  translations: Record<LanguageCode, Record<string, any>>,
): ValidationResult {
  const allErrors: ValidationError[] = [];
  const allWarnings: ValidationError[] = [];
  const coverage: Record<string, number> = {};
  const schemaKeys = getSchemaKeys();
  const totalKeys = schemaKeys.length;

  // Validate each language
  for (const [language, translation] of Object.entries(translations)) {
    const errors = validateTranslation(translation, language);
    allErrors.push(...errors.filter((e) => e.type !== "extra"));
    allWarnings.push(...errors.filter((e) => e.type === "extra"));

    // Calculate coverage
    const missingCount = errors.filter((e) => e.type === "missing").length;
    coverage[language] = Math.round(((totalKeys - missingCount) / totalKeys) * 100);
  }

  // Check for consistency between languages
  for (const language of Object.keys(translations)) {
    const hasAllKeys = !allErrors.some((e) => e.type === "missing" && e.language === language);
    if (!hasAllKeys) {
      allWarnings.push({
        type: "invalid",
        key: "language-consistency",
        language,
        message: `Language ${language} is incomplete`,
      });
    }
  }

  const valid = allErrors.length === 0;

  return {
    valid,
    errors: allErrors,
    warnings: allWarnings,
    coverage,
    summary: {
      totalKeys,
      languages: Object.keys(translations) as LanguageCode[],
      missing: allErrors.filter((e) => e.type === "missing").length,
      extra: allWarnings.filter((e) => e.type === "extra").length,
    },
  };
}

/**
 * Generate missing translation template
 */
export function generateMissingTemplate(
  translation: Record<string, any>,
  targetLanguage: string,
): Record<string, any> {
  const schemaKeys = getSchemaKeys();
  const translationKeys = new Set(Object.keys(flattenKeys(translation)));
  const template: Record<string, any> = JSON.parse(JSON.stringify(translation));

  // Add missing keys with placeholder values
  for (const key of schemaKeys) {
    if (!translationKeys.has(key)) {
      const keys = key.split(".");
      let current = template;

      for (let i = 0; i < keys.length - 1; i++) {
        if (!current[keys[i]]) {
          current[keys[i]] = {};
        }
        current = current[keys[i]];
      }

      current[keys[keys.length - 1]] = `[${targetLanguage}: ${key}]`;
    }
  }

  return template;
}

/**
 * Get validation report as formatted string
 */
export function formatValidationReport(result: ValidationResult): string {
  let report = "=== Translation Validation Report ===\n\n";

  // Summary
  report += "Summary:\n";
  report += `  Languages: ${result.summary.languages.join(", ")}\n`;
  report += `  Total Keys: ${result.summary.totalKeys}\n`;
  report += `  Missing Keys: ${result.summary.missing}\n`;
  report += `  Extra Keys: ${result.summary.extra}\n`;
  report += `  Status: ${result.valid ? "✓ Valid" : "✗ Invalid"}\n\n`;

  // Coverage
  report += "Coverage:\n";
  for (const [language, percent] of Object.entries(result.coverage)) {
    const bar = "█".repeat(Math.round(percent / 5)) + "░".repeat(20 - Math.round(percent / 5));
    report += `  ${language}: [${bar}] ${percent}%\n`;
  }

  // Errors
  if (result.errors.length > 0) {
    report += "\nErrors:\n";
    for (const error of result.errors) {
      report += `  • [${error.language}] ${error.message}\n`;
    }
  }

  // Warnings
  if (result.warnings.length > 0) {
    report += "\nWarnings:\n";
    for (const warning of result.warnings) {
      report += `  ⚠ [${warning.language}] ${warning.message}\n`;
    }
  }

  return report;
}
