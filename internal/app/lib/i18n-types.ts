export type Locale = 'pt' | 'en' | 'es' | 'it';

export interface TranslationEntry {
  one: string;
  other: string;
  [key: string]: string;
}

export interface Translations {
  [key: string]: TranslationEntry;
}
