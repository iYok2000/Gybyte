import { ImageResponse } from "next/og";
import { readFile } from "node:fs/promises";
import { join } from "node:path";

// Dynamic favicon — crops public/logo.png into a circle (transparent corners).
export const size = { width: 64, height: 64 };
export const contentType = "image/png";

export default async function Icon() {
  const data = await readFile(join(process.cwd(), "public/logo.png"));
  const src = `data:image/png;base64,${data.toString("base64")}`;

  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          borderRadius: "50%",
          overflow: "hidden",
        }}
      >
        <img
          src={src}
          width={64}
          height={64}
          style={{ objectFit: "cover", borderRadius: "50%" }}
          alt=""
        />
      </div>
    ),
    { ...size },
  );
}
