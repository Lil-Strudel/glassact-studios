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
  FileFieldErrorMessage,
} from "./filefield";
import { cn } from "./cn";
import {
  IoClose,
  IoCheckmarkCircle,
  IoAlertCircle,
  IoReload,
} from "solid-icons/io";
import { createSignal, Show, Switch, Match, createEffect } from "solid-js";

export interface UploadResponse {
  url: string;
  filename: string;
  size: number;
  content_type: string;
  key: string;
  uploaded_at: string;
}

type UploadStatus = "idle" | "uploading" | "success" | "error";

interface FileUploadItemState {
  file: File;
  status: UploadStatus;
  url?: string;
  error?: string;
  progress: number;
}

export interface FileUploadProps {
  onUrlChange: (url: string | null | string[]) => void;

  accept?: string;
  maxSizeBytes?: number;
  fileTypeLabel?: string;

  multiple?: boolean;

  uploadPath: string;

  label?: string;
  description?: string;
  placeholder?: string;
  class?: string;
  dropzoneClass?: string;
  previewClass?: string;

  initialUrls?: string | string[];
  initialFilenames?: string | string[];

  disabled?: boolean;

  uploadFn?: (file: File, uploadPath: string) => Promise<UploadResponse>;
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 Bytes";
  const k = 1024;
  const sizes = ["Bytes", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + " " + sizes[i];
}

function generateUUID(): string {
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, function (c) {
    const r = (Math.random() * 16) | 0;
    const v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

function parseAcceptString(accept?: string): string[] {
  if (!accept) return [];
  return accept
    .split(",")
    .map((ext) => {
      const trimmed = ext.trim();
      return trimmed.startsWith(".") ? trimmed : `.${trimmed}`;
    })
    .map((ext) => ext.toLowerCase());
}

function validateFile(
  file: File,
  acceptedExtensions: string[],
  maxSize: number,
  fileTypeLabel: string,
): { valid: boolean; error?: string } {
  if (acceptedExtensions.length > 0) {
    const hasValidExtension = acceptedExtensions.some((ext) =>
      file.name.toLowerCase().endsWith(ext),
    );

    if (!hasValidExtension) {
      return {
        valid: false,
        error: `Please select a valid ${fileTypeLabel} file`,
      };
    }
  }

  if (file.size > maxSize) {
    return {
      valid: false,
      error: `File must be smaller than ${formatBytes(maxSize)}`,
    };
  }

  return { valid: true };
}

function simulateFakeProgress(
  onProgress: (percent: number) => void,
): ReturnType<typeof setInterval> {
  const intervals: number[] = [0, 30, 60, 90];
  let current = 0;

  const timer = setInterval(() => {
    if (current < intervals.length - 1) {
      const percent = intervals[current];
      if (percent !== undefined) {
        onProgress(percent);
      }
      current++;
    }
  }, 300);

  return timer;
}

async function uploadFileToS3(
  file: File,
  uploadPath: string,
): Promise<UploadResponse> {
  const formData = new FormData();
  formData.append("file", file);
  formData.append("uploadPath", uploadPath);

  const response = await fetch("/api/upload", {
    method: "POST",
    body: formData,
  });

  if (!response.ok) {
    const error = await response.text();
    throw new Error(error || "Upload failed");
  }

  return response.json();
}

function getUploadFunction(
  customUploadFn?: (file: File, uploadPath: string) => Promise<UploadResponse>,
): (file: File, uploadPath: string) => Promise<UploadResponse> {
  return customUploadFn || uploadFileToS3;
}

export const FileUpload = (props: FileUploadProps) => {
  const [fileStates, setFileStates] = createSignal<
    Map<File, FileUploadItemState>
  >(new Map());

  const [validationError, setValidationError] = createSignal<string | null>(
    null,
  );

  const acceptedExtensions = parseAcceptString(props.accept);
  const maxSizeBytes =
    props.maxSizeBytes !== undefined ? props.maxSizeBytes : 50 * 1024 * 1024;
  const fileTypeLabel = props.fileTypeLabel ?? "file";
  const isMultiple = props.multiple ?? false;

  createEffect(() => {
    if (props.initialUrls) {
      const urls = Array.isArray(props.initialUrls)
        ? props.initialUrls
        : [props.initialUrls];

      const filenames = Array.isArray(props.initialFilenames)
        ? props.initialFilenames
        : [props.initialFilenames ?? ""];

      urls.forEach((url, idx) => {
        const filename = filenames[idx] ?? url.split("/").pop() ?? "file";
        const file = new File([], filename);
        setFileStates((prev) => {
          const next = new Map(prev);
          next.set(file, {
            file,
            status: "success",
            url,
            progress: 100,
          });
          return next;
        });
      });
    }
  });

  async function handleFileAccept(files: File[]) {
    setValidationError(null);

    if (!isMultiple && fileStates().size > 0) {
      setFileStates(new Map());
    }

    for (const file of files) {
      const validation = validateFile(
        file,
        acceptedExtensions,
        maxSizeBytes,
        fileTypeLabel,
      );

      if (!validation.valid) {
        setValidationError(validation.error ?? "Validation failed");
        continue;
      }

      setFileStates((prev) => {
        const next = new Map(prev);
        next.set(file, {
          file,
          status: "uploading",
          progress: 0,
        });
        return next;
      });

      try {
        const progressTimer = simulateFakeProgress((percent: number) => {
          setFileStates((prev) => {
            const next = new Map(prev);
            const current = next.get(file);
            if (current) {
              next.set(file, { ...current, progress: percent });
            }
            return next;
          });
        });

        const uploadFn = getUploadFunction(props.uploadFn);
        const result = await uploadFn(file, props.uploadPath);

        clearInterval(progressTimer);

        setFileStates((prev) => {
          const next = new Map(prev);
          next.set(file, {
            file,
            status: "success",
            url: result.url,
            progress: 100,
          });
          return next;
        });

        const allUrls = Array.from(fileStates().values())
          .map((state) => state.url)
          .filter((url): url is string => !!url);

        if (isMultiple) {
          props.onUrlChange(allUrls);
        } else {
          props.onUrlChange(allUrls[0] ?? null);
        }
      } catch (error) {
        const errorMessage =
          error instanceof Error ? error.message : "Upload failed";

        setFileStates((prev) => {
          const next = new Map(prev);
          next.set(file, {
            file,
            status: "error",
            error: errorMessage,
            progress: 0,
          });
          return next;
        });
      }
    }
  }

  function handleFileChange(details: {
    acceptedFiles: File[];
    rejectedFiles: Array<{
      file: File;
      errors: Array<
        | "TOO_MANY_FILES"
        | "FILE_INVALID_TYPE"
        | "FILE_TOO_LARGE"
        | "FILE_TOO_SMALL"
      >;
    }>;
  }) {
    const currentFiles = new Set(details.acceptedFiles);
    setFileStates((prev) => {
      const next = new Map(prev);
      for (const [file] of next) {
        if (!currentFiles.has(file)) {
          next.delete(file);
        }
      }
      return next;
    });

    if (details.rejectedFiles.length > 0) {
      const rejectionErrors = details.rejectedFiles.map((rf) => {
        if (rf.errors.includes("FILE_TOO_LARGE")) {
          return `${rf.file.name} is too large`;
        }
        if (rf.errors.includes("FILE_INVALID_TYPE")) {
          return `${rf.file.name} is not a valid ${fileTypeLabel}`;
        }
        return `${rf.file.name} was rejected`;
      });
      setValidationError(rejectionErrors.join(", "));
    }
  }

  function getFileState(file: File): FileUploadItemState | undefined {
    return fileStates().get(file);
  }

  function handleDelete(file: File): void {
    setFileStates((prev) => {
      const next = new Map(prev);
      next.delete(file);
      return next;
    });

    const allUrls = Array.from(fileStates().values())
      .filter((state) => state.file !== file)
      .map((state) => state.url)
      .filter((url): url is string => !!url);

    if (isMultiple) {
      props.onUrlChange(allUrls.length > 0 ? allUrls : []);
    } else {
      props.onUrlChange(allUrls[0] ?? null);
    }
  }

  return (
    <FileFieldRoot
      multiple={isMultiple}
      onFileAccept={handleFileAccept}
      onFileChange={handleFileChange}
      disabled={props.disabled}
      class={cn("w-full", props.class)}
    >
      {props.label && <FileFieldLabel>{props.label}</FileFieldLabel>}

      <FileFieldDropzone
        class={cn(
          "flex justify-center items-center gap-2 min-h-[120px] w-full rounded-md border-2 border-dashed border-input bg-muted/30 px-4 py-6 text-sm transition-colors hover:border-primary hover:bg-muted/50 data-[invalid]:border-destructive data-[invalid]:bg-destructive/5",
          props.dropzoneClass,
        )}
      >
        <div class="flex flex-col items-center gap-2 text-center">
          <div class="text-sm font-medium">Drag & drop your files or</div>
          <FileFieldTrigger class="inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring text-primary underline-offset-4 hover:underline">
            Browse
          </FileFieldTrigger>
          <Show when={props.description}>
            <div class="text-xs text-muted-foreground">{props.description}</div>
          </Show>
        </div>
      </FileFieldDropzone>

      <FileFieldHiddenInput />

      <Show when={validationError()}>
        <div class="mt-2 flex items-center gap-2 rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          <IoAlertCircle size={16} class="flex-shrink-0" />
          <div class="flex-1">{validationError()}</div>
          <button
            type="button"
            onClick={() => setValidationError(null)}
            class="flex-shrink-0 p-1 hover:bg-destructive/20 rounded"
          >
            <IoClose size={16} />
          </button>
        </div>
      </Show>

      <FileFieldItemList class={cn("gap-2 mt-4", props.previewClass)}>
        {(file) => {
          const state = () => getFileState(file);
          const status = () => state()?.status ?? "idle";
          const progress = () => state()?.progress ?? 0;
          const error = () => state()?.error;

          return (
            <FileFieldItem
              class={cn(
                "relative overflow-hidden rounded-lg aspect-video bg-black flex items-center justify-center",
                props.previewClass,
              )}
            >
              <FileFieldItemPreviewImage class="absolute inset-0 w-full h-full object-cover" />

              <div
                class={cn(
                  "absolute inset-0 bg-gradient-to-b to-transparent transition-colors",
                  status() === "success" &&
                    "from-green-500/80 via-green-500/20",
                  status() === "error" && "from-red-500/80 via-red-500/20",
                  status() === "uploading" &&
                    "from-blue-500/80 via-blue-500/20",
                  status() === "idle" && "from-gray-500/30 via-gray-500/10",
                )}
              />

              <div class="absolute inset-0 flex flex-col justify-between p-4 text-white z-10">
                <div class="flex justify-between items-start">
                  <div class="flex flex-col">
                    <FileFieldItemName class="font-medium text-xs truncate max-w-[150px]" />
                    <FileFieldItemSize class="text-[10px] opacity-90" />
                  </div>

                  <div class="flex flex-col items-end gap-1">
                    <Switch
                      fallback={
                        <div class="text-xs font-medium text-gray-200">
                          Ready
                        </div>
                      }
                    >
                      <Match when={status() === "uploading"}>
                        <div class="text-xs font-medium flex flex-col items-end gap-1">
                          <div class="flex items-center gap-1">
                            <IoReload size={12} class="animate-spin" />
                            Uploading...
                          </div>
                          <div class="text-[10px] opacity-90">
                            {progress()}%
                          </div>
                        </div>
                      </Match>
                      <Match when={status() === "success"}>
                        <div class="text-xs font-medium">Upload complete</div>
                      </Match>
                      <Match when={status() === "error"}>
                        <div class="text-xs font-medium text-red-300">
                          Upload failed
                        </div>
                      </Match>
                    </Switch>
                  </div>
                </div>

                <div class="flex items-center justify-between">
                  <Show when={status() === "error"}>
                    <div class="text-[10px] text-red-200 max-w-[200px]">
                      {error()}
                    </div>
                  </Show>

                  <div class="flex items-center gap-1 ml-auto">
                    <Show when={status() === "success"}>
                      <IoCheckmarkCircle size={16} class="text-green-300" />
                    </Show>
                    <Show when={status() === "error"}>
                      <IoAlertCircle size={16} class="text-red-300" />
                    </Show>
                    <FileFieldItemDeleteTrigger
                      onClick={() => handleDelete(file)}
                      class="p-1 bg-black/40 hover:bg-black/60 rounded-full transition-colors cursor-pointer"
                    >
                      <IoClose size={16} class="text-white" />
                    </FileFieldItemDeleteTrigger>
                  </div>
                </div>
              </div>

              <Show when={status() === "uploading"}>
                <div class="absolute bottom-0 left-0 right-0 h-1 bg-black/50">
                  <div
                    class="h-full bg-blue-500 transition-all duration-300"
                    style={{ width: `${progress()}%` }}
                  />
                </div>
              </Show>
            </FileFieldItem>
          );
        }}
      </FileFieldItemList>
      <FileFieldErrorMessage />
    </FileFieldRoot>
  );
};
