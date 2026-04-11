const fs = require("fs");
const path = require("path");

const targetDir = path.join("extensions", "vscode", "src");

function walk(dir, callback) {
  fs.readdirSync(dir).forEach((f) => {
    let dirPath = path.join(dir, f);
    // SKIP roo-internal
    if (dirPath.includes("roo-internal")) return;

    let isDirectory = fs.statSync(dirPath).isDirectory();
    isDirectory ? walk(dirPath, callback) : callback(path.join(dir, f));
  });
}

const refinementRules = [
  {
    pattern: /import {[^}]+} from ["']react-i18next["']/g,
    replacement: 'import { useTranslation, Trans } from "@src/i18n/TranslationContext"',
  },
  {
    pattern: /import { t } from ["']i18next["']/g,
    replacement: 'import { useTranslation } from "@src/i18n/TranslationContext"',
  },
  {
    pattern: /import i18next from ["']i18next["']/g,
    replacement: 'import { i18n as i18next, useTranslation, Trans } from "@src/i18n/TranslationContext"',
  },
  {
    pattern: /from ["']\.\.\/\.\.\/context\/(.*)["']/g,
    replacement: 'from "@context/$1"',
  },
  {
    pattern: /from ["']\.\.\/context\/(.*)["']/g,
    replacement: 'from "@context/$1"',
  },
  {
    pattern: /from ["']\.\.\/\.\.\/hooks\/(.*)["']/g,
    replacement: 'from "@hooks/$1"',
  },
  {
    pattern: /from ["']\.\.\/hooks\/(.*)["']/g,
    replacement: 'from "@hooks/$1"',
  },
  {
    pattern: /from ["']\.\.\/\.\.\/utils\/(.*)["']/g,
    replacement: 'from "@utils/$1"',
  },
  {
    pattern: /from ["']\.\.\/utils\/(.*)["']/g,
    replacement: 'from "@utils/$1"',
  },
  {
    pattern: /from ["']\.\.\/\.\.\/shared\/(.*)["']/g,
    replacement: 'from "@shared/$1"',
  },
  {
    pattern: /from ["']\.\.\/shared\/(.*)["']/g,
    replacement: 'from "@shared/$1"',
  },
];

const deduplicateImports = (content) => {
  let lines = content.split("\n");
  let seen = new Set();
  let newLines = [];
  for (let line of lines) {
    let trimmed = line.trim();
    if (trimmed.startsWith('import { useTranslation, Trans } from "@src/i18n/TranslationContext"')) {
      if (seen.has("i18n-import")) continue;
      seen.add("i18n-import");
    }
    newLines.push(line);
  }
  return newLines.join("\n");
};

walk(targetDir, (filePath) => {
  if (filePath.endsWith(".tsx") || filePath.endsWith(".ts")) {
    let content = fs.readFileSync(filePath, "utf8");
    const original = content;

    refinementRules.forEach((rule) => {
      if (rule.pattern.test(content)) {
        content = content.replace(rule.pattern, rule.replacement);
      }
    });

    content = deduplicateImports(content);

    if (content !== original) {
      fs.writeFileSync(filePath, content);
      console.log(`Refined: ${filePath}`);
    }
  }
});

console.log("UNIFIED i18n & path refinement v9 (shim-safe) complete.");
