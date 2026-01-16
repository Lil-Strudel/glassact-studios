import {
  FileFieldRoot,
  FileFieldLabel,
  FileFieldDropzone,
  FileFieldTrigger,
  FileFieldHiddenInput,
  FileFieldItemList,
  FileFieldItem,
  FileFieldItemPreviewImage,
  FileFieldItemSize,
  FileFieldItemName,
  FileFieldItemDeleteTrigger,
  FileFieldDescription,
  FileFieldErrorMessage,
  cn,
} from "@glassact/ui";
import {
  IoClose,
  IoCheckmarkCircle,
  IoAlertCircle,
  IoReload,
} from "solid-icons/io";
import { createSignal, Show, Switch, Match } from "solid-js";
import { postUpload, type UploadResponse } from "../queries/upload";

type UploadStatus = "idle" | "uploading" | "success" | "error";

interface FileUploadState {
  file: File;
  status: UploadStatus;
  result?: UploadResponse;
  error?: string;
}

export const FileUpload = () => {
  const [fileStates, setFileStates] = createSignal<Map<File, FileUploadState>>(
    new Map(),
  );

  async function handleFileAccept(files: File[]) {
    for (const file of files) {
      setFileStates((prev) => {
        const next = new Map(prev);
        next.set(file, {
          file,
          status: "uploading",
        });
        return next;
      });

      try {
        const result = await postUpload(file);
        setFileStates((prev) => {
          const next = new Map(prev);
          const current = next.get(file);
          if (current) {
            next.set(file, {
              ...current,
              status: "success",
              result,
            });
          }
          return next;
        });
      } catch (error) {
        const errorMessage =
          error instanceof Error ? error.message : "Upload failed";
        setFileStates((prev) => {
          const next = new Map(prev);
          const current = next.get(file);
          if (current) {
            next.set(file, {
              ...current,
              status: "error",
              error: errorMessage,
            });
          }
          return next;
        });
      }
    }
  }

  function getFileState(file: File): FileUploadState | undefined {
    return fileStates().get(file);
  }

  function handleFileChange(details: {
    acceptedFiles: File[];
    rejectedFiles: {
      file: File;
      errors: (
        | "TOO_MANY_FILES"
        | "FILE_INVALID_TYPE"
        | "FILE_TOO_LARGE"
        | "FILE_TOO_SMALL"
      )[];
    }[];
  }) {
    setFileStates((prev) => {
      const next = new Map(prev);
      const currentFiles = new Set(details.acceptedFiles);
      for (const [file] of next) {
        if (!currentFiles.has(file)) {
          next.delete(file);
        }
      }
      return next;
    });
  }

  return (
    <FileFieldRoot
      multiple
      onFileAccept={handleFileAccept}
      onFileChange={handleFileChange}
    >
      <FileFieldLabel>My Label</FileFieldLabel>
      <FileFieldDropzone class="flex justify-center items-center gap-2 h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-shadow file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-[2px] focus-visible:ring-primary disabled:cursor-not-allowed disabled:opacity-50">
        Drag & Drop your files or
        <FileFieldTrigger class="inline-flex items-center justify-center rounded-md text-sm font-medium transition-[color,background-color,box-shadow] focus-visible:outline-none focus-visible:ring-[1.5px] focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50 text-primary underline-offset-4 hover:underline">
          Browse
        </FileFieldTrigger>
      </FileFieldDropzone>
      <FileFieldHiddenInput />
      <FileFieldItemList>
        {(file) => {
          const state = () => getFileState(file);
          const status = () => state()?.status ?? "idle";
          const result = () => state()?.result;
          const error = () => state()?.error;

          return (
            <FileFieldItem class="mt-2 rounded-lg relative overflow-hidden bg-black aspect-video">
              <FileFieldItemPreviewImage class="absolute inset-0 w-full h-full object-cover" />
              <div
                class={cn(
                  "absolute inset-0 bg-gradient-to-b to-transparent",
                  status() === "success" &&
                    "from-green-500/80 via-green-500/20",
                  status() === "error" && "from-red-500/80 via-red-500/20",
                  status() !== "success" &&
                    status() !== "error" &&
                    "from-blue-500/80 via-blue-500/20",
                )}
              />
              <div class="absolute inset-0 flex flex-col justify-between p-4 text-white z-10">
                <div class="flex justify-between items-start">
                  <div class="flex flex-col">
                    <FileFieldItemName class="font-medium text-xs" />
                    <FileFieldItemSize class="text-[10px] opacity-90" />
                  </div>
                  <div class="flex align-center gap-2">
                    <div class="flex flex-col items-end">
                      <Switch
                        fallback={
                          <div class="text-xs font-medium">Ready to upload</div>
                        }
                      >
                        <Match when={status() === "uploading"}>
                          <div class="text-xs font-medium flex items-center gap-1">
                            <IoReload size={12} class="animate-spin" />
                            Uploading...
                          </div>
                        </Match>
                        <Match when={status() === "success"}>
                          <div class="text-xs font-medium">Upload complete</div>
                          <div class="text-[10px] opacity-90">
                            Tap to delete
                          </div>
                        </Match>
                        <Match when={status() === "error"}>
                          <div class="text-xs font-medium text-red-300">
                            Upload failed
                          </div>
                          <div class="text-[10px] opacity-90">{error()}</div>
                        </Match>
                      </Switch>
                    </div>
                    <div class="flex items-center gap-1">
                      <Show when={status() === "success"}>
                        <IoCheckmarkCircle size={16} class="text-green-300" />
                      </Show>
                      <Show when={status() === "error"}>
                        <IoAlertCircle size={16} class="text-red-300" />
                      </Show>
                      <FileFieldItemDeleteTrigger class="p-1 bg-black/40 hover:bg-black/60 rounded-full transition-colors">
                        <IoClose size={16} class="text-white" />
                      </FileFieldItemDeleteTrigger>
                    </div>
                  </div>
                </div>
              </div>
            </FileFieldItem>
          );
        }}
      </FileFieldItemList>
      <FileFieldDescription>My field Description</FileFieldDescription>
      <FileFieldErrorMessage>My Error message</FileFieldErrorMessage>
    </FileFieldRoot>
  );
};
