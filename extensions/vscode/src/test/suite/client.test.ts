import * as assert from 'assert';
import * as vscode from 'vscode';
import { Client } from '../../client';

suite('Client Unit Test Suite', () => {
	
	test('Client should correlate request IDs', async () => {
		// This is a unit test that would ideally mock child_process.spawn.
		// For now, let's test the logical properties of the Client class 
		// that don't depend on a real process, or use a dummy command.
		
		// In a real TDD flow, we would use a mock here. 
		// Since we want to proceed with Phase 2, I'll implement a TestClient 
		// that overrides sendRaw to simulate immediate responses.
		
		class MockClient extends Client {
			public lastSentMessage: any;
			
			protected override sendRaw(msg: any): void {
				this.lastSentMessage = msg;
				// Simulate internal process behavior by feeding data back to onData
				if (msg.id !== undefined) {
					const response = {
						jsonrpc: '2.0',
						id: msg.id,
						result: { status: 'ok', method: msg.method }
					};
					// We need to access private onData, but for tests we can use casting
					(this as any).onData(Buffer.from(JSON.stringify(response) + '\n'));
				}
			}
			
			// Override start to not actually spawn anything
			public override async start(): Promise<void> {
				(this as any)._isDisposed = false;
				(this as any).process = { 
					stdin: { writable: true, write: () => true },
					kill: () => true,
					on: () => true
				};
				return Promise.resolve();
			}
		}

		const client = new MockClient('Test', 'dummy');
		await client.start();

		const result = await client.request('test/method', { foo: 'bar' });
		assert.strictEqual(result.status, 'ok');
		assert.strictEqual(result.method, 'test/method');
		assert.strictEqual(client.lastSentMessage.method, 'test/method');
	});

	test('Client should handle timeouts', async () => {
		class TimeoutMockClient extends Client {
			protected override sendRaw(_msg: any): void {
				// Do nothing, simulate no response
			}
			public override async start(): Promise<void> {
				(this as any)._isDisposed = false;
				(this as any).process = { stdin: { writable: true } };
				return Promise.resolve();
			}
		}

		const client = new TimeoutMockClient('TimeoutTest', 'dummy');
		await client.start();

		try {
			await client.request('slow/method', {}, 100); // 100ms timeout
			assert.fail('Should have timed out');
		} catch (err: any) {
			assert.ok(err.message.includes('timed out'), `Expected timeout error, got: ${err.message}`);
		}
	});

	test('Client should emit notifications', (done) => {
		class NotificationMockClient extends Client {
			public triggerNotification(notif: any) {
				(this as any).onData(Buffer.from(JSON.stringify(notif) + '\n'));
			}
			public override async start(): Promise<void> {
				(this as any)._isDisposed = false;
				(this as any).process = { stdin: { writable: true } };
				return Promise.resolve();
			}
		}

		const client = new NotificationMockClient('NotifTest', 'dummy');
		client.onNotification.event((n) => {
			assert.strictEqual(n.method, 'test/event');
			assert.strictEqual(n.params.data, 'hello');
			done();
		});

		client.start().then(() => {
			client.triggerNotification({ jsonrpc: '2.0', method: 'test/event', params: { data: 'hello' } });
		});
	});
});
