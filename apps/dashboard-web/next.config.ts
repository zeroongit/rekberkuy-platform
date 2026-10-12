import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactCompiler: true,
  async redirects() {
    return [
      { source: '/transactions', destination: '/dashboard/transactions', permanent: true },
      { source: '/ledger', destination: '/dashboard/ledger', permanent: true },
      { source: '/sengketa', destination: '/dashboard/disputes', permanent: true },
      { source: '/marketplace', destination: '/dashboard/catalog', permanent: true },
      { source: '/kyc', destination: '/dashboard/kyc', permanent: true },
      { source: '/admin', destination: '/dashboard/admin', permanent: true },
    ];
  },
};

export default nextConfig;
