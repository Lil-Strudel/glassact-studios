import { cn } from "./cn";
import { textfieldLabel } from "./textfield";
import type { PolymorphicProps } from "@kobalte/core/polymorphic";
import type {
  RadioGroupDescriptionProps,
  RadioGroupErrorMessageProps,
  RadioGroupItemLabelProps,
  RadioGroupLabelProps,
  RadioGroupRootProps,
} from "@kobalte/core/radio-group";
import { RadioGroup as RadioGroupPrimitive } from "@kobalte/core/radio-group";
import type { ValidComponent } from "solid-js";
import { splitProps } from "solid-js";

export const RadioGroupItem = RadioGroupPrimitive.Item;
export const RadioGroupItemInput = RadioGroupPrimitive.ItemInput;
export const RadioGroupItemControl = RadioGroupPrimitive.ItemControl;
export const RadioGroupItemIndicator = RadioGroupPrimitive.ItemIndicator;

type radioGroupProps<T extends ValidComponent = "div"> =
  RadioGroupRootProps<T> & { class?: string };

export const RadioGroup = <T extends ValidComponent = "div">(
  props: PolymorphicProps<T, radioGroupProps<T>>,
) => {
  const [local, rest] = splitProps(props as radioGroupProps, ["class"]);

  return (
    <RadioGroupPrimitive class={cn("space-y-1", local.class)} {...rest} />
  );
};

type radioGroupLabelProps<T extends ValidComponent = "span"> =
  RadioGroupLabelProps<T> & { class?: string };

export const RadioGroupLabel = <T extends ValidComponent = "span">(
  props: PolymorphicProps<T, radioGroupLabelProps<T>>,
) => {
  const [local, rest] = splitProps(props as radioGroupLabelProps, ["class"]);

  return (
    <RadioGroupPrimitive.Label
      class={cn(textfieldLabel(), "block", local.class)}
      {...rest}
    />
  );
};

type radioGroupDescriptionProps<T extends ValidComponent = "div"> =
  RadioGroupDescriptionProps<T> & { class?: string };

export const RadioGroupDescription = <T extends ValidComponent = "div">(
  props: PolymorphicProps<T, radioGroupDescriptionProps<T>>,
) => {
  const [local, rest] = splitProps(props as radioGroupDescriptionProps, [
    "class",
  ]);

  return (
    <RadioGroupPrimitive.Description
      class={cn(
        textfieldLabel({ description: true, label: false }),
        local.class,
      )}
      {...rest}
    />
  );
};

type radioGroupErrorMessageProps<T extends ValidComponent = "div"> =
  RadioGroupErrorMessageProps<T> & { class?: string };

export const RadioGroupErrorMessage = <T extends ValidComponent = "div">(
  props: PolymorphicProps<T, radioGroupErrorMessageProps<T>>,
) => {
  const [local, rest] = splitProps(props as radioGroupErrorMessageProps, [
    "class",
  ]);

  return (
    <RadioGroupPrimitive.ErrorMessage
      class={cn(textfieldLabel({ error: true }), local.class)}
      {...rest}
    />
  );
};

type radioGroupItemLabelProps<T extends ValidComponent = "label"> =
  RadioGroupItemLabelProps<T> & { class?: string };

export const RadioGroupItemLabel = <T extends ValidComponent = "label">(
  props: PolymorphicProps<T, radioGroupItemLabelProps<T>>,
) => {
  const [local, rest] = splitProps(props as radioGroupItemLabelProps, ["class"]);

  return (
    <RadioGroupPrimitive.ItemLabel
      class={cn(textfieldLabel(), local.class)}
      {...rest}
    />
  );
};

// The button face of a segmented control. The item's hidden input stays
// focusable, so the ring is projected onto the label sitting next to it.
export const RadioGroupButtonLabel = <T extends ValidComponent = "label">(
  props: PolymorphicProps<T, radioGroupItemLabelProps<T>>,
) => {
  const [local, rest] = splitProps(props as radioGroupItemLabelProps, ["class"]);

  return (
    <RadioGroupPrimitive.ItemLabel
      class={cn(
        "cursor-pointer select-none rounded px-3 py-1.5 text-sm font-medium text-muted-foreground transition-colors",
        "hover:text-foreground",
        "data-[checked]:bg-background data-[checked]:text-foreground data-[checked]:shadow-sm",
        "data-[disabled]:cursor-not-allowed data-[disabled]:opacity-50",
        local.class,
      )}
      {...rest}
    />
  );
};
