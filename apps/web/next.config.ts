import type { NextConfig } from "next";

const coreAPI = process.env.CORE_API_URL ?? "http://127.0.0.1:8090";
const intelligenceAPI = process.env.INTELLIGENCE_API_URL ?? "http://127.0.0.1:8091";

const nextConfig: NextConfig = {
  poweredByHeader: false,
  async rewrites() {
    return [
      { source: "/api/core/:path*", destination: `${coreAPI}/:path*` },
      { source: "/api/intelligence/:path*", destination: `${intelligenceAPI}/:path*` },
    ];
  },
  async headers() {
    return [{
      source: "/:path*",
      headers: [
        { key: "X-Content-Type-Options", value: "nosniff" },
        { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
        { key: "X-Frame-Options", value: "DENY" },
        { key: "Permissions-Policy", value: "camera=(), microphone=(), geolocation=()" },
      ],
    }];
  },
};

export default nextConfig;
