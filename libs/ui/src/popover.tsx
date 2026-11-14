import { cn } from "./cn";
import type { PolymorphicProps } from "@kobalte/core/polymorphic";
import type {
  PopoverContentProps,
  PopoverRootProps,
} from "@kobalte/core/popover";
import { Popover as PopoverPrimitive } from "@kobalte/core/popover";
import { type ValidComponent, mergeProps, splitProps } from "solid-js";

export const PopoverTrigger = PopoverPrimitive.Trigger;

export const Popover = (props: PopoverRootProps) => {
  const merge = mergeProps<PopoverRootProps[]>({ gutter: 4 }, props);

  return <PopoverPrimitive {...merge} />;
};

type popoverContentProps<T extends ValidComponent = "div"> =
  PopoverContentProps<T> & {
    class?: string;
  };

export const PopoverContent = <T extends ValidComponent = "div">(
  props: PolymorphicProps<T, popoverContentProps<T>>,
) => {
  const [local, rest] = splitProps(props as popoverContentProps, ["class"]);

  return (
    <PopoverPrimitive.Portal>
      <PopoverPrimitive.Content
        class={cn(
          "z-50 w-72 rounded-md border border-border bg-popover p-3 text-popover-foreground shadow-2xl shadow-black/50 outline-none data-[expanded]:animate-in data-[closed]:animate-out data-[closed]:fade-out-0 data-[expanded]:fade-in-0 data-[closed]:zoom-out-95 data-[expanded]:zoom-in-95",
          local.class,
        )}
        {...rest}
      />
    </PopoverPrimitive.Portal>
  );
};
