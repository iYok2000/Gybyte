import { ImageResponse } from "next/og";

// Default social share image (Open Graph + Twitter) for the whole site.
// Rendered at build/request time via next/og. Uses English + brand only so it
// renders reliably with the default font (no Thai glyph loading needed).
export const alt = "AdReady — free Google AdSense readiness checker";
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

export default function OpengraphImage() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          alignItems: "flex-start",
          justifyContent: "center",
          backgroundColor: "#000000",
          padding: "80px",
          color: "#E1E0CC",
          fontFamily: "sans-serif",
        }}
      >
        <div
          style={{
            fontSize: 40,
            color: "#DEDBC8",
            opacity: 0.7,
            letterSpacing: 4,
            textTransform: "uppercase",
          }}
        >
          AdSense readiness checker
        </div>
        <div
          style={{
            display: "flex",
            alignItems: "flex-start",
            fontSize: 200,
            fontWeight: 700,
            lineHeight: 1,
            marginTop: 16,
            letterSpacing: -6,
          }}
        >
          AdReady
          <span style={{ fontSize: 60, marginTop: 20 }}>*</span>
        </div>
        <div
          style={{
            fontSize: 40,
            color: "#DEDBC8",
            opacity: 0.85,
            marginTop: 24,
            maxWidth: 900,
          }}
        >
          Check ads.txt, HTTPS, robots.txt &amp; policy pages — get a 0–100
          readiness score and the exact fixes to get approved.
        </div>
      </div>
    ),
    { ...size },
  );
}
