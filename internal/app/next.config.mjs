/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'export',
  images: {
    unoptimized: true,
  },
  trailingSlash: true,
  // Desativa o Hono de build do Next.js já que o Wails servirá apenas estáticos
  webpack: (config) => {
    config.externals.push({
      'hono/vercel': 'commonjs hono/vercel',
    });
    return config;
  },
};

export default nextConfig;
