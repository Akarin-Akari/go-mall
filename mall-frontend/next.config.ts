import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  /* config options here */
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8080/api/:path*', // Go后端API地址
      },
    ];
  },

  // 环境变量配置
  env: {
    NEXT_PUBLIC_API_BASE_URL: process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080',
    NEXT_PUBLIC_APP_NAME: 'Mall Frontend',
    NEXT_PUBLIC_APP_VERSION: '1.0.0',
  },

  // 图片优化配置
  images: {
    domains: ['localhost', '127.0.0.1'],
    formats: ['image/webp', 'image/avif'],
  },

  // 实验性功能
  experimental: {
    turbo: {
      rules: {
        '*.svg': {
          loaders: ['@svgr/webpack'],
          as: '*.js',
        },
      },
    },
  },

  // 编译配置
  compiler: {
    removeConsole: process.env.NODE_ENV === 'production',
  },

  // 输出配置
  output: 'standalone',

  // TypeScript配置
  typescript: {
    ignoreBuildErrors: false,
  },

  // ESLint配置
  eslint: {
    ignoreDuringBuilds: false,
  },
};

export default nextConfig;
