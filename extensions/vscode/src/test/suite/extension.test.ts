import * as assert from 'assert';
import * as vscode from 'vscode';

suite('Vectora Extension Test Suite', () => {
	vscode.window.showInformationMessage('Start all tests.');

	test('Extension should be present', () => {
		assert.ok(vscode.extensions.getExtension('kaffyn.vectora'));
	});

	test('Extension should activate', async () => {
		const extension = vscode.extensions.getExtension('kaffyn.vectora');
		await extension?.activate();
		assert.strictEqual(extension?.isActive, true);
	});

	test('Commands should be registered', async () => {
		const commands = await vscode.commands.getCommands(true);
		assert.ok(commands.includes('vectora.newSession'));
		assert.ok(commands.includes('vectora.explainCode'));
		assert.ok(commands.includes('vectora.refactorCode'));
	});

	test('Chat view should be registered', async () => {
		// We can't easily check if a webview is rendered, 
		// but we can check if the view is registered by attempting to focus it
		try {
			await vscode.commands.executeCommand('vectora.chatView.focus');
			assert.ok(true);
		} catch (e) {
			assert.fail('Chat view was not registered correctly');
		}
	});
});
