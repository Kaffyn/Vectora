import fs from 'fs';
import path from 'path';
import { Locale, Translations } from './i18n-types';

/**
 * Loads and parses the specific i18n CSV file for a locale.
 * Supports: key,one,other (and zero,few,many columns if added).
 */
export function getTranslations(locale: Locale): Translations {
  // 1. Get English as the absolute baseline for all languages (Fallback System)
  const englishBaseline = internal_get_all("en");
  
  if (locale === "en") return englishBaseline;

  // 2. Get the target translations and merge atop English
  const targetTranslations = internal_get_all(locale);
  
  return {
    ...englishBaseline,
    ...targetTranslations,
  };
}

function internal_get_all(locale: Locale): Translations {
  // Use process.cwd() as it points to 'src/web' during runtime execution
  const filePath = path.resolve(process.cwd(), 'locales', `${locale}.csv`);
  
  if (!fs.existsSync(filePath)) {
    // Attempt debug log to find where it is looking
    console.warn(`[i18n Info] Looking for ${locale}.csv in: ${filePath}`);
    return {};
  }

  try {
    const csvContent = fs.readFileSync(filePath, 'utf-8');
    return parseCSV(csvContent);
  } catch (err) {
    console.error(`[i18n Error] Failed to read ${filePath}:`, err);
    return {};
  }
}

function parseCSV(csvContent: string): Translations {
  const lines = csvContent.split('\n').filter(line => line.trim() !== '');
  const headers = lines[0]?.split(',').map(h => h.trim()) || [];
  
  const translations: Translations = {};
  
  for (let i = 1; i < lines.length; i++) {
    const columns = lines[i]?.split(',').map(c => c.trim());
    if (columns && columns.length > 0) {
      const key = columns[0];
      if (key) {
        translations[key] = {};
        headers.forEach((header, index) => {
          if (index > 0 && columns[index] !== undefined) {
            // @ts-ignore
            translations[key][header] = columns[index];
          }
        });
      }
    }
  }
  
  return translations;
}

/**
 * Advanced t() function with pluralization support.
 */
export function createTranslate(locale: Locale = 'pt') {
  const tArr = getTranslations(locale);
  const pluralRules = new Intl.PluralRules(locale === 'en' ? 'en-US' : locale === 'pt' ? 'pt-BR' : locale);

  return (key: string, count?: number, variables?: Record<string, string | number>) => {
    const entry = tArr[key] || { one: key, other: key };
    
    let result: string;
    if (count !== undefined) {
      const pluralKey = pluralRules.select(count);
      // Fallback from specific key ('few') to 'other'
      // @ts-ignore
      result = entry[pluralKey] || entry.other || entry.one || key;
    } else {
      result = entry.one || entry.other || key;
    }

    if (variables) {
      Object.entries(variables).forEach(([k, v]) => {
        result = result.replace(new RegExp(`{${k}}`, 'g'), String(v));
      });
    }
    
    // Always support simple {} for count variable as convenience
    if (count !== undefined) {
      result = result.replace('{}', String(count));
    }
    
    return result;
  };
}
