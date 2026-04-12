/**
 * Chat Text Area Component
 * Placeholder para Phase 2-3 refactoring
 */

import React, { forwardRef } from "react";

export interface ChatTextAreaProps {
  onSubmit?: (text: string) => void;
  placeholder?: string;
  disabled?: boolean;
}

export const ChatTextArea = forwardRef<HTMLTextAreaElement, ChatTextAreaProps>(
  (props, ref) => {
    return (
      <textarea
        ref={ref}
        placeholder={props.placeholder || "Type a message..."}
        disabled={props.disabled}
        className="w-full p-2 border rounded"
      />
    );
  },
);

ChatTextArea.displayName = "ChatTextArea";
