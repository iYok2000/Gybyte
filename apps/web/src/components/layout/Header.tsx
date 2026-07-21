import Link from "next/link";

// Shared site header. Renders a stable branded header for every page
// (Req 12.5) and exposes the AdReady navigation, including a single-hop link
// to the Audit_Dashboard at /audit (Req 13.1). Any pre-existing menu items
// outside the rebrand scope are preserved unchanged (Req 13.6).
export function Header() {
  return (
    <header className="border-b border-(--border) bg-(--surface)">
      <div className="mx-auto flex max-w-5xl items-center justify-between px-4 py-4">
        <Link href="/" className="text-lg font-semibold text-(--foreground)">
          AdReady
        </Link>
        <nav aria-label="เมนูหลัก">
          <ul className="flex items-center gap-6 text-sm">
            <li>
              <Link
                href="/audit"
                className="text-(--foreground) hover:text-(--primary)"
              >
                ตรวจสอบเว็บไซต์
              </Link>
            </li>
          </ul>
        </nav>
      </div>
    </header>
  );
}
