import type { ButtonHTMLAttributes, ReactNode } from "react";
import { cx } from "@/lib/cx";

// Shared Button component (Req 10.3, 15.7).
// Colors resolve to the CSS custom-property design tokens from globals.css.

export type ButtonVariant =
  | "primary"
  | "secondary"
  | "large"
  | "small"
  | "pill"
  | "link"
  | "ghost";

export type ButtonSize = "default" | "small" | "large";

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  /** Shows an inline spinner, sets aria-busy, and disables the button. */
  loading?: boolean;
  /** Stretches the button to fill its container width. */
  fullWidth?: boolean;
  /** Icon rendered before the label. */
  icon?: ReactNode;
  /** Icon rendered after the label. */
  iconAfter?: ReactNode;
}

const baseClasses =
  "inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-(--focus-ring) disabled:cursor-not-allowed disabled:opacity-60";

const variantClasses: Record<ButtonVariant, string> = {
  primary: "bg-(--primary) text-white hover:bg-(--primary-hover)",
  secondary:
    "bg-(--surface) text-(--foreground) border border-(--border) hover:border-(--border-hover)",
  // large/small are style presets that also carry sizing hints.
  large: "bg-(--primary) text-white hover:bg-(--primary-hover) px-6 py-3 text-lg",
  small: "bg-(--primary) text-white hover:bg-(--primary-hover) px-3 py-1 text-sm",
  pill: "bg-(--primary) text-white hover:bg-(--primary-hover) rounded-full",
  link: "bg-transparent text-(--primary) hover:text-(--primary-hover) underline underline-offset-2 p-0",
  ghost: "bg-transparent text-(--foreground) hover:bg-(--surface-muted)",
};

const sizeClasses: Record<ButtonSize, string> = {
  default: "px-4 py-2 text-sm",
  small: "px-3 py-1 text-sm",
  large: "px-6 py-3 text-base",
};

// Variants that bring their own padding shouldn't also get size padding.
const variantsWithOwnSizing = new Set<ButtonVariant>(["large", "small", "link"]);

export function Button({
  variant = "primary",
  size = "default",
  loading = false,
  fullWidth = false,
  icon,
  iconAfter,
  disabled,
  className,
  children,
  type,
  ...rest
}: ButtonProps) {
  const isDisabled = disabled || loading;

  return (
    <button
      // Default to type="button" so buttons don't accidentally submit forms.
      type={type ?? "button"}
      disabled={isDisabled}
      aria-busy={loading || undefined}
      className={cx(
        baseClasses,
        variantClasses[variant],
        !variantsWithOwnSizing.has(variant) && sizeClasses[size],
        fullWidth && "w-full",
        className,
      )}
      {...rest}
    >
      {loading ? (
        <span
          aria-hidden="true"
          className="h-4 w-4 animate-spin rounded-full border-2 border-current border-t-transparent"
        />
      ) : (
        icon
      )}
      {children}
      {!loading && iconAfter}
    </button>
  );
}

export default Button;
