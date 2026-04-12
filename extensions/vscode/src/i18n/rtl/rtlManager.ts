/**
 * RTL Manager
 * Phase 11: RTL Languages Support & i18n Enhancement
 * Manages RTL state, styles, and direction for the application
 */

import type { SupportedLanguage, RTLConfig, RTLAPI } from "../types";
import { isRTLLanguage, getTextDirection } from "./rtlDetector";

/**
 * RTL Manager implementation
 */
export class RTLManager implements RTLAPI {
  private currentLanguage: SupportedLanguage;
  private config: Required<RTLConfig>;
  private rootElement: HTMLElement | null;

  constructor(language: SupportedLanguage, config: RTLConfig = {}) {
    this.currentLanguage = language;
    this.config = {
      enabled: config.enabled ?? true,
      rtlClass: config.rtlClass || "rtl",
      ltrClass: config.ltrClass || "ltr",
      useCSSLogicalProperties: config.useCSSLogicalProperties ?? true,
      autoFlipLayout: config.autoFlipLayout ?? true,
      flipIcons: config.flipIcons ?? true,
    };
    this.rootElement = typeof document !== "undefined" ? document.documentElement : null;
    this.initialize();
  }

  /**
   * Initialize RTL manager
   */
  private initialize(): void {
    if (!this.config.enabled) {
      return;
    }

    this.applyDirection(this.currentLanguage);
  }

  /**
   * Get current text direction
   */
  get direction(): "ltr" | "rtl" {
    return getTextDirection(this.currentLanguage);
  }

  /**
   * Check if current language is RTL
   */
  isRTL(language?: SupportedLanguage): boolean {
    const lang = language || this.currentLanguage;
    return isRTLLanguage(lang);
  }

  /**
   * Get direction class for a language
   */
  getDirectionClass(language?: SupportedLanguage): string {
    const lang = language || this.currentLanguage;
    return this.isRTL(lang) ? this.config.rtlClass : this.config.ltrClass;
  }

  /**
   * Apply RTL/LTR direction to application
   */
  applyDirection(language: SupportedLanguage): void {
    if (!this.config.enabled || !this.rootElement) {
      return;
    }

    this.currentLanguage = language;
    const direction = this.direction;

    // Set dir attribute
    this.rootElement.setAttribute("dir", direction);

    // Set lang attribute
    this.rootElement.setAttribute("lang", language);

    // Set/remove CSS classes
    this.rootElement.classList.remove(this.config.rtlClass, this.config.ltrClass);
    this.rootElement.classList.add(this.getDirectionClass());

    // Apply CSS variable for direction
    this.rootElement.style.setProperty("--text-direction", direction);

    // Apply CSS for logical properties if available
    if (this.config.useCSSLogicalProperties) {
      this.applyLogicalProperties(direction);
    }

    // Dispatch custom event
    if (typeof window !== "undefined") {
      window.dispatchEvent(
        new CustomEvent("rtl-change", {
          detail: {
            language,
            direction,
            rtl: this.isRTL(language),
          },
        }),
      );
    }
  }

  /**
   * Apply CSS logical properties
   */
  private applyLogicalProperties(direction: "ltr" | "rtl"): void {
    if (!this.rootElement) {
      return;
    }

    // Set CSS variables for logical properties
    if (direction === "rtl") {
      this.rootElement.style.setProperty("--start", "right");
      this.rootElement.style.setProperty("--end", "left");
      this.rootElement.style.setProperty("--flex-direction", "row-reverse");
      this.rootElement.style.setProperty("--text-align-start", "right");
      this.rootElement.style.setProperty("--text-align-end", "left");
    } else {
      this.rootElement.style.setProperty("--start", "left");
      this.rootElement.style.setProperty("--end", "right");
      this.rootElement.style.setProperty("--flex-direction", "row");
      this.rootElement.style.setProperty("--text-align-start", "left");
      this.rootElement.style.setProperty("--text-align-end", "right");
    }
  }

  /**
   * Apply RTL styles to a specific element
   */
  applyRTLStyles(element: HTMLElement): void {
    if (!this.config.enabled) {
      return;
    }

    if (this.isRTL()) {
      element.classList.add(this.config.rtlClass);
      element.classList.remove(this.config.ltrClass);
      element.style.direction = "rtl";
    } else {
      element.classList.add(this.config.ltrClass);
      element.classList.remove(this.config.rtlClass);
      element.style.direction = "ltr";
    }
  }

  /**
   * Remove RTL styles from an element
   */
  removeRTLStyles(element: HTMLElement): void {
    element.classList.remove(this.config.rtlClass, this.config.ltrClass);
    element.style.direction = "";
  }

  /**
   * Check if an icon should be flipped for RTL
   */
  shouldFlipIcon(iconName: string): boolean {
    if (!this.config.flipIcons || !this.isRTL()) {
      return false;
    }

    // Icons that should be flipped in RTL
    const flipIcons = [
      "arrow-left",
      "arrow-right",
      "chevron-left",
      "chevron-right",
      "angle-left",
      "angle-right",
      "arrow-from-left",
      "arrow-from-right",
      "arrow-to-left",
      "arrow-to-right",
      "caret-left",
      "caret-right",
    ];

    return flipIcons.some(
      (icon) =>
        iconName === icon ||
        iconName.includes(icon) ||
        iconName.toLowerCase().includes(icon),
    );
  }

  /**
   * Get flipped icon name for RTL
   */
  getFlippedIconName(iconName: string): string {
    if (!this.shouldFlipIcon(iconName)) {
      return iconName;
    }

    // Icon flip mappings
    const flipMap: Record<string, string> = {
      "arrow-left": "arrow-right",
      "arrow-right": "arrow-left",
      "chevron-left": "chevron-right",
      "chevron-right": "chevron-left",
      "angle-left": "angle-right",
      "angle-right": "angle-left",
      "arrow-from-left": "arrow-from-right",
      "arrow-from-right": "arrow-from-left",
      "arrow-to-left": "arrow-to-right",
      "arrow-to-right": "arrow-to-left",
      "caret-left": "caret-right",
      "caret-right": "caret-left",
    };

    // Find and replace
    for (const [key, value] of Object.entries(flipMap)) {
      if (iconName === key) {
        return value;
      }
      if (iconName.includes(key)) {
        return iconName.replace(key, value);
      }
    }

    return iconName;
  }
}

/**
 * Global RTL manager instance
 */
let globalRTLManager: RTLManager | null = null;

/**
 * Initialize RTL manager
 */
export function initRTLManager(language: SupportedLanguage, config?: RTLConfig): RTLAPI {
  globalRTLManager = new RTLManager(language, config);
  return globalRTLManager;
}

/**
 * Get RTL manager instance
 */
export function getRTLManager(): RTLManager {
  if (!globalRTLManager) {
    throw new Error("RTL manager not initialized. Call initRTLManager() first.");
  }
  return globalRTLManager;
}
