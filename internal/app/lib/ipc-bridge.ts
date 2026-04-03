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
    console.warn("Wails Bindings not found. Using HTTP fallback...");
    const response = await fetch(`/api/v1/${method}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: payloadStr
    });

    if (!response.ok) {
        throw new Error(`HTTP Error: ${response.status}`);
    }

    return response.json();
}
