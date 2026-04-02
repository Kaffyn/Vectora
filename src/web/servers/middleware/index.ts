import { createMiddleware } from 'hono/factory';

/**
 * Custom Hono middleware for Vectora authentication or validation.
 */
export const authMiddleware = createMiddleware(async (c, next) => {
  // Authentication logic goes here
  await next();
});
