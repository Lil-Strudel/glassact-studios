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
} from "@glassact/ui";
import { IoClose } from "solid-icons/io";

export const FileUpload = () => {
  function handleFileAccept(file: File[]) {
    console.log(file);
  }

  return (
    <FileFieldRoot
      multiple
      onFileAccept={handleFileAccept}
      onFileReject={(data) => console.log("data reject", data)}
      onFileChange={(data) => console.log("data change", data)}
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
        {(file) => (
          <FileFieldItem class="mt-2 rounded-lg relative overflow-hidden bg-black aspect-video">
            <FileFieldItemPreviewImage class="absolute inset-0 w-full h-full object-cover" />
            <div class="absolute inset-0 bg-gradient-to-b from-green-500/80 via-green-500/20 to-transparent" />
            <div class="absolute inset-0 flex flex-col justify-between p-4 text-white z-10">
              <div class="flex justify-between items-start">
                <div class="flex flex-col">
                  <FileFieldItemName class="font-medium text-xs" />
                  <FileFieldItemSize class="text-[10px] opacity-90" />
                </div>
                <div class="flex align-center gap-2">
                  <div class="flex flex-col items-end">
                    <div class="text-xs font-medium">Upload complete</div>
                    <div class="text-[10px] opacity-90">Tap to delete</div>
                  </div>
                  <div>
                    <FileFieldItemDeleteTrigger class="p-1 bg-black/40 hover:bg-black/60 rounded-full transition-colors">
                      <IoClose size={16} class="text-white" />
                    </FileFieldItemDeleteTrigger>
                  </div>
                </div>
              </div>
            </div>
          </FileFieldItem>
        )}
      </FileFieldItemList>
      <FileFieldDescription>My field Description</FileFieldDescription>
      <FileFieldErrorMessage>My Error message</FileFieldErrorMessage>
    </FileFieldRoot>
  );
};
