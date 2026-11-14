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
    allowNegative?: boolean;
    onChange?: (value: string) => void;
  }
>;

const filterNumberInput = (
  value: string,
  int: boolean = false,
  decimalPlaces: number = 1,
  allowNegative: boolean = false,
): string => {
  const sign = allowNegative && value.trimStart().startsWith("-") ? "-" : "";

  if (int) {
    return sign + value.replace(/[^\d]/g, "");
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

  return sign + filtered;
};

export const NumberField = <T extends ValidComponent = "input">(
  props: PolymorphicProps<T, numberFieldInputProps<T>>,
) => {
  const [local, rest] = splitProps(props as numberFieldInputProps, [
    "class",
    "int",
    "decimalPlaces",
    "allowNegative",
    "onChange",
  ]);

  const handleChange = (e: Event) => {
    const target = e.currentTarget as HTMLInputElement;
    const filtered = filterNumberInput(
      target.value,
      local.int || false,
      local.decimalPlaces || 1,
      local.allowNegative || false,
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
        "flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm text-foreground transition-[border-color,box-shadow] duration-150 file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-faint focus-visible:border-gilt focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-gilt/25 aria-[invalid=true]:border-error-foreground data-[invalid]:border-error-foreground disabled:cursor-not-allowed disabled:opacity-45",
        local.class,
      )}
      onChange={handleChange}
      {...rest}
    />
  );
};
