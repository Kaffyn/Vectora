import { handle } from 'hono/vercel';
import { app } from '@/servers';

/**
 * Entry point for Hono API in Next.js App Router.
 * This handles all requests matching /api/[[...route]]
 */

export const GET = handle(app);
export const POST = handle(app);
export const PATCH = handle(app);
export const DELETE = handle(app);
export const PUT = handle(app);
export const OPTIONS = handle(app);
