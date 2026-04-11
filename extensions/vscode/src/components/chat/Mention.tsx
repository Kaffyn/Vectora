import React from "react";

export const Mention = ({ text }: { text: string }) => (
  <span className="text-vscode-linkForeground cursor-pointer hover:underline">
    {text}
  </span>
);
