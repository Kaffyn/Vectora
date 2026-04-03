import net from 'node:net';

/**
 * Specialized IPC client for establishing connection with the Vectora Pipe.
 * On Windows it uses Named Pipes (\\.\pipe\vectora).
 * On Unix it uses Sockets (/tmp/vectora.sock).
 */

const PIPE_PATH = process.platform === 'win32' 
  ? '\\\\.\\pipe\\vectora' 
  : (process.env.VECTORA_SOCKET || '/tmp/vectora.sock');

const FRAME_DELIMITER = '\n';

export interface IPCRequest {
  id: string;
  type: string;
  method: string;
  payload: any;
}

export interface IPCResponse {
  id: string;
  type: string;
  payload?: any;
  error?: {
    code: string;
    message: string;
  };
}

/**
 * Sends a request via IPC and waits for the response.
 */
export function callVectora(method: string, payload: any): Promise<any> {
  return new Promise((resolve, reject) => {
    const client = net.createConnection(PIPE_PATH);
    const id = Math.random().toString(36).substring(7);

    const request: IPCRequest = {
      id,
      type: 'request',
      method,
      payload,
    };

    let buffer = '';

    client.on('connect', () => {
      client.write(JSON.stringify(request) + FRAME_DELIMITER);
    });

    client.on('data', (data) => {
      buffer += data.toString();
      if (buffer.endsWith(FRAME_DELIMITER)) {
        try {
          const response: IPCResponse = JSON.parse(buffer.trim());
          if (response.id === id) {
            client.end();
            if (response.error) {
              reject(new Error(response.error.message));
            } else {
              resolve(response.payload);
            }
          }
        } catch (e) {
          reject(new Error('Failed to decode IPC response'));
        }
      }
    });

    client.on('error', (err) => {
      reject(new Error('Daemon offline: ' + err.message));
    });

    // Timeout de 30s
    client.setTimeout(30000, () => {
      client.destroy();
      reject(new Error('IPC communication timeout'));
    });
  });
}
