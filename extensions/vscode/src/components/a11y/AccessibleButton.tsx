import React, { forwardRef, ButtonHTMLAttributes } from "react";
import { useTranslator } from "../../hooks/useTranslation";
import "../../../styles/a11y.css";

interface AccessibleButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  /**
   * Button variant: primary, secondary, danger, etc.
   */
  variant?: "primary" | "secondary" | "danger" | "ghost";
  /**
   * Button size: sm, md, lg
   */
  size?: "sm" | "md" | "lg";
  /**
   * Loading state
   */
  isLoading?: boolean;
  /**
   * Disabled state
   */
  disabled?: boolean;
  /**
   * ARIA label for screen readers
   */
  ariaLabel?: string;
  /**
   * Show loading spinner
   */
  showSpinner?: boolean;
}

/**
 * Accessible button component with WCAG 2.1 AA compliance
 * - Proper focus management
 * - ARIA attributes
 * - Keyboard navigation
 * - Loading states
 */
export const AccessibleButton = forwardRef<HTMLButtonElement, AccessibleButtonProps>(
  (
    {
      variant = "primary",
      size = "md",
      isLoading = false,
      disabled = false,
      ariaLabel,
      showSpinner = true,
      className = "",
      children,
      onClick,
      onKeyDown,
      ...rest
    },
    ref,
  ) => {
    const translator = useTranslator();
    const buttonId = React.useId();

    // Combine classes for accessibility
    const buttonClass = [
      "accessible-button",
      `button-${variant}`,
      `button-${size}`,
      isLoading && "button-loading",
      disabled && "button-disabled",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    // Handle keyboard events
    const handleKeyDown = (e: React.KeyboardEvent<HTMLButtonElement>) => {
      // Space or Enter activates button
      if ((e.key === " " || e.key === "Enter") && !disabled && !isLoading) {
        e.preventDefault();
        const mouseEvent = new MouseEvent("click", {
          bubbles: true,
          cancelable: true,
          view: window,
        });
        onClick?.(mouseEvent as any);
      }
      onKeyDown?.(e);
    };

    // Loading text for screen readers
    const loadingText = translator("common:loading");

    return (
      <button
        ref={ref}
        id={buttonId}
        type="button"
        className={buttonClass}
        disabled={disabled || isLoading}
        aria-disabled={disabled || isLoading}
        aria-label={ariaLabel || (typeof children === "string" ? children : undefined)}
        aria-busy={isLoading}
        onKeyDown={handleKeyDown}
        {...rest}
      >
        {isLoading && showSpinner && (
          <span className="button-spinner" aria-hidden="true">
            ⏳
          </span>
        )}
        <span className="button-content">{isLoading ? `${loadingText}...` : children}</span>
      </button>
    );
  },
);

AccessibleButton.displayName = "AccessibleButton";
