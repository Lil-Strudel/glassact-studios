import { cn } from "./cn";
import { cva, type VariantProps } from "class-variance-authority";
import type { JSX } from "solid-js";

const badgeVariants = cva(
  "inline-flex items-center rounded-md px-2.5 py-0.5 text-xs font-semibold transition-colors",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground",
        secondary: "bg-secondary text-secondary-foreground",
        outline: "border border-input bg-background",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
);

export interface BadgeProps extends JSX.HTMLAttributes<HTMLDivElement> {
  variant?: "default" | "secondary" | "outline";
  children?: JSX.Element;
}

export function Badge(props: BadgeProps) {
  return (
    <div class={cn(badgeVariants({ variant: props.variant }), props.class)} {...props}>
      {props.children}
    </div>
  );
}

export { badgeVariants };
