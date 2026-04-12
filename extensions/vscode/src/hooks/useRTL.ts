/**
 * useRTL Hook
 * Phase 11: RTL Languages Support & i18n Enhancement
 * React hook for RTL language management and direction detection
 */

import { useEffect, useState, useCallback, useContext } from "react";
import type { SupportedLanguage } from "../i18n/types";
import { TranslationContext } from "../i18n/TranslationContext";

/**
 * RTL state
 */
export interface RTLState {
  isRTL: boolean;
  direction: "ltr" | "rtl";
  language: SupportedLanguage;
  rtlClass: string;
  ltrClass: string;
}

/**
 * useRTL hook
 * Provides RTL state and utilities for React components
 */
export function useRTL(): RTLState & {
  shouldFlipIcon: (iconName: string) => boolean;
  getFlippedIcon: (iconName: string) => string;
  applyRTLClass: (element: HTMLElement) => void;
  removeRTLClass: (element: HTMLElement) => void;
} {
  const context = useContext(TranslationContext);
  const [rtlState, setRTLState] = useState<RTLState>({
    isRTL: false,
    direction: "ltr",
    language: "en-US",
    rtlClass: "rtl",
    ltrClass: "ltr",
  });

  useEffect(() => {
    if (!context) {
      console.warn("useRTL must be used within a TranslationProvider");
      return;
    }

    const updateRTLState = () => {
      const language = context.language;
      const isRTL = context.isRTL();
      const direction = isRTL ? "rtl" : "ltr";

      setRTLState({
        isRTL,
        direction,
        language,
        rtlClass: "rtl",
        ltrClass: "ltr",
      });

      // Update document direction
      if (typeof document !== "undefined") {
        document.documentElement.dir = direction;
        document.documentElement.lang = language;
        document.documentElement.classList.remove("rtl", "ltr");
        document.documentElement.classList.add(direction);
      }
    };

    // Initial update
    updateRTLState();

    // Listen for language changes
    const handleLanguageChange = () => {
      updateRTLState();
    };

    if (typeof window !== "undefined") {
      window.addEventListener("rtl-change", handleLanguageChange as EventListener);
    }

    return () => {
      if (typeof window !== "undefined") {
        window.removeEventListener("rtl-change", handleLanguageChange as EventListener);
      }
    };
  }, [context]);

  /**
   * Check if an icon should be flipped
   */
  const shouldFlipIcon = useCallback((iconName: string): boolean => {
    if (!rtlState.isRTL) {
      return false;
    }

    const flipIcons = [
      "arrow-left",
      "arrow-right",
      "chevron-left",
      "chevron-right",
      "angle-left",
      "angle-right",
      "caret-left",
      "caret-right",
    ];

    return flipIcons.some(
      (icon) =>
        iconName === icon ||
        iconName.includes(icon) ||
        iconName.toLowerCase().includes(icon),
    );
  }, [rtlState.isRTL]);

  /**
   * Get flipped icon name
   */
  const getFlippedIcon = useCallback((iconName: string): string => {
    if (!shouldFlipIcon(iconName)) {
      return iconName;
    }

    const flipMap: Record<string, string> = {
      "arrow-left": "arrow-right",
      "arrow-right": "arrow-left",
      "chevron-left": "chevron-right",
      "chevron-right": "chevron-left",
      "angle-left": "angle-right",
      "angle-right": "angle-left",
      "caret-left": "caret-right",
      "caret-right": "caret-left",
    };

    for (const [key, value] of Object.entries(flipMap)) {
      if (iconName === key) {
        return value;
      }
      if (iconName.includes(key)) {
        return iconName.replace(key, value);
      }
    }

    return iconName;
  }, [shouldFlipIcon]);

  /**
   * Apply RTL class to element
   */
  const applyRTLClass = useCallback((element: HTMLElement) => {
    if (!element) return;

    element.classList.remove(rtlState.rtlClass, rtlState.ltrClass);
    element.classList.add(rtlState.isRTL ? rtlState.rtlClass : rtlState.ltrClass);
    element.style.direction = rtlState.direction;
  }, [rtlState]);

  /**
   * Remove RTL class from element
   */
  const removeRTLClass = useCallback((element: HTMLElement) => {
    if (!element) return;

    element.classList.remove(rtlState.rtlClass, rtlState.ltrClass);
    element.style.direction = "";
  }, [rtlState.rtlClass, rtlState.ltrClass]);

  return {
    ...rtlState,
    shouldFlipIcon,
    getFlippedIcon,
    applyRTLClass,
    removeRTLClass,
  };
}

/**
 * useRTLStyles hook
 * Provides direction-aware CSS utilities
 */
export function useRTLStyles() {
  const { isRTL, direction } = useRTL();

  /**
   * Get directional margin
   */
  const getMargin = useCallback(
    (top: string, right: string, bottom: string, left: string): string => {
      if (isRTL) {
        return `${top} ${left} ${bottom} ${right}`;
      }
      return `${top} ${right} ${bottom} ${left}`;
    },
    [isRTL],
  );

  /**
   * Get directional padding
   */
  const getPadding = useCallback(
    (top: string, right: string, bottom: string, left: string): string => {
      if (isRTL) {
        return `${top} ${left} ${bottom} ${right}`;
      }
      return `${top} ${right} ${bottom} ${left}`;
    },
    [isRTL],
  );

  /**
   * Get directional position
   */
  const getPosition = useCallback(
    (side: "left" | "right", value: string): Record<string, string> => {
      if (isRTL) {
        const opposite = side === "left" ? "right" : "left";
        return { [opposite]: value };
      }
      return { [side]: value };
    },
    [isRTL],
  );

  /**
   * Get text alignment
   */
  const getTextAlign = useCallback(
    (align: string): string => {
      if (!isRTL) return align;

      const alignMap: Record<string, string> = {
        left: "right",
        right: "left",
        start: "end",
        end: "start",
      };

      return alignMap[align] || align;
    },
    [isRTL],
  );

  /**
   * Get flex direction
   */
  const getFlexDirection = useCallback(
    (baseDirection: string = "row"): string => {
      if (isRTL && baseDirection === "row") {
        return "row-reverse";
      }
      return baseDirection;
    },
    [isRTL],
  );

  return {
    isRTL,
    direction,
    getMargin,
    getPadding,
    getPosition,
    getTextAlign,
    getFlexDirection,
  };
}

/**
 * useLanguageDirection hook
 * Simpler hook for just getting direction
 */
export function useLanguageDirection(): {
  direction: "ltr" | "rtl";
  isRTL: boolean;
} {
  const { direction, isRTL } = useRTL();
  return { direction, isRTL };
}
