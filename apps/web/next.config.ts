import type { NextConfig } from "next";

// AdReady web app runs as a standalone server output so it can be containerized
// independently of the monorepo build tooling (Req 15.1, 15.2).
const nextConfig: NextConfig = {
  output: "standalone",
};

export default nextConfig;
