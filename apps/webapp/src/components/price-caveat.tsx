import { cn } from "@glassact/ui";
import { PRICE_CAVEAT } from "../utils/format-money";

interface PriceCaveatProps {
  class?: string;
}

export function PriceCaveat(props: PriceCaveatProps) {
  return (
    <p class={cn("text-xs text-gray-400", props.class)}>{PRICE_CAVEAT}</p>
  );
}
