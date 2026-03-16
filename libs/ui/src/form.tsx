import type { AnyFieldApi } from "@tanstack/solid-form";
import {
  TextFieldRoot,
  TextFieldLabel,
  TextField,
  TextFieldDescription,
  TextFieldErrorMessage,
  textfieldLabel,
} from "./textfield";
import { NumberField } from "./numberfield";
import { cn } from "./cn";
import { createEffect, createSignal, JSX, Show, For } from "solid-js";
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
import {
  FileFieldRoot,
  FileFieldLabel,
  FileFieldDescription,
  FileFieldErrorMessage,
} from "./filefield";
import { FileUpload, type UploadResponse } from "./file-upload";
import { Checkbox, CheckboxControl, CheckboxLabel } from "./checkbox";

function useValidationState(getField: () => AnyFieldApi) {
  const [validationState, setValidationState] = createSignal<
    "valid" | "invalid"
  >("valid");

  createEffect(() => {
    const field = getField();
    if (field.state.meta.errors.length > 0 && field.state.meta.isTouched) {
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
  const validationState = useValidationState(() => props.field());

  return (
    <TextFieldRoot
      class={cn((props.fullWidth ?? true) && "w-full", props.class)}
      validationState={validationState()}
    >
      {props.label && <TextFieldLabel>{props.label}</TextFieldLabel>}
      <TextField
        placeholder={props.placeholder}
        name={props.field().name}
        value={props.field().state.value}
        onBlur={() => props.field().handleBlur()}
        onChange={(e) => props.field().handleChange(e.target.value)}
      />
      {props.description && (
        <TextFieldDescription>{props.description}</TextFieldDescription>
      )}
      <TextFieldErrorMessage>
        {props.field()
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
  const validationState = useValidationState(() => props.field());

  return (
    <TextFieldRoot
      class={cn((props.fullWidth ?? true) && "w-full", props.class)}
      validationState={validationState()}
    >
      {props.label && <TextFieldLabel>{props.label}</TextFieldLabel>}
      <TextArea
        placeholder={props.placeholder}
        name={props.field().name}
        value={props.field().state.value}
        onBlur={() => props.field().handleBlur()}
        onChange={(e) => props.field().handleChange(e.target.value)}
      />
      {props.description && (
        <TextFieldDescription>{props.description}</TextFieldDescription>
      )}
      <TextFieldErrorMessage>
        {props.field()
          .state.meta.errors.map((error) => error?.message)
          .join(", ")}
      </TextFieldErrorMessage>
    </TextFieldRoot>
  );
}

interface FormNumberFieldProps {
  field: () => AnyFieldApi;
  class?: JSX.HTMLAttributes<"div">["class"];
  label?: string;
  placeholder?: string;
  description?: string;
  fullWidth?: boolean;
  int?: boolean;
  decimalPlaces?: number;
}
function FormNumberField(props: FormNumberFieldProps) {
  const validationState = useValidationState(() => props.field());

  const handleChange = (value: string) => {
    if (value === "") {
      props.field().handleChange("");
    } else {
      const numValue = parseFloat(value);
      if (!isNaN(numValue)) {
        props.field().handleChange(numValue);
      }
    }
  };

  return (
    <TextFieldRoot
      class={cn((props.fullWidth ?? true) && "w-full", props.class)}
      validationState={validationState()}
    >
      {props.label && <TextFieldLabel>{props.label}</TextFieldLabel>}
      <NumberField
        placeholder={props.placeholder}
        name={props.field().name}
        value={props.field().state.value?.toString() || ""}
        onBlur={() => props.field().handleBlur()}
        onChange={handleChange}
        int={props.int}
        decimalPlaces={props.decimalPlaces}
      />
      {props.description && (
        <TextFieldDescription>{props.description}</TextFieldDescription>
      )}
      <TextFieldErrorMessage>
        {props.field()
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
  const validationState = useValidationState(() => props.field());

  const optionLabels = () => props.options.map((o) => o.label);

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
    props.options.find((o) => o.value === props.field().state.value)?.label ?? undefined;

  const handleChange = (label: string | null) =>
    props.field().handleChange(
      props.options.find((o) => o.label === label)?.value ?? undefined,
    );

  const handleInputChange = (input: string) => {
    if (input === "") {
      props.field().handleChange(undefined);
    }
  };

  return (
    <Combobox
      options={optionLabels()}
      validationState={validationState()}
      placeholder={props.placeholder}
      name={props.field().name}
      value={value()}
      onBlur={() => props.field().handleBlur()}
      onChange={handleChange}
      onInputChange={handleInputChange}
      itemComponent={(props) => (
        <ComboboxItem item={props.item}>{props.item.rawValue}</ComboboxItem>
      )}
    >
      {props.label && <ComboboxLabel>{props.label}</ComboboxLabel>}
      <ComboboxControl class={cn((props.fullWidth ?? true) && "w-full", props.class)}>
        <ComboboxInput />
        <ComboboxTrigger />
      </ComboboxControl>
      {props.description && <ComboboxDescription>{props.description}</ComboboxDescription>}
      <ComboboxErrorMessage>
        {props.field()
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
  return (
    <Show when={props.field().state.meta.errors.length > 0}>
      <span class={cn(textfieldLabel({ error: true }), props.class)}>
        {props.field()
          .state.meta.errors.map((error) => error?.message)
          .join(", ")}
      </span>
    </Show>
  );
}

interface FormFileUploadProps {
  field: () => AnyFieldApi;
  accept?: string;
  maxSizeBytes?: number;
  fileTypeLabel?: string;
  uploadPath: string;
  multiple?: boolean;
  label?: string;
  placeholder?: string;
  description?: string;
  class?: string;
  fullWidth?: boolean;
  uploadFn?: (params: { file: File; uploadPath: string }) => Promise<UploadResponse>;
}

function FormFileUpload(props: FormFileUploadProps) {
  const validationState = useValidationState(() => props.field());

  return (
    <FileFieldRoot
      class={cn((props.fullWidth ?? true) && "w-full", props.class)}
      data-invalid={validationState() === "invalid"}
    >
      {props.label && <FileFieldLabel>{props.label}</FileFieldLabel>}
      <FileUpload
        onUrlChange={(url) => props.field().handleChange(url)}
        initialUrls={props.field().state.value}
        accept={props.accept}
        maxSizeBytes={props.maxSizeBytes}
        fileTypeLabel={props.fileTypeLabel}
        uploadPath={props.uploadPath}
        multiple={props.multiple}
        placeholder={props.placeholder}
        description={props.description}
        uploadFn={props.uploadFn}
      />
      {props.description && (
        <FileFieldDescription>{props.description}</FileFieldDescription>
      )}
      <FileFieldErrorMessage>
        {props.field()
          .state.meta.errors.map((error) => error?.message)
          .join(", ")}
      </FileFieldErrorMessage>
    </FileFieldRoot>
  );
}

interface FormSelectProps {
  field: () => AnyFieldApi;
  options: { label: string; value: string | number }[];
  class?: JSX.HTMLAttributes<"div">["class"];
  label?: string;
  placeholder?: string;
  description?: string;
  fullWidth?: boolean;
}

function FormSelect(props: FormSelectProps) {
  const validationState = useValidationState(() => props.field());

  return (
    <TextFieldRoot
      class={cn((props.fullWidth ?? true) && "w-full", props.class)}
      validationState={validationState()}
    >
      {props.label && <TextFieldLabel>{props.label}</TextFieldLabel>}
      <select
        value={props.field().state.value || ""}
        onChange={(e) => {
          const val = e.currentTarget.value;
          props.field().handleChange(
            val ? (isNaN(Number(val)) ? val : Number(val)) : undefined,
          );
        }}
        onBlur={() => props.field().handleBlur()}
        class="rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
      >
        <option value="">{props.placeholder ?? "Select..."}</option>
        <For each={props.options}>
          {(opt) => <option value={opt.value}>{opt.label}</option>}
        </For>
      </select>
      {props.description && (
        <TextFieldDescription>{props.description}</TextFieldDescription>
      )}
      <TextFieldErrorMessage>
        {props.field()
          .state.meta.errors.map((error) => error?.message)
          .join(", ")}
      </TextFieldErrorMessage>
    </TextFieldRoot>
  );
}

interface FormCheckboxProps {
  field: () => AnyFieldApi;
  class?: JSX.HTMLAttributes<"div">["class"];
  label?: string;
  description?: string;
}

function FormCheckbox(props: FormCheckboxProps) {
  const validationState = useValidationState(() => props.field());

  return (
    <div class={cn("flex flex-col gap-2", props.class)}>
      <Checkbox
        checked={props.field().state.value}
        onChange={(checked) => props.field().handleChange(checked)}
        onBlur={() => props.field().handleBlur()}
        validationState={validationState()}
      >
        <div class="flex items-center gap-2">
          <CheckboxControl />
          {props.label && <CheckboxLabel>{props.label}</CheckboxLabel>}
        </div>
      </Checkbox>
      {props.description && (
        <p class={cn(textfieldLabel({ description: true, label: false }))}>
          {props.description}
        </p>
      )}
      <Show when={props.field().state.meta.errors.length > 0}>
        <span class={cn(textfieldLabel({ error: true }))}>
          {props.field()
            .state.meta.errors.map((error) => error?.message)
            .join(", ")}
        </span>
      </Show>
    </div>
  );
}

export const Form = {
  TextField: FormTextField,
  TextArea: FormTextArea,
  NumberField: FormNumberField,
  Combobox: FormCombobox,
  FileUpload: FormFileUpload,
  Select: FormSelect,
  Checkbox: FormCheckbox,
  ErrorLabel: FormErrorLabel,
} as const;
