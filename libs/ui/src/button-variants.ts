import { cva, VariantProps } from "class-variance-authority";

export const buttonVariants = cva(
  "inline-flex items-center justify-center rounded-md text-sm font-medium transition-[color,background-color,border-color,box-shadow] duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:pointer-events-none disabled:opacity-45",
  {
    variants: {
      variant: {
        default:
          "border border-crimson-light/30 bg-primary text-primary-foreground shadow-[inset_0_1px_0_hsl(var(--primary-foreground)/0.14)] hover:brightness-110",
        destructive:
          "border border-error-foreground/45 text-error-foreground hover:bg-error hover:border-error-foreground/70",
        "destructive-solid":
          "bg-destructive text-destructive-foreground hover:brightness-110",
        outline:
          "border border-input bg-transparent text-foreground hover:bg-accent hover:border-gilt/50",
        secondary:
          "border border-input bg-secondary text-secondary-foreground hover:border-gilt/60",
        ghost:
          "text-muted-foreground hover:bg-accent hover:text-foreground",
        link: "text-crimson-light underline-offset-4 hover:underline",
        text: "text-crimson-light hover:text-foreground",
        gilt: "border border-gilt/55 text-gilt hover:bg-gilt/10 hover:border-gilt",
      },
      size: {
        default: "h-9 px-4 py-2",
        sm: "h-8 rounded-md px-3 text-xs",
        lg: "h-11 rounded-md px-8 text-base",
        icon: "h-9 w-9",
        "icon-sm": "h-8 w-8",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
);

export type buttonVariantsProps = VariantProps<typeof buttonVariants>;
