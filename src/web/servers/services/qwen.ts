const GO_BASE = (process.env.RAG_BACKEND_URL || 'http://localhost:8080/api/chat').replace('/api/chat', '');

/**
 * Utilitário para proxyear as requisições para o Backend Go (Core Vectora).
 * Modo: Qwen Local GGUF Engine
 */
export async function proxyToGo(path: string, method: string, body?: unknown) {
  try {
    const res = await fetch(`${GO_BASE}${path}`, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });
    const text = await res.text();
    return new Response(text, { 
      status: res.status, 
      headers: { 'Content-Type': 'application/json' } 
    });
  } catch (e) {
    return Response.json({ error: 'Go Core offline' }, { status: 503 });
  }
}
