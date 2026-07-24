import type { MetadataRoute } from "next";

// Site base URL — set NEXT_PUBLIC_SITE_URL to the real domain in production.
const baseUrl = process.env.NEXT_PUBLIC_SITE_URL || "http://localhost:3000";

// /robots.txt — allow all crawlers and point them at the sitemap (SEO).
export default function robots(): MetadataRoute.Robots {
  return {
    rules: {
      userAgent: "*",
      allow: "/",
    },
    sitemap: `${baseUrl}/sitemap.xml`,
    host: baseUrl,
  };
}
