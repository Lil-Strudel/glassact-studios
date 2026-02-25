import { cn } from "./cn";
import type { PolymorphicProps } from "@kobalte/core/polymorphic";
import type { TextFieldInputProps } from "@kobalte/core/text-field";
import { TextField as TextFieldPrimitive } from "@kobalte/core/text-field";
import type { ValidComponent, VoidProps } from "solid-js";
import { splitProps } from "solid-js";
import {
  TextFieldRoot,
  TextFieldLabel,
  TextFieldDescription,
  TextFieldErrorMessage,
  textfieldLabel,
} from "./textfield";

export {
  TextFieldRoot as NumberFieldRoot,
  TextFieldLabel as NumberFieldLabel,
  TextFieldDescription as NumberFieldDescription,
  TextFieldErrorMessage as NumberFieldErrorMessage,
  textfieldLabel as numberFieldLabel,
};

type numberFieldInputProps<T extends ValidComponent = "input"> = VoidProps<
  TextFieldInputProps<T> & {
    class?: string;
    int?: boolean;
    decimalPlaces?: number;
    onChange?: (value: string) => void;
  }
>;

const filterNumberInput = (
  value: string,
  int: boolean = false,
  decimalPlaces: number = 1,
): string => {
  if (int) {
    return value.replace(/[^\d]/g, "");
  }

  const parts = value.split(".");

  if (parts.length > 2) {
    return value;
  }

  const intPart = parts[0] ?? "";
  let filtered = intPart.replace(/[^\d]/g, "");

  if (parts.length === 2) {
    const decimalPart = parts[1] ?? "";
    const decimals = decimalPart.replace(/[^\d]/g, "").slice(0, decimalPlaces);
    filtered = filtered ? `${filtered}.${decimals}` : `.${decimals}`;
  }

  return filtered;
};

export const NumberField = <T extends ValidComponent = "input">(
  props: PolymorphicProps<T, numberFieldInputProps<T>>,
) => {
  const [local, rest] = splitProps(props as numberFieldInputProps, [
    "class",
    "int",
    "decimalPlaces",
    "onChange",
  ]);

  const handleChange = (e: Event) => {
    const target = e.currentTarget as HTMLInputElement;
    const filtered = filterNumberInput(
      target.value,
      local.int || false,
      local.decimalPlaces || 1,
    );
    target.value = filtered;

    if (local.onChange) {
      local.onChange(filtered);
    }
  };

  return (
    <TextFieldPrimitive.Input
      type="text"
      inputMode="decimal"
      class={cn(
        "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-shadow file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-[2px] focus-visible:ring-primary  disabled:cursor-not-allowed disabled:opacity-50",
        local.class,
      )}
      onChange={handleChange}
      {...rest}
    />
  );
};
