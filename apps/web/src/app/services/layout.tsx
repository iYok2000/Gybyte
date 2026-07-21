import { SiteNav } from "@/components/layout/SiteNav";
import { BodyLayout } from "@/components/layout/BodyLayout";

// AdReady shell for /services/* pages — same floating-pill SiteNav as the rest
// of the site for a consistent header (Req 12.5). pt-16 clears the fixed nav.
export default function ServicesLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div
      className="flex min-h-screen flex-col bg-black pt-16 text-[#E1E0CC]"
      style={{
        fontFamily:
          "var(--font-almarai), -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
      }}
    >
      <SiteNav />
      <BodyLayout>{children}</BodyLayout>
    </div>
  );
}
