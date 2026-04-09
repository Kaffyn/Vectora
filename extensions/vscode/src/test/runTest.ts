import * as path from 'path';
import * as os from 'os';
import { runTests } from '@vscode/test-electron';

async function main() {
	try {
        // Detect local VS Code path on Windows
        let vscodeExecutablePath: string | undefined = process.env.VECTORA_VSCODE_PATH;
        
        if (!vscodeExecutablePath && os.platform() === 'win32') {
            // Standard User Installation Path
            const localAppData = process.env.LOCALAPPDATA || path.join(os.homedir(), 'AppData', 'Local');
            const standardPath = path.join(localAppData, 'Programs', 'Microsoft VS Code', 'Code.exe');
            vscodeExecutablePath = standardPath;
        }

		// The folder containing the Extension Manifest package.json
		const extensionDevelopmentPath = path.resolve(__dirname, '../../');

		// The path to test runner
		const extensionTestsPath = path.resolve(__dirname, './suite/index');

		console.log(`Using VS Code executable: ${vscodeExecutablePath}`);

		// Run integration tests
		await runTests({ 
            vscodeExecutablePath,
            extensionDevelopmentPath, 
            extensionTestsPath,
            launchArgs: ['--disable-extensions']
        });

        // Add 10s delay after tests so the user can verify the state visually
        console.log('[TEST] Tests completed. Waiting 10 seconds before closing VS Code...');
        await new Promise(resolve => setTimeout(resolve, 10000));
	} catch (err) {
		console.error('Failed to run tests');
		process.exit(1);
	}
}

main();
