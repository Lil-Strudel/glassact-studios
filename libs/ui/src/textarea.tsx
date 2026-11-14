import { cn } from "./cn";
import type { PolymorphicProps } from "@kobalte/core/polymorphic";
import type { TextFieldTextAreaProps } from "@kobalte/core/text-field";
import { TextArea as TextFieldPrimitive } from "@kobalte/core/text-field";
import type { ValidComponent, VoidProps } from "solid-js";
import { splitProps } from "solid-js";

type textAreaProps<T extends ValidComponent = "textarea"> = VoidProps<
  TextFieldTextAreaProps<T> & {
    class?: string;
  }
>;

export const TextArea = <T extends ValidComponent = "textarea">(
  props: PolymorphicProps<T, textAreaProps<T>>,
) => {
  const [local, rest] = splitProps(props as textAreaProps, ["class"]);

  return (
    <TextFieldPrimitive
      class={cn(
        "flex min-h-[60px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground transition-[border-color,box-shadow] duration-150 placeholder:text-faint focus-visible:border-gilt focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-gilt/25 data-[invalid]:border-error-foreground disabled:cursor-not-allowed disabled:opacity-45",
        local.class,
      )}
      {...rest}
    />
  );
};
