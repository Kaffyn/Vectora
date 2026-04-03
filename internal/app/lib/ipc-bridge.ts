/**
 * IPC Bridge: Unified interface for the Vectora Frontend.
 * Automatically detects whether in Wails (Desktop) or Browser (Hono Fallback) environment.
 */

export async function callVectora(method: string, payload: any = {}): Promise<any> {
    const payloadStr = JSON.stringify(payload);

    // 1. Try Wails (Desktop)
    if (typeof window !== 'undefined' && (window as any).go?.main?.App?.CallIPC) {
        try {
            const resultStr = await (window as any).go.main.App.CallIPC(method, payloadStr);
            return JSON.parse(resultStr);
        } catch (err: any) {
            console.error("IPC Error (Wails):", err);
            throw err;
        }
    }

    // 2. Fallback to Hono/Next API (Useful for Traditional Web Development)
    // 2. Fallback to Local Go Daemon (Useful for Browser Iteration)
    const isDev = process.env.NODE_ENV === 'development';
    const baseUrl = isDev ? "http://localhost:42700/api/v1" : "/api/v1";
    
    console.warn(`Wails Bindings not found. Using ${isDev ? 'Local Daemon' : 'HTTP'} fallback...`);
    
    const response = await fetch(`${baseUrl}/${method}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: payloadStr
    });

    if (!response.ok) {
        throw new Error(`HTTP Error: ${response.status} from ${method}`);
    }

    return response.json();
}
