/**
 * RTL Utilities
 * Phase 11: RTL Languages Support & i18n Enhancement
 * Helper functions for RTL layout management
 */

import type { SupportedLanguage } from "../types";

/**
 * Convert LTR CSS properties to RTL equivalents
 */
export const rtlPropertyMap: Record<string, string> = {
  "margin-left": "margin-right",
  "margin-right": "margin-left",
  "padding-left": "padding-right",
  "padding-right": "padding-left",
  "border-left": "border-right",
  "border-right": "border-left",
  "border-left-color": "border-right-color",
  "border-right-color": "border-left-color",
  "border-left-width": "border-right-width",
  "border-right-width": "border-left-width",
  "border-left-style": "border-right-style",
  "border-right-style": "border-left-style",
  left: "right",
  right: "left",
  "text-align": "text-align", // Special: flips value instead
  float: "float", // Special: flips value instead
  clear: "clear", // Special: flips value instead
};

/**
 * CSS properties that should have their values flipped for RTL
 */
export const flipValueProperties = ["text-align", "float", "clear"];

/**
 * CSS property value flips
 */
export const flipValueMap: Record<string, string> = {
  left: "right",
  right: "left",
  "text-align:left": "text-align:right",
  "text-align:right": "text-align:left",
  "float:left": "float:right",
  "float:right": "float:left",
  "clear:left": "clear:right",
  "clear:right": "clear:left",
};

/**
 * Flip horizontal CSS values for RTL
 */
export function flipCSSValue(property: string, value: string): string {
  if (!value) return value;

  const map: Record<string, string> = {
    left: "right",
    right: "left",
  };

  if (property === "flex-direction") {
    if (value === "row") return "row-reverse";
    if (value === "row-reverse") return "row";
  }

  if (flipValueProperties.includes(property)) {
    const lowerValue = value.toLowerCase();
    return map[lowerValue] || value;
  }

  return value;
}

/**
 * Convert LTR styles to RTL styles
 */
export function flipStyles(
  styles: Record<string, string>,
  isRTL: boolean,
): Record<string, string> {
  if (!isRTL) return styles;

  const flipped: Record<string, string> = {};

  for (const [property, value] of Object.entries(styles)) {
    const mappedProperty = rtlPropertyMap[property] || property;
    const mappedValue = flipCSSValue(mappedProperty, value);
    flipped[mappedProperty] = mappedValue;
  }

  return flipped;
}

/**
 * Get direction-aware margin
 */
export function getDirectionalMargin(
  direction: "ltr" | "rtl",
  top: string,
  right: string,
  bottom: string,
  left: string,
): string {
  if (direction === "rtl") {
    return `${top} ${left} ${bottom} ${right}`;
  }
  return `${top} ${right} ${bottom} ${left}`;
}

/**
 * Get direction-aware padding
 */
export function getDirectionalPadding(
  direction: "ltr" | "rtl",
  top: string,
  right: string,
  bottom: string,
  left: string,
): string {
  if (direction === "rtl") {
    return `${top} ${left} ${bottom} ${right}`;
  }
  return `${top} ${right} ${bottom} ${left}`;
}

/**
 * Get direction-aware position
 */
export function getDirectionalPosition(
  direction: "ltr" | "rtl",
  position: "left" | "right",
  value: string,
): Record<string, string> {
  if (direction === "rtl") {
    const positionKey = position === "left" ? "right" : "left";
    return { [positionKey]: value };
  }
  return { [position]: value };
}

/**
 * Convert flex-direction based on RTL
 */
export function getFlexDirection(isRTL: boolean, baseDirection: string = "row"): string {
  if (isRTL && baseDirection === "row") {
    return "row-reverse";
  }
  if (isRTL && baseDirection === "column") {
    return "column"; // Columns don't flip
  }
  return baseDirection;
}

/**
 * Get text alignment based on direction
 */
export function getTextAlign(isRTL: boolean, align: string): string {
  if (!isRTL) return align;

  const alignMap: Record<string, string> = {
    left: "right",
    right: "left",
    start: "end",
    end: "start",
  };

  return alignMap[align] || align;
}

/**
 * Apply RTL transform to element
 * Flips the element horizontally
 */
export function applyRTLTransform(element: HTMLElement): void {
  element.style.transform = "scaleX(-1)";
}

/**
 * Remove RTL transform from element
 */
export function removeRTLTransform(element: HTMLElement): void {
  element.style.transform = "";
}

/**
 * Get inline styles for RTL/LTR
 */
export function getDirectionalStyles(
  isRTL: boolean,
  styles: Record<string, string>,
): string {
  const finalStyles = flipStyles(styles, isRTL);
  return Object.entries(finalStyles)
    .map(([key, value]) => `${key}:${value}`)
    .join(";");
}

/**
 * Create a CSS class for direction
 */
export function createDirectionalClass(
  baseName: string,
  direction: "ltr" | "rtl",
): string {
  return `${baseName}--${direction}`;
}

/**
 * Get opposite direction
 */
export function getOppositeDirection(direction: "ltr" | "rtl"): "ltr" | "rtl" {
  return direction === "ltr" ? "rtl" : "ltr";
}

/**
 * Validate if string contains RTL characters
 */
export function containsRTLCharacters(text: string): boolean {
  const rtlChars =
    /[\u0600-\u06FF\u0750-\u077F\u08A0-\u08FF\u0900-\u097F\u0590-\u05FF\uFB1D-\uFB4F]/g;
  return rtlChars.test(text);
}

/**
 * Detect text direction from content
 */
export function detectDirectionFromContent(text: string): "ltr" | "rtl" | "mixed" {
  if (!text) return "ltr";

  const rtlRegex = /[\u0600-\u06FF\u0750-\u077F\u08A0-\u08FF\u0590-\u05FF\uFB1D-\uFB4F]/g;
  const ltrRegex = /[a-zA-Z0-9]/g;

  const rtlMatches = text.match(rtlRegex);
  const ltrMatches = text.match(ltrRegex);

  const rtlCount = rtlMatches ? rtlMatches.length : 0;
  const ltrCount = ltrMatches ? ltrMatches.length : 0;

  if (rtlCount === 0) return "ltr";
  if (ltrCount === 0) return "rtl";
  return "mixed";
}

/**
 * Get logical CSS properties map
 * Maps traditional LTR properties to CSS logical properties
 */
export const logicalPropertyMap: Record<string, string> = {
  left: "inset-inline-start",
  right: "inset-inline-end",
  "margin-left": "margin-inline-start",
  "margin-right": "margin-inline-end",
  "padding-left": "padding-inline-start",
  "padding-right": "padding-inline-end",
  "border-left": "border-inline-start",
  "border-right": "border-inline-end",
  "border-left-color": "border-inline-start-color",
  "border-right-color": "border-inline-end-color",
  "border-left-width": "border-inline-start-width",
  "border-right-width": "border-inline-end-width",
  "border-left-style": "border-inline-start-style",
  "border-right-style": "border-inline-end-style",
  "border-top-left-radius": "border-start-start-radius",
  "border-top-right-radius": "border-start-end-radius",
  "border-bottom-left-radius": "border-end-start-radius",
  "border-bottom-right-radius": "border-end-end-radius",
  "text-align": "text-align", // No logical equivalent
  float: "float", // No logical equivalent
  clear: "clear", // No logical equivalent
};

/**
 * Convert to CSS logical properties
 */
export function convertToLogicalProperties(
  styles: Record<string, string>,
): Record<string, string> {
  const logical: Record<string, string> = {};

  for (const [property, value] of Object.entries(styles)) {
    const logicalProperty = logicalPropertyMap[property] || property;
    logical[logicalProperty] = value;
  }

  return logical;
}
