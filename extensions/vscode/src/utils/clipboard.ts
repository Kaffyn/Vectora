/**
 * Clipboard utility functions
 */

export interface UseClipboardReturn {
  copyWithFeedback: (text: string) => Promise<boolean>;
}

export function useCopyToClipboard(): UseClipboardReturn {
  const copyWithFeedback = async (text: string): Promise<boolean> => {
    try {
      await navigator.clipboard.writeText(text);
      return true;
    } catch (error) {
      console.error("Failed to copy to clipboard:", error);
      return false;
    }
  };

  return {
    copyWithFeedback,
  };
}
