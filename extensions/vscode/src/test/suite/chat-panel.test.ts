import * as assert from 'assert';
import * as vscode from 'vscode';
import { ChatViewProvider } from '../../chat-panel';
import { Client } from '../../client';

suite('Chat Panel Unit Test Suite', () => {

	test('ChatViewProvider should post messages to webview on client notifications', async () => {
		let lastPostedMessage: any;

		const mockWebview: any = {
			postMessage: (msg: any) => { lastPostedMessage = msg; return Promise.resolve(true); },
			onDidReceiveMessage: () => ({ dispose: () => {} }),
			options: {},
			asWebviewUri: (uri: vscode.Uri) => uri,
			cspSource: ''
		};

		const mockView: any = {
			webview: mockWebview,
			show: () => {},
			onDidDispose: () => ({ dispose: () => {} }),
			visible: true
		};

		const mockClient = new Client('Test', 'dummy');
		// Force the webview to be resolved
		const provider = new ChatViewProvider(mockClient, { extensionUri: vscode.Uri.file('/') } as any);
		
		provider.resolveWebviewView(mockView, {} as any, {} as any);

		// Simulate a session/update notification from the client
		const notification = {
			method: 'session/update',
			params: {
				sessionId: 'test',
				update: {
					sessionUpdate: 'agent_message_chunk',
					content: [{
						type: 'content',
						content: { type: 'text', text: 'Hello from mock' }
					}]
				}
			}
		};

		// Trigger handleNotification via client event
		(mockClient.onNotification as any).fire(notification);

		assert.ok(lastPostedMessage, 'Should have posted a message to the webview');
		assert.strictEqual(lastPostedMessage.type, 'agent_chunk');
		assert.strictEqual(lastPostedMessage.text, 'Hello from mock');
	});

	test('ChatViewProvider should handle tool_call notifications', async () => {
		let lastPostedMessage: any;
		const mockWebview: any = {
			postMessage: (msg: any) => { lastPostedMessage = msg; return Promise.resolve(true); },
			onDidReceiveMessage: () => ({ dispose: () => {} }),
			options: {}
		};
		const mockView: any = { webview: mockWebview, show: () => {} };
		const mockClient = new Client('Test', 'dummy');
		const provider = new ChatViewProvider(mockClient, { extensionUri: vscode.Uri.file('/') } as any);
		
		provider.resolveWebviewView(mockView, {} as any, {} as any);

		const notification = {
			method: 'session/update',
			params: {
				sessionId: 'test',
				update: {
					sessionUpdate: 'tool_call',
					toolCallId: 'call_1',
					title: 'Reading File',
					status: 'in_progress'
				}
			}
		};

		(mockClient.onNotification as any).fire(notification);

		assert.strictEqual(lastPostedMessage.type, 'tool_call');
		assert.strictEqual(lastPostedMessage.toolCallId, 'call_1');
		assert.strictEqual(lastPostedMessage.title, 'Reading File');
	});
});
