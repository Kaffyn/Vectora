import { Hono } from 'hono';
import { logger } from 'hono/logger';
import { cors } from 'hono/cors';

import { mainRouter } from './routes';

const app = new Hono().basePath('/api');

// Middleware
app.use('*', logger());
app.use('*', cors());

// Routes
app.route('/v1', mainRouter);

app.get('/health', async (c) => {
  try {
    const GO_BASE = (process.env.RAG_BACKEND_URL || 'http://localhost:8080');
    // Faz ping no backend Go (que por sua vez foi iniciado start.go)
    const goRes = await fetch(`${GO_BASE}/api/health`, { signal: AbortSignal.timeout(2000) });
    if (!goRes.ok) throw new Error("Go backend responded with error");
    const data = await goRes.json();
    return c.json(data, 200);
  } catch (err) {
    return c.json({ status: 'error', message: 'Go Core Offline' }, 503);
  }
});

// For Next.js handler
export { app };
export default app;
