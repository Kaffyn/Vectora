/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        vscode: {
          bg: "var(--vscode-sideBar-background)",
          fg: "var(--vscode-sideBar-foreground)",
          inputBg: "var(--vscode-input-background)",
          inputFg: "var(--vscode-input-foreground)",
          border: "var(--vscode-input-border)",
          accent: "var(--vscode-button-background)",
          accentHover: "var(--vscode-button-hoverBackground)",
          accentFg: "var(--vscode-button-foreground)",
        },
      },
    },
  },
  plugins: [],
};
