import type { AnyFieldApi } from "@tanstack/solid-form";
import {
  TextFieldRoot,
  TextFieldLabel,
  TextField,
  TextFieldDescription,
  TextFieldErrorMessage,
  textfieldLabel,
} from "./textfield";
import { cn } from "./cn";
import { createEffect, createMemo, createSignal, JSX, Show } from "solid-js";
import { TextArea } from "./textarea";
import {
  Combobox,
  ComboboxContent,
  ComboboxControl,
  ComboboxDescription,
  ComboboxErrorMessage,
  ComboboxInput,
  ComboboxItem,
  ComboboxLabel,
  ComboboxTrigger,
} from "./combobox";

function useValidationState(field: () => AnyFieldApi) {
  const [validationState, setValidationState] = createSignal<
    "valid" | "invalid"
  >("valid");

  createEffect(() => {
    if (field().state.meta.errors.length > 0 && field().state.meta.isTouched) {
      setValidationState("invalid");
    } else {
      setValidationState("valid");
    }
  });

  return validationState;
}

interface FormTextFieldProps {
  field: () => AnyFieldApi;
  class?: JSX.HTMLAttributes<"div">["class"];
  label?: string;
  placeholder?: string;
  description?: string;
  fullWidth?: boolean;
}
function FormTextField(props: FormTextFieldProps) {
  const { field, label, placeholder, description, fullWidth = true } = props;
  const validationState = useValidationState(field);

  return (
    <TextFieldRoot
      class={cn(fullWidth && "w-full", props.class)}
      validationState={validationState()}
    >
      {label && <TextFieldLabel>{label}</TextFieldLabel>}
      <TextField
        placeholder={placeholder}
        name={field().name}
        value={field().state.value}
        onBlur={field().handleBlur}
        onChange={(e) => field().handleChange(e.target.value)}
      />
      {description && (
        <TextFieldDescription>{description}</TextFieldDescription>
      )}
      <TextFieldErrorMessage>
        {field()
          .state.meta.errors.map((error) => error?.message)
          .join(", ")}
      </TextFieldErrorMessage>
    </TextFieldRoot>
  );
}

interface FormTextAreaProps {
  field: () => AnyFieldApi;
  class?: JSX.HTMLAttributes<"div">["class"];
  label?: string;
  placeholder?: string;
  description?: string;
  fullWidth?: boolean;
}
function FormTextArea(props: FormTextAreaProps) {
  const { field, label, placeholder, description, fullWidth = true } = props;
  const validationState = useValidationState(field);

  return (
    <TextFieldRoot
      class={cn(fullWidth && "w-full", props.class)}
      validationState={validationState()}
    >
      {label && <TextFieldLabel>{label}</TextFieldLabel>}
      <TextArea
        placeholder={placeholder}
        name={field().name}
        value={field().state.value}
        onBlur={field().handleBlur}
        onChange={(e) => field().handleChange(e.target.value)}
      />
      {description && (
        <TextFieldDescription>{description}</TextFieldDescription>
      )}
      <TextFieldErrorMessage>
        {field()
          .state.meta.errors.map((error) => error?.message)
          .join(", ")}
      </TextFieldErrorMessage>
    </TextFieldRoot>
  );
}

interface FormComboboxProps<T> {
  field: () => AnyFieldApi;
  options: { label: string; value: T; disabled?: boolean }[];
  class?: JSX.HTMLAttributes<"div">["class"];
  label?: string;
  placeholder?: string;
  description?: string;
  fullWidth?: boolean;
}
function FormCombobox<T>(props: FormComboboxProps<T>) {
  const {
    field,
    label,
    placeholder,
    description,
    fullWidth = true,
    options,
  } = props;
  const validationState = useValidationState(field);

  const optionLabels = () => options.map((o) => o.label);

  createEffect(() => {
    const labels = optionLabels();
    const uniqueLabels = new Set(labels);

    if (labels.length !== uniqueLabels.size) {
      const duplicates = labels.filter(
        (label, index) => labels.indexOf(label) !== index,
      );
      const uniqueDuplicates = [...new Set(duplicates)];
      throw new Error(
        `FormCombobox: Duplicate option labels detected - ${uniqueDuplicates.join(", ")}`,
      );
    }
  });

  const value = () =>
    options.find((o) => o.value === field().state.value)?.label ?? undefined;

  const handleChange = (label: string | null) =>
    field().handleChange(
      options.find((o) => o.label === label)?.value ?? undefined,
    );

  const handleInputChange = (input: string) => {
    if (input === "") {
      field().handleChange(undefined);
    }
  };

  return (
    <Combobox
      options={optionLabels()}
      validationState={validationState()}
      placeholder={placeholder}
      name={field().name}
      value={value()}
      onBlur={field().handleBlur}
      onChange={handleChange}
      onInputChange={handleInputChange}
      itemComponent={(props) => (
        <ComboboxItem item={props.item}>{props.item.rawValue}</ComboboxItem>
      )}
    >
      {label && <ComboboxLabel>{label}</ComboboxLabel>}
      <ComboboxControl class={cn(fullWidth && "w-full", props.class)}>
        <ComboboxInput />
        <ComboboxTrigger />
      </ComboboxControl>
      {description && <ComboboxDescription>{description}</ComboboxDescription>}
      <ComboboxErrorMessage>
        {field()
          .state.meta.errors.map((error) => error?.message)
          .join(", ")}
      </ComboboxErrorMessage>
      <ComboboxContent />
    </Combobox>
  );
}

interface FormErrorLabelProps {
  field: () => AnyFieldApi;
  class?: JSX.HTMLAttributes<"div">["class"];
}
function FormErrorLabel(props: FormErrorLabelProps) {
  const { field } = props;

  return (
    <Show when={field().state.meta.errors.length > 0}>
      <span class={cn(textfieldLabel({ error: true }), props.class)}>
        {field()
          .state.meta.errors.map((error) => error?.message)
          .join(", ")}
      </span>
    </Show>
  );
}

export const Form = {
  TextField: FormTextField,
  TextArea: FormTextArea,
  Combobox: FormCombobox,
  ErrorLabel: FormErrorLabel,
} as const;
