import { cn } from "./cn";
import { cva } from "class-variance-authority";
import type { JSX, ValidComponent } from "solid-js";
import { splitProps } from "solid-js";
import type { PolymorphicProps } from "@kobalte/core/polymorphic";
import { Polymorphic } from "@kobalte/core/polymorphic";

const alertVariants = cva(
  "relative w-full rounded-lg border px-4 py-3 text-sm [&>svg+div]:translate-y-[-3px] [&>svg]:absolute [&>svg]:left-4 [&>svg]:top-4 [&>svg~*]:pl-7",
  {
    variants: {
      variant: {
        default:
          "border-border bg-card text-foreground [&>svg]:text-muted-foreground",
        destructive:
          "border-error-foreground/30 bg-error text-error-foreground [&>svg]:text-error-foreground",
        success:
          "border-success-foreground/30 bg-success text-success-foreground [&>svg]:text-success-foreground",
        warning:
          "border-warning-foreground/30 bg-warning text-warning-foreground [&>svg]:text-warning-foreground",
        info: "border-info-foreground/30 bg-info text-info-foreground [&>svg]:text-info-foreground",
        gilt: "border-gilt/45 bg-card text-foreground [&>svg]:text-gilt",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
);

type AlertProps<T extends ValidComponent = "div"> = PolymorphicProps<
  T,
  {
    variant?: "default" | "destructive" | "success" | "warning" | "info" | "gilt";
    class?: string;
    children?: JSX.Element;
  }
>;

export function Alert<T extends ValidComponent = "div">(
  props: AlertProps<T>,
) {
  const [local, rest] = splitProps(props as AlertProps, ["class", "variant"]);
  return (
    <Polymorphic
      as="div"
      role="alert"
      class={cn(alertVariants({ variant: local.variant }), local.class)}
      {...rest}
    />
  );
}

type AlertTitleProps<T extends ValidComponent = "h5"> = PolymorphicProps<
  T,
  { class?: string; children?: JSX.Element }
>;

export function AlertTitle<T extends ValidComponent = "h5">(
  props: AlertTitleProps<T>,
) {
  const [local, rest] = splitProps(props as AlertTitleProps, ["class"]);
  return (
    <Polymorphic
      as="h5"
      class={cn("mb-1 font-semibold leading-snug", local.class)}
      {...rest}
    />
  );
}

type AlertDescriptionProps<T extends ValidComponent = "div"> = PolymorphicProps<
  T,
  { class?: string; children?: JSX.Element }
>;

export function AlertDescription<T extends ValidComponent = "div">(
  props: AlertDescriptionProps<T>,
) {
  const [local, rest] = splitProps(props as AlertDescriptionProps, ["class"]);
  return (
    <Polymorphic
      as="div"
      class={cn("text-sm [&_p]:leading-relaxed", local.class)}
      {...rest}
    />
  );
}
