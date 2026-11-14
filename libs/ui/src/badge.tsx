import { cn } from "./cn";
import { cva } from "class-variance-authority";
import type { JSX } from "solid-js";
import { Show, splitProps } from "solid-js";

const tonePill = "rounded-full border px-2.5 py-0.5";

const badgeVariants = cva(
  "inline-flex items-center gap-1.5 whitespace-nowrap text-xs font-medium transition-colors",
  {
    variants: {
      variant: {
        default: "rounded-full bg-primary px-2.5 py-0.5 text-primary-foreground",
        secondary:
          "rounded-full border border-input bg-secondary px-2.5 py-0.5 text-secondary-foreground",
        outline:
          "rounded-full border border-input px-2.5 py-0.5 text-muted-foreground",
        warning: `${tonePill} border-warning-foreground/35 bg-warning text-warning-foreground`,
        success: `${tonePill} border-success-foreground/35 bg-success text-success-foreground`,
        info: `${tonePill} border-info-foreground/35 bg-info text-info-foreground`,
        error: `${tonePill} border-error-foreground/35 bg-error text-error-foreground`,
        gilt: `${tonePill} border-gilt/40 bg-gilt/10 text-gilt`,
        crimson: `${tonePill} border-crimson-light/35 bg-crimson-deep/40 text-crimson-light`,
        draft: `${tonePill} border-status-draft/35 bg-status-draft/15 text-status-draft`,
        ordered: `${tonePill} border-status-ordered/35 bg-status-ordered/15 text-status-ordered`,
        production: `${tonePill} border-status-production/35 bg-status-production/15 text-status-production`,
        shipped: `${tonePill} border-status-shipped/35 bg-status-shipped/15 text-status-shipped`,
        invoiced: `${tonePill} border-status-invoiced/35 bg-status-invoiced/15 text-status-invoiced`,
        completed: `${tonePill} border-status-completed/35 bg-status-completed/15 text-status-completed`,
        cancelled: `${tonePill} border-status-cancelled/35 bg-status-cancelled/15 text-status-cancelled line-through`,
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
);

export type BadgeVariant =
  | "default"
  | "secondary"
  | "outline"
  | "warning"
  | "success"
  | "info"
  | "error"
  | "gilt"
  | "crimson"
  | "draft"
  | "ordered"
  | "production"
  | "shipped"
  | "invoiced"
  | "completed"
  | "cancelled";

export interface BadgeProps extends JSX.HTMLAttributes<HTMLDivElement> {
  variant?: BadgeVariant;
  lozenge?: boolean;
  children?: JSX.Element;
}

export function Badge(props: BadgeProps) {
  const [local, rest] = splitProps(props, [
    "variant",
    "lozenge",
    "class",
    "children",
  ]);

  return (
    <div
      class={cn(badgeVariants({ variant: local.variant }), local.class)}
      {...rest}
    >
      <Show when={local.lozenge}>
        <span class="lozenge" aria-hidden="true" />
      </Show>
      {local.children}
    </div>
  );
}

export { badgeVariants };
