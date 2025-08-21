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
import { createEffect, createSignal, JSX, Show } from "solid-js";
import { TextArea } from "./textarea";

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
  ErrorLabel: FormErrorLabel,
} as const;
