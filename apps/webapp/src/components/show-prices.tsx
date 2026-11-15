import { Show, type JSX, type ParentProps } from "solid-js";
import { useCustomerMode } from "../hooks/use-customer-mode";

interface ShowPricesProps {
  fallback?: JSX.Element;
}

export function ShowPrices(props: ParentProps<ShowPricesProps>) {
  const { isCustomerMode } = useCustomerMode();

  return (
    <Show when={!isCustomerMode()} fallback={props.fallback}>
      {props.children}
    </Show>
  );
}
