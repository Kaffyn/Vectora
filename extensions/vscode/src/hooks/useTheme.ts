import { useState, useEffect, useCallback } from "react";

type Theme = "light" | "dark" | "high-contrast" | "auto";

interface ThemeContextValue {
  theme: Theme;
  setTheme: (theme: Theme) => void;
  currentTheme: "light" | "dark" | "high-contrast";
}

/**
 * Hook for theme management
 * Supports: light, dark, high-contrast, auto (system preference)
 */
export function useTheme(): ThemeContextValue {
  const [theme, setThemeState] = useState<Theme>(() => {
    // Get from localStorage
    const stored = localStorage.getItem("vectora-theme");
    return (stored as Theme) || "auto";
  });

  const [systemTheme, setSystemTheme] = useState<"light" | "dark">(() => {
    if (window.matchMedia("(prefers-color-scheme: dark)").matches) {
      return "dark";
    }
    return "light";
  });

  // Detect system theme changes
  useEffect(() => {
    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const handleChange = (e: MediaQueryListEvent) => {
      setSystemTheme(e.matches ? "dark" : "light");
    };

    mediaQuery.addEventListener("change", handleChange);
    return () => mediaQuery.removeEventListener("change", handleChange);
  }, []);

  // Resolve current theme
  const resolveTheme = useCallback((): "light" | "dark" | "high-contrast" => {
    if (theme === "auto") {
      return systemTheme;
    }
    return theme as "light" | "dark" | "high-contrast";
  }, [theme, systemTheme]);

  const currentTheme = resolveTheme();

  // Apply theme to DOM
  useEffect(() => {
    const htmlElement = document.documentElement;
    htmlElement.setAttribute("data-theme", currentTheme);

    // Also set color-scheme for native elements
    document.documentElement.style.colorScheme =
      currentTheme === "dark" ? "dark" : "light";
  }, [currentTheme]);

  const setTheme = useCallback((newTheme: Theme) => {
    setThemeState(newTheme);
    localStorage.setItem("vectora-theme", newTheme);

    // Announce theme change to screen readers
    const themeNames: Record<Theme, string> = {
      light: "Light theme activated",
      dark: "Dark theme activated",
      "high-contrast": "High contrast theme activated",
      auto: "Automatic theme based on system preference",
    };

    const announcement = document.createElement("div");
    announcement.setAttribute("role", "status");
    announcement.setAttribute("aria-live", "polite");
    announcement.className = "sr-only";
    announcement.textContent = themeNames[newTheme];
    document.body.appendChild(announcement);
    setTimeout(() => announcement.remove(), 1000);
  }, []);

  return {
    theme,
    setTheme,
    currentTheme,
  };
}
