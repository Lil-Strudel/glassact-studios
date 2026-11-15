import { Show } from "solid-js";
import { Button, Tooltip, TooltipContent, TooltipTrigger } from "@glassact/ui";
import { IoPeopleOutline } from "solid-icons/io";
import { useUserContext } from "../providers/user";
import { useCustomerMode } from "../hooks/use-customer-mode";

interface CustomerModeToggleProps {
  size?: "default" | "sm";
}

export function CustomerModeToggle(props: CustomerModeToggleProps) {
  const userContext = useUserContext();
  const { isCustomerMode, setCustomerMode } = useCustomerMode();

  return (
    <Show when={userContext.isDealership()}>
      <Tooltip>
        <TooltipTrigger as="span">
          <Button
            variant={isCustomerMode() ? "gilt" : "outline"}
            size={props.size}
            class="gap-2"
            aria-pressed={isCustomerMode()}
            onClick={() => setCustomerMode(!isCustomerMode())}
          >
            <IoPeopleOutline class="size-4" />
            {isCustomerMode() ? "Customer mode on" : "Customer mode"}
          </Button>
        </TooltipTrigger>
        <TooltipContent class="max-w-xs text-left">
          {isCustomerMode()
            ? "Prices are hidden on every page. Click to show them again."
            : "Hide every price so you can walk a customer through the project."}
        </TooltipContent>
      </Tooltip>
    </Show>
  );
}
