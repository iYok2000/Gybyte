import type { ReactNode } from "react";
import { cx } from "@/lib/cx";

// Shared Card component (Req 10.3, 15.7).
// Uses the CSS custom-property design tokens from globals.css so every surface
// stays visually consistent with the rest of apps/web.

export type CardVariant =
  | "default"
  | "bordered"
  | "elevated"
  | "flat"
  | "glass"
  | "premium";

export type CardPadding = "none" | "xs" | "sm" | "md" | "lg";

export interface CardProps {
  children: ReactNode;
  variant?: CardVariant;
  padding?: CardPadding;
  /** Optional content rendered above the body, separated by a divider. */
  header?: ReactNode;
  /** Optional content rendered below the body, separated by a divider. */
  footer?: ReactNode;
  /** When provided the whole card becomes an interactive button. */
  onClick?: () => void;
  /** Adds a subtle hover treatment (implied when onClick is set). */
  hover?: boolean;
  className?: string;
}

// Base look shared by every variant.
const baseClasses = "rounded-xl text-(--foreground) transition-colors";

const variantClasses: Record<CardVariant, string> = {
  default: "bg-(--card) border border-(--border)",
  bordered: "bg-(--card) border-2 border-(--border)",
  elevated: "bg-(--card) border border-(--border) shadow-md",
  flat: "bg-(--surface-muted)",
  glass: "bg-(--card)/70 border border-(--border) backdrop-blur-md",
  premium:
    "bg-(--card) border border-(--primary) shadow-lg ring-1 ring-(--primary-soft)",
};

const paddingClasses: Record<CardPadding, string> = {
  none: "p-0",
  xs: "p-2",
  sm: "p-3",
  md: "p-5",
  lg: "p-8",
};

export function Card({
  children,
  variant = "default",
  padding = "md",
  header,
  footer,
  onClick,
  hover,
  className,
}: CardProps) {
  const isInteractive = typeof onClick === "function";
  const interactiveHover = hover || isInteractive;

  // Header/footer get their own padding + divider so the body padding stays clean.
  const body = (
    <>
      {header != null && (
        <div className={cx(paddingClasses[padding], "border-b border-(--border)")}>
          {header}
        </div>
      )}
      <div className={paddingClasses[padding]}>{children}</div>
      {footer != null && (
        <div className={cx(paddingClasses[padding], "border-t border-(--border)")}>
          {footer}
        </div>
      )}
    </>
  );

  const classes = cx(
    baseClasses,
    variantClasses[variant],
    "overflow-hidden",
    interactiveHover && "hover:border-(--border-hover)",
    isInteractive &&
      "w-full text-left cursor-pointer focus:outline-none focus-visible:ring-2 focus-visible:ring-(--focus-ring)",
    className,
  );

  if (isInteractive) {
    return (
      <button type="button" onClick={onClick} className={classes}>
        {body}
      </button>
    );
  }

  return <div className={classes}>{body}</div>;
}

export default Card;
