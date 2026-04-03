/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  images: {
    unoptimized: true,
  },
  trailingSlash: true,
  /* 
   * Turbopack: Default build engine in Next.js 16.
   * Setted to empty object to avoid custom config error.
   */
  turbopack: {},
};

export default nextConfig;
