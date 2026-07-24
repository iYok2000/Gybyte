import { ImageResponse } from "next/og";
import { readFile } from "node:fs/promises";
import { join } from "node:path";

// Apple touch icon (iOS home screen) — same circular crop of the logo,
// on a black background so it looks intentional when iOS adds its own mask.
export const size = { width: 180, height: 180 };
export const contentType = "image/png";

export default async function AppleIcon() {
  const data = await readFile(join(process.cwd(), "public/logo.png"));
  const src = `data:image/png;base64,${data.toString("base64")}`;

  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          backgroundColor: "#000000",
        }}
      >
        <img
          src={src}
          width={180}
          height={180}
          style={{ objectFit: "cover", borderRadius: "50%" }}
          alt=""
        />
      </div>
    ),
    { ...size },
  );
}
