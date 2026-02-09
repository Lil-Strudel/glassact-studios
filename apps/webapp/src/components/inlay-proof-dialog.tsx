import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogTitle,
  DialogDescription,
  DialogHeader,
  Button,
  DialogTriggerProps,
  Form,
  FileUpload,
} from "@glassact/ui";
import { createForm } from "@tanstack/solid-form";
import { useQueryClient } from "@tanstack/solid-query";
import { IoIceCreamOutline } from "solid-icons/io";
import { type Component } from "solid-js";
import { z } from "zod";

const InlayProofDialog: Component = () => {
  const queryClient = useQueryClient();
  const form = createForm(() => ({
    defaultValues: {
      message: "",
    },
    validators: {
      onSubmit: z.object({
        message: z.string(),
      }),
    },
    onSubmit: async ({ value }) => {},
  }));

  return (
    <Dialog>
      <DialogTrigger
        as={(props: DialogTriggerProps) => (
          <Button size="icon" {...props}>
            <IoIceCreamOutline size={24} />
          </Button>
        )}
      />
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Upload a proof!</DialogTitle>
          <DialogDescription>
            Upload everything we need to see if the customer like the design!
          </DialogDescription>
        </DialogHeader>
        <form
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
          class="flex flex-col gap-4"
        >
          <form.Field
            name="message"
            children={(field) => (
              <Form.TextField field={field} label="Message" />
            )}
          />

          <FileUpload
            onUrlChange={(url) => {
              console.log("Proof uploaded:", url);
            }}
            uploadPath="proofs"
            accept=".pdf,.png,.jpg,.jpeg"
            fileTypeLabel="PDF or Image"
            label="Proof File"
            description="Upload your proof design"
          />

          <Button type="submit">Upload</Button>
        </form>
      </DialogContent>
    </Dialog>
  );
};

export default InlayProofDialog;
