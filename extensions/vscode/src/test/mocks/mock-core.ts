import * as readline from 'readline';

/**
 * Simple ACP Mock Server for E2E testing.
 * Responds to basic JSON-RPC 2.0 messages via Stdio.
 */
const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
    terminal: false
});

rl.on('line', (line) => {
    try {
        const msg = JSON.parse(line);
        if (msg.method === 'initialize') {
            sendResponse(msg.id, {
                protocolVersion: 1,
                agentCapabilities: { loadSession: true },
                agentInfo: { name: 'mock-core', title: 'Mock Core', version: '1.0.0' }
            });
        } else if (msg.method === 'session/new') {
            sendResponse(msg.id, { sessionId: 'mock_session_123' });
        } else if (msg.method === 'session/prompt') {
            // First respond to the prompt request
            sendResponse(msg.id, { stopReason: 'end_turn' });
            
            // Then send a session/update notification with a chunk
            sendNotification('session/update', {
                sessionId: 'mock_session_123',
                update: {
                    sessionUpdate: 'agent_message_chunk',
                    content: [{
                        type: 'content',
                        content: { type: 'text', text: 'Mock response from E2E script' }
                    }]
                }
            });
        }
    } catch (err) {
        // Silent fail
    }
});

function sendResponse(id: any, result: any) {
    console.log(JSON.stringify({
        jsonrpc: '2.0',
        id: id,
        result: result
    }));
}

function sendNotification(method: string, params: any) {
    console.log(JSON.stringify({
        jsonrpc: '2.0',
        method: method,
        params: params
    }));
}
