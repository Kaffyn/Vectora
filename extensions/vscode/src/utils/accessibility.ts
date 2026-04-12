/**
 * Accessibility utilities for WCAG 2.1 AA compliance
 */

/**
 * Check if text has sufficient contrast ratio with background
 * WCAG 2.1 AA requires 4.5:1 for normal text, 3:1 for large text
 */
export function checkContrast(foreground: string, background: string): number {
  const fgLuminance = getRelativeLuminance(foreground);
  const bgLuminance = getRelativeLuminance(background);

  const lighter = Math.max(fgLuminance, bgLuminance);
  const darker = Math.min(fgLuminance, bgLuminance);

  return (lighter + 0.05) / (darker + 0.05);
}

/**
 * Calculate relative luminance of a color
 * Used for contrast ratio calculation
 */
export function getRelativeLuminance(color: string): number {
  // Parse hex or rgb color
  let r, g, b;

  if (color.startsWith("#")) {
    const hex = color.slice(1);
    r = parseInt(hex.slice(0, 2), 16) / 255;
    g = parseInt(hex.slice(2, 4), 16) / 255;
    b = parseInt(hex.slice(4, 6), 16) / 255;
  } else if (color.startsWith("rgb")) {
    const matches = color.match(/\d+/g);
    if (matches) {
      r = parseInt(matches[0]) / 255;
      g = parseInt(matches[1]) / 255;
      b = parseInt(matches[2]) / 255;
    } else {
      return 0;
    }
  } else {
    return 0;
  }

  // Apply gamma correction
  r = r <= 0.03928 ? r / 12.92 : Math.pow((r + 0.055) / 1.055, 2.4);
  g = g <= 0.03928 ? g / 12.92 : Math.pow((g + 0.055) / 1.055, 2.4);
  b = b <= 0.03928 ? b / 12.92 : Math.pow((b + 0.055) / 1.055, 2.4);

  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/**
 * Generate unique ARIA ID
 */
export function generateAriaId(prefix = "aria"): string {
  return `${prefix}-${Math.random().toString(36).slice(2, 9)}`;
}

/**
 * Check if element is visible (for screen readers)
 */
export function isElementVisible(element: HTMLElement): boolean {
  return !!(
    element.offsetParent ||
    element.offsetWidth ||
    element.offsetHeight ||
    element.getClientRects().length
  );
}

/**
 * Trap focus within an element (for modals)
 */
export function trapFocus(element: HTMLElement, event: KeyboardEvent): void {
  if (event.key !== "Tab") return;

  const focusableElements = element.querySelectorAll(
    'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
  );

  const firstElement = focusableElements[0] as HTMLElement;
  const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement;
  const activeElement = document.activeElement as HTMLElement;

  if (event.shiftKey) {
    if (activeElement === firstElement) {
      lastElement.focus();
      event.preventDefault();
    }
  } else {
    if (activeElement === lastElement) {
      firstElement.focus();
      event.preventDefault();
    }
  }
}

/**
 * Announce message to screen readers
 */
export function announceToScreenReader(
  message: string,
  priority: "polite" | "assertive" = "polite",
): void {
  const announcement = document.createElement("div");
  announcement.setAttribute("role", "status");
  announcement.setAttribute("aria-live", priority);
  announcement.setAttribute("aria-atomic", "true");
  announcement.className = "sr-only";
  announcement.textContent = message;

  document.body.appendChild(announcement);

  // Remove after announcement
  setTimeout(() => announcement.remove(), 1000);
}

/**
 * Focus management utilities
 */
export const FocusManager = {
  /**
   * Save current focus
   */
  save(): HTMLElement | null {
    return document.activeElement as HTMLElement | null;
  },

  /**
   * Restore focus to element
   */
  restore(element: HTMLElement | null): void {
    if (element && typeof element.focus === "function") {
      element.focus();
    }
  },

  /**
   * Move focus to first focusable element in container
   */
  moveToFirst(container: HTMLElement): void {
    const firstFocusable = container.querySelector(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])',
    ) as HTMLElement;

    if (firstFocusable) {
      firstFocusable.focus();
    }
  },

  /**
   * Move focus to element with role
   */
  moveToRole(role: string, container?: HTMLElement): void {
    const target = (container || document).querySelector(`[role="${role}"]`) as HTMLElement;
    if (target) {
      target.focus();
    }
  },
};

/**
 * Check if user prefers reduced motion
 */
export function prefersReducedMotion(): boolean {
  return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}

/**
 * Check if user prefers dark mode
 */
export function prefersDarkMode(): boolean {
  return window.matchMedia("(prefers-color-scheme: dark)").matches;
}

/**
 * Validate ARIA attributes on element
 */
export function validateAriaAttributes(element: HTMLElement): string[] {
  const issues: string[] = [];

  // Check for aria-label or aria-labelledby
  const hasLabel = element.hasAttribute("aria-label") || element.hasAttribute("aria-labelledby");
  if (!hasLabel && (element.tagName === "BUTTON" || element.tagName === "INPUT")) {
    issues.push(`${element.tagName} missing aria-label or aria-labelledby`);
  }

  // Check for role="button" on non-button elements
  if (element.getAttribute("role") === "button" && element.tagName !== "BUTTON") {
    if (!element.hasAttribute("aria-pressed")) {
      issues.push("Non-button element with role=button missing aria-pressed");
    }
  }

  return issues;
}
