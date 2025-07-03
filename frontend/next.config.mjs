/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    remotePatterns: [
      // Development configuration
      {
        protocol: 'http',
        hostname: 'localhost',
        port: '8080',
        pathname: '/uploads/**',
      },
      // Production configuration - add your domain here
      ...(process.env.NEXT_PUBLIC_DOMAIN ? [
        {
          protocol: 'https',
          hostname: process.env.NEXT_PUBLIC_DOMAIN,
          pathname: '/uploads/**',
        },
        {
          protocol: 'https',
          hostname: `www.${process.env.NEXT_PUBLIC_DOMAIN}`,
          pathname: '/uploads/**',
        }
      ] : []),
    ],
  },
  // Enable standalone output for Docker production builds
  output: process.env.NODE_ENV === 'production' ? 'standalone' : undefined,
};

export default nextConfig;
