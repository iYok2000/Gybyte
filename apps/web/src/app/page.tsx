import About from "./prisma/_components/About";
import Features from "./prisma/_components/Features";
import Hero from "./prisma/_components/Hero";

// Root route ("/") — the Prisma creative-studio landing page.
// Almarai is the default font for this subtree (matches the original design).
export default function HomePage() {
  return (
    <main
      className="bg-black min-h-screen"
      style={{
        fontFamily:
          "var(--font-almarai), -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
      }}
    >
      <Hero />
      <About />
      <Features />
    </main>
  );
}
