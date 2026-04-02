import React from "react";
import HomeClient from "./home-client";
import { getTranslations, Locale } from "../lib/i18n";
import { headers, cookies } from "next/headers";

/**
 * Server Component: Vectora Dashboard Root.
 * Detects locale from headers and serves the hydration payload to HomeClient.
 */
export default async function Home() {
  const cookieStore = await cookies();
  const savedLocale = cookieStore.get("vectora_locale")?.value as Locale;
  
  let locale: Locale = "pt";

  if (savedLocale) {
    locale = savedLocale;
  } else {
    const headersList = await headers();
    const acceptLanguage = headersList.get("accept-language") || "pt";
    
    if (acceptLanguage.includes("en")) locale = "en";
    else if (acceptLanguage.includes("es")) locale = "es";
    else if (acceptLanguage.includes("it")) locale = "it";
  }

  const translations = getTranslations(locale);

  return (
    <HomeClient initialTranslations={translations} initialLocale={locale} />
  );
}
