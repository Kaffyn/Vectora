import { NextRequest, NextResponse } from 'next/server';
import { callVectora } from '@/lib/ipc-client';

/**
 * Novo Handler de API para Next.js (Adeus Hono!)
 * Atua como uma ponte leve entre o Frontend e o Daemon Go via IPC.
 */

/**
 * Force dynamic for Next.js 16 to avoid static generation attempts 
 * during build, even if ignoring this route in export.
 */
export const dynamic = 'force-dynamic';

async function handleIPC(req: NextRequest) {
  const url = new URL(req.url);
  const path = url.pathname.replace('/api/v1/', '');
  
  // Legacy route mapping for IPC methods
  const routeMap: Record<string, string> = {
    'chat': 'workspace.query',
    'conversations': 'chat.list',
    'settings': 'provider.get',
    'search': 'memory.search',
  };

  const method = routeMap[path] || path;
  
  try {
    let payload: any = {};
    if (req.method === 'POST' || req.method === 'PATCH') {
      payload = await req.json();
    } else {
      // Convert query params to payload if needed
      payload = Object.fromEntries(url.searchParams.entries());
    }

    const result = await callVectora(method, payload);
    return NextResponse.json(result);
  } catch (err: any) {
    return NextResponse.json(
      { error: err.message || 'IPC Communication Error' }, 
      { status: 503 }
    );
  }
}

export const GET = handleIPC;
export const POST = handleIPC;
export const PATCH = handleIPC;
export const DELETE = handleIPC;
