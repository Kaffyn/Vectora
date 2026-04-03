import React from "react";
import HomeClient from "./home-client";
import { getTranslations } from "../lib/i18n";
import type { Locale } from "../lib/i18n";

/**
 * Server Component: Zyris Dashboard Root.
 * We render a static baseline for Wails/Desktop.
 * Client-side hydration (home-client.tsx) handles dynamic i18n detection from cookies.
 */
export default async function Home() {
  const locale: Locale = "en";
  const translations = getTranslations(locale);

  return (
    <HomeClient initialTranslations={translations} initialLocale={locale} />
  );
}
