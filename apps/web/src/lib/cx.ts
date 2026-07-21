// Tiny className joiner used across the shared UI components.
// Filters out falsy values (undefined/null/false/"") so callers can write
// conditional classes inline, then joins the survivors with a single space.
export function cx(...classes: Array<string | false | null | undefined>): string {
  return classes.filter(Boolean).join(" ");
}

export default cx;
