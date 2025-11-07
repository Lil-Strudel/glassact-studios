import { splitProps, type ValidComponent } from "solid-js";
import { FileField as FileFieldPrimitive } from "@kobalte/core/file-field";
import type {
  FileFieldDropzoneProps,
  FileFieldRootProps,
  FileFieldTriggerProps,
  FileFieldLabelProps,
  FileFieldHiddenInputProps,
  FileFieldItemListProps,
  FileFieldItemPreviewProps,
  FileFieldItemPreviewImageProps,
  FileFieldItemRootProps,
  FileFieldItemNameProps,
  FileFieldItemSizeProps,
  FileFieldItemDeleteTriggerProps,
  FileFieldDescriptionProps,
  FileFieldErrorMessageProps,
} from "@kobalte/core/file-field";
import { PolymorphicProps } from "@kobalte/core/polymorphic";
import { cn } from "./cn";
import { cva } from "class-variance-authority";

type fileFieldProps<T extends ValidComponent = "div"> =
  FileFieldRootProps<T> & {
    class?: string;
  };
export const FileFieldRoot = <T extends ValidComponent = "div">(
  props: PolymorphicProps<T, fileFieldProps<T>>,
) => {
  const [local, rest] = splitProps(props as fileFieldProps, ["class"]);
  return <FileFieldPrimitive class={cn("", local.class)} {...rest} />;
};

type fileFieldDropzoneProps<T extends ValidComponent = "div"> =
  FileFieldDropzoneProps<T> & {
    class?: string;
  };
export const FileFieldDropzone = <T extends ValidComponent = "div">(
  props: PolymorphicProps<T, fileFieldDropzoneProps<T>>,
) => {
  const [local, rest] = splitProps(props as fileFieldDropzoneProps, ["class"]);
  return <FileFieldPrimitive.Dropzone class={cn("", local.class)} {...rest} />;
};

type fileFieldTriggerProps<T extends ValidComponent = "button"> =
  FileFieldTriggerProps<T> & {
    class?: string;
  };
export const FileFieldTrigger = <T extends ValidComponent = "button">(
  props: PolymorphicProps<T, fileFieldTriggerProps<T>>,
) => {
  const [local, rest] = splitProps(props as fileFieldTriggerProps, ["class"]);
  return <FileFieldPrimitive.Trigger class={cn("", local.class)} {...rest} />;
};

export const filefieldLabel = cva(
  "text-sm data-[disabled]:cursor-not-allowed data-[disabled]:opacity-70 font-medium",
  {
    variants: {
      label: {
        true: "data-[invalid]:text-destructive",
      },
      error: {
        true: "text-destructive text-xs",
      },
      description: {
        true: "font-normal text-muted-foreground",
      },
    },
    defaultVariants: {
      label: true,
    },
  },
);

type fileFieldLabelProps<T extends ValidComponent = "label"> =
  FileFieldLabelProps<T> & {
    class?: string;
  };
export const FileFieldLabel = <T extends ValidComponent = "label">(
  props: PolymorphicProps<T, fileFieldLabelProps<T>>,
) => {
  const [local, rest] = splitProps(props as fileFieldLabelProps, ["class"]);
  return (
    <FileFieldPrimitive.Label
      class={cn(filefieldLabel(), local.class)}
      {...rest}
    />
  );
};

type fileFieldHiddenInputProps = FileFieldHiddenInputProps & {
  class?: string;
};
export const FileFieldHiddenInput = (props: fileFieldHiddenInputProps) => {
  const [local, rest] = splitProps(props as fileFieldHiddenInputProps, [
    "class",
  ]);
  return (
    <FileFieldPrimitive.HiddenInput class={cn("", local.class)} {...rest} />
  );
};

type fileFieldItemListProps<T extends ValidComponent = "ul"> =
  FileFieldItemListProps<T> & {
    class?: string;
  };
export const FileFieldItemList = <T extends ValidComponent = "ul">(
  props: PolymorphicProps<T, fileFieldItemListProps<T>>,
) => {
  const [local, rest] = splitProps(props as fileFieldItemListProps, ["class"]);
  return <FileFieldPrimitive.ItemList class={cn("", local.class)} {...rest} />;
};

type fileFieldItemProps<T extends ValidComponent = "li"> =
  FileFieldItemRootProps<T> & {
    class?: string;
  };
export const FileFieldItem = <T extends ValidComponent = "li">(
  props: PolymorphicProps<T, fileFieldItemProps<T>>,
) => {
  const [local, rest] = splitProps(props as fileFieldItemProps, ["class"]);
  return <FileFieldPrimitive.Item class={cn("", local.class)} {...rest} />;
};

type fileFieldItemPreviewProps<T extends ValidComponent = "div"> =
  FileFieldItemPreviewProps<T> & {
    class?: string;
  };
export const FileFieldItemPreview = <T extends ValidComponent = "div">(
  props: PolymorphicProps<T, fileFieldItemPreviewProps<T>>,
) => {
  const [local, rest] = splitProps(props as fileFieldItemPreviewProps, [
    "class",
  ]);
  return (
    <FileFieldPrimitive.ItemPreview class={cn("", local.class)} {...rest} />
  );
};

type fileFieldItemPreviewImageProps<T extends ValidComponent = "img"> =
  FileFieldItemPreviewImageProps<T> & {
    class?: string;
  };
export const FileFieldItemPreviewImage = <T extends ValidComponent = "img">(
  props: PolymorphicProps<T, fileFieldItemPreviewImageProps<T>>,
) => {
  const [local, rest] = splitProps(props as fileFieldItemPreviewImageProps, [
    "class",
  ]);
  return (
    <FileFieldPrimitive.ItemPreviewImage
      class={cn("", local.class)}
      {...rest}
    />
  );
};

type fileFieldItemNameProps<T extends ValidComponent = "span"> =
  FileFieldItemNameProps<T> & {
    class?: string;
  };
export const FileFieldItemName = <T extends ValidComponent = "span">(
  props: PolymorphicProps<T, fileFieldItemNameProps<T>>,
) => {
  const [local, rest] = splitProps(props as fileFieldItemNameProps, ["class"]);
  return <FileFieldPrimitive.ItemName class={cn("", local.class)} {...rest} />;
};

type fileFieldItemSizeProps<T extends ValidComponent = "span"> =
  FileFieldItemSizeProps<T> & {
    class?: string;
  };
export const FileFieldItemSize = <T extends ValidComponent = "span">(
  props: PolymorphicProps<T, fileFieldItemSizeProps<T>>,
) => {
  const [local, rest] = splitProps(props as fileFieldItemSizeProps, ["class"]);
  return <FileFieldPrimitive.ItemSize class={cn("", local.class)} {...rest} />;
};

type fileFieldItemDeleteTriggerProps<T extends ValidComponent = "button"> =
  FileFieldItemDeleteTriggerProps<T> & {
    class?: string;
  };
export const FileFieldItemDeleteTrigger = <T extends ValidComponent = "button">(
  props: PolymorphicProps<T, fileFieldItemDeleteTriggerProps<T>>,
) => {
  const [local, rest] = splitProps(props as fileFieldItemDeleteTriggerProps, [
    "class",
  ]);
  return (
    <FileFieldPrimitive.ItemDeleteTrigger
      class={cn("", local.class)}
      {...rest}
    />
  );
};

type fileFieldDescriptionProps<T extends ValidComponent = "div"> =
  FileFieldDescriptionProps<T> & {
    class?: string;
  };
export const FileFieldDescription = <T extends ValidComponent = "div">(
  props: PolymorphicProps<T, fileFieldDescriptionProps<T>>,
) => {
  const [local, rest] = splitProps(props as fileFieldDescriptionProps, [
    "class",
  ]);
  return (
    <FileFieldPrimitive.Description
      class={cn(
        filefieldLabel({ description: true, label: false }),
        local.class,
      )}
      {...rest}
    />
  );
};

type fileFieldErrorMessageProps<T extends ValidComponent = "div"> =
  FileFieldErrorMessageProps<T> & {
    class?: string;
  };
export const FileFieldErrorMessage = <T extends ValidComponent = "div">(
  props: PolymorphicProps<T, fileFieldErrorMessageProps<T>>,
) => {
  const [local, rest] = splitProps(props as fileFieldErrorMessageProps, [
    "class",
  ]);
  return (
    <FileFieldPrimitive.ErrorMessage
      class={cn(filefieldLabel({ error: true }), local.class)}
      {...rest}
    />
  );
};
