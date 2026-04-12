import React from "react";

interface SkipLinkProps {
  /**
   * Target ID to skip to
   */
  targetId: string;
  /**
   * Link text
   */
  text?: string;
}

/**
 * Skip link for keyboard navigation (WCAG 2.1 A)
 * Allows users to skip to main content
 */
export const SkipLink: React.FC<SkipLinkProps> = ({ targetId, text = "Skip to main content" }) => {
  return (
    <a href={`#${targetId}`} className="skip-link">
      {text}
    </a>
  );
};
