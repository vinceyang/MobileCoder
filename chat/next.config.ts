import type { NextConfig } from "next";

// Deployment modes:
// - Docker: use "standalone" output, NO basePath
// - GitHub Pages: use "export" output with basePath
// - Dev: no special output
const deploymentType = process.env.NEXT_DEPLOYMENT_TYPE;
const isDev = process.env.NODE_ENV !== "production";
const useStandalone = deploymentType === "docker";
const useExport = !isDev && !useStandalone;

const nextConfig: NextConfig = {
  // Disable devtools overlay in development
  devIndicators: false,

  // Output mode based on deployment type
  output: isDev ? undefined : (useStandalone ? "standalone" : "export"),

  // Disable image optimization for static exports
  images: {
    unoptimized: !useStandalone,
  },

  // Only use basePath for GitHub Pages export
  basePath: useExport ? "/chat" : undefined,

  // Only use assetPrefix for GitHub Pages export
  assetPrefix: isDev ? undefined : (useExport ? "/chat/" : undefined),

  // Only use trailingSlash for GitHub Pages export
  trailingSlash: useExport,
};

export default nextConfig;
