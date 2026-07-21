import type { ReactNode } from "react";
import { cx } from "@/lib/cx";

// Shared Badge component (Req 10.3, 15.7).
// Used by the audit results list to distinguish pass/fail statuses visually
// (Req 11.2). Colors resolve to the CSS custom-property design tokens.

export type BadgeVariant =
  | "default"
  | "primary"
  | "success"
  | "warning"
  | "error"
  | "outline"
  | "violet"
  | "glow";

export type BadgeSize = "xs" | "sm" | "md" | "lg";

export interface BadgeProps {
  children: ReactNode;
  variant?: BadgeVariant;
  size?: BadgeSize;
  /** Optional leading icon. */
  icon?: ReactNode;
  /** When provided, renders a removable "×" control invoking this callback. */
  onRemove?: () => void;
  /** Adds a pulsing dot to draw attention. */
  pulse?: boolean;
  className?: string;
}

const baseClasses =
  "inline-flex items-center gap-1 rounded-full font-medium leading-none";

const variantClasses: Record<BadgeVariant, string> = {
  default: "bg-(--surface-muted) text-(--muted)",
  primary: "bg-(--primary-soft) text-(--primary)",
  success: "bg-(--success)/15 text-(--success)",
  warning: "bg-(--warning)/15 text-(--warning)",
  error: "bg-(--error)/15 text-(--error)",
  outline: "border border-(--border) text-(--foreground)",
  violet: "bg-(--violet)/15 text-(--violet)",
  glow: "bg-(--primary) text-white shadow-[0_0_12px_var(--primary-soft)]",
};

const sizeClasses: Record<BadgeSize, string> = {
  xs: "px-1.5 py-0.5 text-[10px]",
  sm: "px-2 py-0.5 text-xs",
  md: "px-2.5 py-1 text-sm",
  lg: "px-3 py-1.5 text-base",
};

export function Badge({
  children,
  variant = "default",
  size = "sm",
  icon,
  onRemove,
  pulse = false,
  className,
}: BadgeProps) {
  return (
    <span className={cx(baseClasses, variantClasses[variant], sizeClasses[size], className)}>
      {pulse && (
        <span
          aria-hidden="true"
          className="h-1.5 w-1.5 animate-pulse rounded-full bg-current"
        />
      )}
      {icon}
      {children}
      {typeof onRemove === "function" && (
        <button
          type="button"
          onClick={onRemove}
          aria-label="Remove"
          className="ml-0.5 inline-flex items-center justify-center rounded-full leading-none opacity-70 hover:opacity-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-(--focus-ring)"
        >
          ×
        </button>
      )}
    </span>
  );
}

export default Badge;
