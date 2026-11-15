import { cn } from "@glassact/ui";
import { INSTALLATION_KIT_PRICE_CENTS } from "@glassact/data";
import { formatMoney } from "../utils/format-money";

interface InstallationKitNoteProps {
  class?: string;
}

export function InstallationKitNote(props: InstallationKitNoteProps) {
  return (
    <p class={cn("text-xs text-faint", props.class)}>
      Add an installation kit for{" "}
      <span class="tabular-nums">
        {formatMoney(INSTALLATION_KIT_PRICE_CENTS / 100)}
      </span>{" "}
      per project: glue, grout and consumables for every inlay in it.
    </p>
  );
}
