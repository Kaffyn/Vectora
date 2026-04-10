import * as path from 'path';
import Mocha from 'mocha';
import { glob } from 'glob';

export async function run(): Promise<void> {
	// Create the mocha test
	const mocha = new Mocha({
		ui: 'tdd',
		color: true
	});

	const testsRoot = __dirname;
	const files = await glob('**/*.test.js', { cwd: testsRoot });

	files.forEach(f => mocha.addFile(path.resolve(testsRoot, f)));

	try {
		// Run the mocha test
		return new Promise((c, e) => {
			mocha.run((failures: number) => {
				// Add 40s delay inside VS Code before returning control to test-electron
				console.log('[TEST] Tests completed. Waiting 40 seconds before closing VS Code...');
				setTimeout(() => {
					if (failures > 0) {
						e(new Error(`${failures} tests failed.`));
					} else {
						c();
					}
				}, 40000);
			});
		});
	} catch (err) {
		console.error(err);
		throw err;
	}
}
