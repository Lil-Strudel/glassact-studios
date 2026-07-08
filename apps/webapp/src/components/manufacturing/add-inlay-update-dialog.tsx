import { createSignal } from "solid-js";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@glassact/ui";
import { AddInlayUpdateForm } from "./add-inlay-update-form";

interface AddInlayUpdateDialogProps {
  inlayUuid: string;
  triggerLabel: string;
  triggerClass?: string;
}

export function AddInlayUpdateDialog(props: AddInlayUpdateDialogProps) {
  const [open, setOpen] = createSignal(false);

  return (
    <Dialog open={open()} onOpenChange={setOpen}>
      <DialogTrigger
        class={props.triggerClass}
        // Stop the pointer event from reaching the draggable card so opening
        // the dialog never starts a drag.
        onPointerDown={(e: PointerEvent) => e.stopPropagation()}
        onClick={(e: MouseEvent) => e.stopPropagation()}
      >
        {props.triggerLabel}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Add Update</DialogTitle>
        </DialogHeader>
        <AddInlayUpdateForm
          inlayUuid={props.inlayUuid}
          onSuccess={() => setOpen(false)}
        />
      </DialogContent>
    </Dialog>
  );
}
