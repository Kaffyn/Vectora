import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

/**
 * Next.js 16 Proxy (formerly middleware).
 * Currently just passes through requests.
 */
export function proxy(request: NextRequest) {
  return NextResponse.next();
}

export default proxy;
