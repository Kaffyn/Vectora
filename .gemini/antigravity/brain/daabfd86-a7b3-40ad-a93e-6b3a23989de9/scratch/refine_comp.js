const fs = require('fs');
const path = require('path');

const targetDir = 'C:\\Users\\bruno\\Desktop\\Vectora\\extensions\\vscode\\src';

function walk(dir) {
    const files = fs.readdirSync(dir);
    for (const file of files) {
        const fullPath = path.join(dir, file);
        if (fs.statSync(fullPath).isDirectory()) {
            walk(fullPath);
        } else if (fullPath.endsWith('.tsx') || fullPath.endsWith('.ts')) {
            refineFile(fullPath);
        }
    }
}

function refineFile(filePath) {
    let content = fs.readFileSync(filePath, 'utf8');
    let changed = false;

    // Replace react-i18next imports with our custom hook
    // Case 1: import { useTranslation, Trans } from 'react-i18next'
    if (content.includes('from "react-i18next"') || content.includes("from 'react-i18next'")) {
        content = content.replace(/import\s*{([^}]*)}\s*from\s*["']react-i18next["']/g, (match, p1) => {
            const imports = p1.trim().split(',').map(i => i.trim());
            return `import { ${imports.join(', ')} } from "../../hooks/useTranslation"`;
        });
        changed = true;
    }

    // Fix context imports if they are absolute or wrong
    if (content.includes('@src/context/ExtensionStateContext')) {
        content = content.replace(/@src\/context\/ExtensionStateContext/g, '../../context/ExtensionStateContext');
        changed = true;
    }

    if (changed) {
        fs.writeFileSync(filePath, content, 'utf8');
        console.log(`Refined: ${filePath}`);
    }
}

walk(targetDir);
console.log('Cleanup complete.');
