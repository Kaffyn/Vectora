import { getTranslations, Locale } from '../../lib/i18n';
import { Hono } from "hono";

/**
 * Main application routes combining all domain services.
 */
export const mainRouter = new Hono();

const GO_BASE = (process.env.RAG_BACKEND_URL || 'http://localhost:8080/api/chat').replace('/api/chat', '');

import { zValidator } from '@hono/zod-validator';
import { chatRequestSchema } from '../utils/validators';
import { proxyToGo } from '../services/qwen';
import { handleGeminiChat } from '../services/gemini';

// ── Chat ───────────────────────────────────────────────────────────────────
mainRouter.get('/chat', (c) => {
  return c.json({
    message: 'Welcome to Vectora API',
    capabilities: ['RAG', 'DeepSearch', 'MCP'],
    providers: ['qwen', 'gemini']
  });
});

mainRouter.post('/chat', zValidator('json', chatRequestSchema), async (c) => {
  const reqMeta = c.req.valid('json');

  if (reqMeta.provider === 'gemini') {
    const geminiRes = await handleGeminiChat(reqMeta.message, reqMeta.api_key);
    return c.json(geminiRes.body, geminiRes.status as any);
  }

  // Fallback / Padrão: Qwen Go Native
  return proxyToGo('/api/chat', 'POST', reqMeta);
});

// ── Settings ───────────────────────────────────────────────────────────────
mainRouter.get('/settings', (c) => proxyToGo('/api/settings', 'GET'));
mainRouter.post('/settings', async (c) => proxyToGo('/api/settings', 'POST', await c.req.json()));

// ── Conversations ──────────────────────────────────────────────────────────
mainRouter.get('/conversations', (c) => proxyToGo('/api/conversations', 'GET'));
mainRouter.post('/conversations', async (c) => proxyToGo('/api/conversations', 'POST', await c.req.json()));
mainRouter.get('/conversations/:id', (c) => proxyToGo(`/api/conversations/${c.req.param('id')}`, 'GET'));
mainRouter.patch('/conversations/:id', async (c) => proxyToGo(`/api/conversations/${c.req.param('id')}`, 'PATCH', await c.req.json()));
mainRouter.delete('/conversations/:id', (c) => proxyToGo(`/api/conversations/${c.req.param('id')}`, 'DELETE'));
mainRouter.post('/conversations/:id/messages', async (c) => proxyToGo(`/api/conversations/${c.req.param('id')}/messages`, 'POST', await c.req.json()));

// ── RAG Search ─────────────────────────────────────────────────────────────
mainRouter.post('/search', async (c) => {
  const body = await c.req.json();
  return proxyToGo('/api/search', 'POST', body);
});

// ── MCP Tool Calls ─────────────────────────────────────────────────────────
mainRouter.get('/mcp/tools', (c) => proxyToGo('/api/mcp/tools', 'GET'));
mainRouter.post('/mcp/tools/call', async (c) => {
  const body = await c.req.json();
  return proxyToGo('/api/mcp/tools/call', 'POST', body);
});

// ── i18n ───────────────────────────────────────────────────────────────────
mainRouter.get('/i18n', (c) => {
  const locale = (c.req.query('locale') || 'pt') as Locale;
  const translations = getTranslations(locale);
  return c.json(translations);
});
