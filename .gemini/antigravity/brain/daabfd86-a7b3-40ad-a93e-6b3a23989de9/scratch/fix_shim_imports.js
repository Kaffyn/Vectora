const fs = require("fs");
const path = require("path");

const targetDir = path.join("extensions", "vscode", "src", "roo-internal", "vscode-shim");

function walk(dir, callback) {
  if (!fs.existsSync(dir)) return;
  fs.readdirSync(dir).forEach((f) => {
    let dirPath = path.join(dir, f);
    let isDirectory = fs.statSync(dirPath).isDirectory();
    isDirectory ? walk(dirPath, callback) : callback(path.join(dir, f));
  });
}

walk(targetDir, (filePath) => {
  if (filePath.endsWith(".tsx") || filePath.endsWith(".ts")) {
    let content = fs.readFileSync(filePath, "utf8");
    let changed = false;

    // General alias to relative conversion for the shim
    const aliasToRelative = [
      { alias: "@context/", relative: "../context/" },
      { alias: "@utils/", relative: "../utils/" },
      { alias: "@classes/", relative: "../classes/" },
      { alias: "@interfaces/", relative: "../interfaces/" },
      { alias: "@api/", relative: "../api/" },
    ];

    aliasToRelative.forEach((rule) => {
      if (content.includes(rule.alias)) {
        // Determine depth of the file to adjust the number of ../
        // This is a bit tricky if files are at different levels.
        // Most shim files are in src/api/, src/context/, src/storage/, etc.
        // So they are at src/<folder>/<file>.
        // ../<folder>/ would be correct.
        content = content.split(rule.alias).join(rule.relative);
        changed = true;
      }
    });

    if (changed) {
      fs.writeFileSync(filePath, content);
      console.log(`Fixed shim import in: ${filePath}`);
    }
  }
});

console.log("Shim import repair complete.");
