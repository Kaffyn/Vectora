export type Locale = 'pt' | 'en' | 'es' | 'it';

export interface TranslationEntry {
  one: string;
  other: string;
}

export interface Translations {
  [key: string]: TranslationEntry;
}
