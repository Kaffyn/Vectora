import { Hono } from 'hono';

/**
 * Vectora Engine Hono Proxy.
 * Provides a central layer for RAG routing.
 */
export const engineProxy = new Hono();

engineProxy.get('/v1/status', (c) => {
  return c.json({
    status: 'connected',
    model: 'qwen-1.5b',
    latency: '15ms',
    uptime: process.uptime()
  });
});

export default engineProxy;
