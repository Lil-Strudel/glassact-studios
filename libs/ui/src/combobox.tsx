import { cn } from "./cn";
import type {
  ComboboxContentProps,
  ComboboxInputProps,
  ComboboxItemProps,
  ComboboxTriggerProps,
  ComboboxControlProps,
  ComboboxLabelProps,
  ComboboxErrorMessageProps,
  ComboboxDescriptionProps,
  ComboboxSectionProps,
  ComboboxRootProps,
} from "@kobalte/core/combobox";
import { Combobox as ComboboxPrimitive } from "@kobalte/core/combobox";
import type { PolymorphicProps } from "@kobalte/core/polymorphic";
import type { ParentProps, ValidComponent, VoidProps } from "solid-js";
import { Show, splitProps } from "solid-js";
import { textfieldLabel } from "./textfield";

export const ComboboxItemDescription = ComboboxPrimitive.ItemDescription;
export const ComboboxHiddenSelect = ComboboxPrimitive.HiddenSelect;

export const Combobox = <
  Option,
  OptGroup = never,
  T extends ValidComponent = "div",
>(
  props: PolymorphicProps<T, ComboboxRootProps<Option, OptGroup, T>>,
) => {
  return <ComboboxPrimitive triggerMode="focus" {...props} />;
};

type comboboxSectionProps<T extends ValidComponent = "li"> =
  ComboboxSectionProps<T> & { class?: string };

export const ComboboxSection = <T extends ValidComponent = "li">(
  props: PolymorphicProps<T, comboboxSectionProps<T>>,
) => {
  const [local, others] = splitProps(props as comboboxSectionProps, ["class"]);
  return (
    <ComboboxPrimitive.Section
      class={cn(
        "eyebrow overflow-hidden px-2 py-1.5 text-faint",
        local.class,
      )}
      {...others}
    />
  );
};

type comboboxDescriptionProps<T extends ValidComponent = "div"> =
  ComboboxDescriptionProps<T> & {
    class?: string;
  };

export const ComboboxDescription = <T extends ValidComponent = "div">(
  props: PolymorphicProps<T, comboboxDescriptionProps<T>>,
) => {
  const [local, rest] = splitProps(props as comboboxDescriptionProps, [
    "class",
  ]);

  return (
    <ComboboxPrimitive.Description
      class={cn(
        textfieldLabel({ description: true, label: false }),
        local.class,
      )}
      {...rest}
    />
  );
};

type comboboxErrorMessageProps<T extends ValidComponent = "div"> =
  ComboboxErrorMessageProps<T> & { class?: string };

export const ComboboxErrorMessage = <T extends ValidComponent = "div">(
  props: PolymorphicProps<T, comboboxErrorMessageProps<T>>,
) => {
  const [local, rest] = splitProps(props as comboboxErrorMessageProps, [
    "class",
  ]);

  return (
    <ComboboxPrimitive.ErrorMessage
      class={cn(textfieldLabel({ error: true }), local.class)}
      {...rest}
    />
  );
};

type comboboxLabelProps<T extends ValidComponent = "label"> =
  ComboboxLabelProps<T> & {
    class?: string;
  };

export const ComboboxLabel = <T extends ValidComponent = "label">(
  props: PolymorphicProps<T, comboboxLabelProps<T>>,
) => {
  const [local, rest] = splitProps(props as comboboxLabelProps, ["class"]);

  return (
    <ComboboxPrimitive.Label
      class={cn(textfieldLabel(), local.class)}
      {...rest}
    />
  );
};

type comboboxInputProps<T extends ValidComponent = "input"> = VoidProps<
  ComboboxInputProps<T> & {
    class?: string;
  }
>;

export const ComboboxInput = <T extends ValidComponent = "input">(
  props: PolymorphicProps<T, comboboxInputProps<T>>,
) => {
  const [local, rest] = splitProps(props as comboboxInputProps, ["class"]);

  return (
    <ComboboxPrimitive.Input
      class={cn(
        "flex size-full rounded-md bg-transparent py-3 text-sm outline-none placeholder:text-faint disabled:cursor-not-allowed disabled:opacity-45",
        local.class,
      )}
      {...rest}
    />
  );
};

type comboboxControlProps<
  U,
  T extends ValidComponent = "div",
> = ComboboxControlProps<U, T> & {
  class?: string | undefined;
};

export const ComboboxControl = <T, U extends ValidComponent = "div">(
  props: PolymorphicProps<U, comboboxControlProps<T>>,
) => {
  const [local, others] = splitProps(props as comboboxControlProps<T>, [
    "class",
  ]);
  return (
    <ComboboxPrimitive.Control
      class={cn(
        "flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm text-foreground transition-[border-color,box-shadow] duration-150 file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-faint focus-within:border-gilt focus-within:ring-2 focus-within:ring-gilt/25 focus-visible:outline-none data-[invalid]:border-error-foreground disabled:cursor-not-allowed disabled:opacity-45",
        local.class,
      )}
      {...others}
    />
  );
};

type comboboxTriggerProps<T extends ValidComponent = "button"> = ParentProps<
  ComboboxTriggerProps<T> & {
    class?: string;
  }
>;
export const ComboboxTrigger = <T extends ValidComponent = "button">(
  props: PolymorphicProps<T, comboboxTriggerProps<T>>,
) => {
  const [local, rest] = splitProps(props as comboboxTriggerProps, [
    "class",
    "children",
  ]);

  return (
    <ComboboxPrimitive.Trigger
      class={cn("my-auto size-4 text-faint", local.class)}
      {...rest}
    >
      <ComboboxPrimitive.Icon>
        <Show
          when={local.children}
          fallback={
            <svg
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="size-4"
            >
              <path d="M8 9l4 -4l4 4" />
              <path d="M16 15l-4 4l-4 -4" />
            </svg>
          }
        >
          {(children) => children()}
        </Show>
      </ComboboxPrimitive.Icon>
    </ComboboxPrimitive.Trigger>
  );
};

type comboboxContentProps<T extends ValidComponent = "div"> =
  ComboboxContentProps<T> & {
    class?: string;
  };

export const ComboboxContent = <T extends ValidComponent = "div">(
  props: PolymorphicProps<T, comboboxContentProps<T>>,
) => {
  const [local, rest] = splitProps(props as comboboxContentProps, ["class"]);

  return (
    <ComboboxPrimitive.Portal>
      <ComboboxPrimitive.Content
        class={cn(
          "relative z-50 min-w-[8rem] overflow-hidden rounded-md border border-border bg-popover text-popover-foreground shadow-2xl shadow-black/50 data-[expanded]:animate-in data-[closed]:animate-out data-[closed]:fade-out-0 data-[expanded]:fade-in-0 data-[closed]:zoom-out-95 data-[expanded]:zoom-in-95 origin-[--kb-combobox-content-transform-origin]",
          local.class,
        )}
        {...rest}
      >
        <ComboboxPrimitive.Listbox class="p-1" />
      </ComboboxPrimitive.Content>
    </ComboboxPrimitive.Portal>
  );
};

type comboboxItemProps<T extends ValidComponent = "li"> = ParentProps<
  ComboboxItemProps<T> & {
    class?: string;
  }
>;

export const ComboboxItem = <T extends ValidComponent = "li">(
  props: PolymorphicProps<T, comboboxItemProps<T>>,
) => {
  const [local, rest] = splitProps(props as comboboxItemProps, [
    "class",
    "children",
  ]);

  return (
    <ComboboxPrimitive.Item
      class={cn(
        "relative flex w-full cursor-default select-none items-center rounded-sm py-1.5 pl-2 pr-8 text-sm outline-none data-[disabled]:pointer-events-none data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground data-[disabled]:opacity-45",
        local.class,
      )}
      {...rest}
    >
      <ComboboxPrimitive.ItemIndicator class="absolute right-2 flex h-3.5 w-3.5 items-center justify-center text-gilt">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          viewBox="0 0 24 24"
          class="h-4 w-4"
        >
          <path
            fill="none"
            stroke="currentColor"
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="m5 12l5 5L20 7"
          />
          <title>Checked</title>
        </svg>
      </ComboboxPrimitive.ItemIndicator>
      <ComboboxPrimitive.ItemLabel>
        {local.children}
      </ComboboxPrimitive.ItemLabel>
    </ComboboxPrimitive.Item>
  );
};
