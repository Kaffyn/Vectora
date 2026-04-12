import React, { forwardRef, InputHTMLAttributes } from "react";
import { useTranslator } from "../../hooks/useTranslation";

interface AccessibleInputProps extends InputHTMLAttributes<HTMLInputElement> {
  /**
   * Label text (for associated label element)
   */
  label?: string;
  /**
   * Helper text below input
   */
  helperText?: string;
  /**
   * Error message (makes input invalid)
   */
  error?: string;
  /**
   * Required field indicator
   */
  required?: boolean;
  /**
   * Input size: sm, md, lg
   */
  size?: "sm" | "md" | "lg";
}

/**
 * Accessible input component with WCAG 2.1 AA compliance
 * - Proper label association
 * - Error messaging
 * - Helper text
 * - Focus management
 */
export const AccessibleInput = forwardRef<HTMLInputElement, AccessibleInputProps>(
  (
    {
      id,
      label,
      helperText,
      error,
      required = false,
      size = "md",
      disabled = false,
      className = "",
      ...rest
    },
    ref,
  ) => {
    const translator = useTranslator();
    const inputId = React.useId();
    const helperId = React.useId();
    const errorId = React.useId();

    const finalInputId = id || inputId;

    // Combine classes
    const inputClass = [
      "accessible-input",
      `input-${size}`,
      error && "input-error",
      disabled && "input-disabled",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    // ARIA attributes for error state
    const ariaDescribedBy = [helperText && helperId, error && errorId].filter(Boolean).join(" ");

    return (
      <div className="accessible-input-wrapper">
        {label && (
          <label htmlFor={finalInputId} className="input-label">
            {label}
            {required && (
              <span className="input-required" aria-label={translator("common:required")}>
                *
              </span>
            )}
          </label>
        )}

        <input
          ref={ref}
          id={finalInputId}
          className={inputClass}
          disabled={disabled}
          required={required}
          aria-invalid={!!error}
          aria-describedby={ariaDescribedBy || undefined}
          {...rest}
        />

        {helperText && (
          <p id={helperId} className="input-helper-text">
            {helperText}
          </p>
        )}

        {error && (
          <p id={errorId} className="input-error-text" role="alert">
            {error}
          </p>
        )}
      </div>
    );
  },
);

AccessibleInput.displayName = "AccessibleInput";
