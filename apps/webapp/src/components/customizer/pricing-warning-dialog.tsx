import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  Button,
} from "@glassact/ui";
import { createSignal } from "solid-js";

interface PricingWarningDialogProps {
  open: boolean;
  onContinue: (dontRemind: boolean) => void;
  onCancel: () => void;
}

// Shown when the user switches to Individual-pieces mode. Recoloring single
// pieces diverges from a color group, which requires GlassAct to cut custom
// manufacturing files and can raise the inlay's price.
export function PricingWarningDialog(props: PricingWarningDialogProps) {
  const [dontRemind, setDontRemind] = createSignal(false);

  return (
    <Dialog
      open={props.open}
      onOpenChange={(open) => {
        if (!open) props.onCancel();
      }}
    >
      <DialogContent class="max-w-md">
        <DialogHeader>
          <DialogTitle>Recoloring individual pieces</DialogTitle>
          <DialogDescription>
            Changing the color of a single piece (instead of a whole color
            group) means it no longer matches its group. GlassAct has to create
            new manufacturing files for custom pieces, which{" "}
            <span class="font-medium text-gray-900">
              may increase the price
            </span>{" "}
            of this inlay.
          </DialogDescription>
        </DialogHeader>

        <label class="flex items-center gap-2 text-sm text-gray-600">
          <input
            type="checkbox"
            checked={dontRemind()}
            onChange={(e) => setDontRemind(e.currentTarget.checked)}
            class="h-4 w-4 rounded border-gray-300"
          />
          Don't remind me again this session
        </label>

        <DialogFooter>
          <Button variant="outline" onClick={() => props.onCancel()}>
            Cancel
          </Button>
          <Button onClick={() => props.onContinue(dontRemind())}>
            Continue
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
