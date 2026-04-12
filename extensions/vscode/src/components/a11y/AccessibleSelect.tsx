import React, { forwardRef, SelectHTMLAttributes } from "react";

interface Option {
  value: string;
  label: string;
}

interface AccessibleSelectProps extends Omit<SelectHTMLAttributes<HTMLSelectElement>, "onChange"> {
  /**
   * Label text
   */
  label?: string;
  /**
   * Options for select
   */
  options: Option[];
  /**
   * Error message
   */
  error?: string;
  /**
   * Helper text
   */
  helperText?: string;
  /**
   * On change handler
   */
  onChange?: (value: string) => void;
  /**
   * Select size: sm, md, lg
   */
  size?: "sm" | "md" | "lg";
}

/**
 * Accessible select component with WCAG 2.1 AA compliance
 * - Proper label association
 * - Keyboard navigation
 * - ARIA attributes
 */
export const AccessibleSelect = forwardRef<HTMLSelectElement, AccessibleSelectProps>(
  (
    {
      id,
      label,
      options,
      error,
      helperText,
      disabled = false,
      required = false,
      size = "md",
      className = "",
      onChange,
      ...rest
    },
    ref,
  ) => {
    const selectId = React.useId();
    const helperId = React.useId();
    const errorId = React.useId();

    const finalSelectId = id || selectId;

    const selectClass = [
      "accessible-select",
      `select-${size}`,
      error && "select-error",
      disabled && "select-disabled",
      className,
    ]
      .filter(Boolean)
      .join(" ");

    const ariaDescribedBy = [helperText && helperId, error && errorId].filter(Boolean).join(" ");

    return (
      <div className="accessible-select-wrapper">
        {label && (
          <label htmlFor={finalSelectId} className="select-label">
            {label}
            {required && <span className="select-required">*</span>}
          </label>
        )}

        <select
          ref={ref}
          id={finalSelectId}
          className={selectClass}
          disabled={disabled}
          required={required}
          aria-invalid={!!error}
          aria-describedby={ariaDescribedBy || undefined}
          onChange={(e) => onChange?.(e.target.value)}
          {...rest}
        >
          <option value="">
            {label ? `Select ${label.toLowerCase()}...` : "Select..."}
          </option>
          {options.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>

        {helperText && (
          <p id={helperId} className="select-helper-text">
            {helperText}
          </p>
        )}

        {error && (
          <p id={errorId} className="select-error-text" role="alert">
            {error}
          </p>
        )}
      </div>
    );
  },
);

AccessibleSelect.displayName = "AccessibleSelect";
