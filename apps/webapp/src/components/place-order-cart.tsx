import { createMemo, createSignal, For, Show } from "solid-js";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import {
  Badge,
  Button,
  Checkbox,
  CheckboxControl,
  CheckboxLabel,
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  showToast,
} from "@glassact/ui";
import type { InlayWithInfo } from "@glassact/data";
import { INSTALLATION_KIT_PRICE_CENTS } from "@glassact/data";
import { postPlaceOrderOpts } from "../queries/order";
import { patchProjectOpts } from "../queries/project";
import { isApiError } from "../utils/is-api-error";
import { formatMoney } from "../utils/format-money";
import { PriceCaveat } from "./price-caveat";
import { formatPriceFormula } from "../utils/format-price-formula";
import { IoCheckmarkCircle } from "solid-icons/io";

interface PlaceOrderCartProps {
  project: { uuid: string; name: string; installation_kit: boolean };
  inlays: InlayWithInfo[];
  disabled?: boolean;
}

export function PlaceOrderCart(props: PlaceOrderCartProps) {
  const queryClient = useQueryClient();
  const placeOrder = useMutation(() => postPlaceOrderOpts());
  const patchProject = useMutation(() => patchProjectOpts());

  const kitPriceDollars = INSTALLATION_KIT_PRICE_CENTS / 100;

  const [selected, setSelected] = createSignal<Set<string>>(new Set());
  const [open, setOpen] = createSignal(false);
  const [orderSuccess, setOrderSuccess] = createSignal(false);

  const readyInlays = createMemo(() =>
    props.inlays.filter((inlay) => inlay.is_ready),
  );

  function initializeSelection() {
    setSelected(new Set(readyInlays().map((inlay) => inlay.uuid)));
  }

  function toggleInlay(uuid: string) {
    const next = new Set(selected());
    if (next.has(uuid)) {
      next.delete(uuid);
    } else {
      next.add(uuid);
    }
    setSelected(next);
  }

  function dollarsFromCents(cents: number | null | undefined) {
    return (cents ?? 0) / 100;
  }

  const subtotalDollars = createMemo(() => {
    const sel = selected();
    return readyInlays()
      .filter((inlay) => sel.has(inlay.uuid))
      .reduce((sum, inlay) => sum + dollarsFromCents(inlay.price_cents), 0);
  });

  // One kit covers the whole project, so this is a flat charge, not a per-inlay
  // multiple.
  const kitTotalDollars = createMemo(() =>
    props.project.installation_kit ? kitPriceDollars : 0,
  );

  const totalDollars = createMemo(() => subtotalDollars() + kitTotalDollars());

  function handleToggleKit(next: boolean) {
    patchProject.mutate(
      { uuid: props.project.uuid, body: { installation_kit: next } },
      {
        onSuccess() {
          queryClient.invalidateQueries({
            queryKey: ["project", props.project.uuid],
          });
        },
        onError(error) {
          if (isApiError(error)) {
            showToast({
              title: "Failed to update installation kit",
              description: error?.data?.error ?? "Unknown error",
              variant: "error",
            });
          }
        },
      },
    );
  }

  const selectedCount = () => selected().size;

  function handleOpenChange(next: boolean) {
    setOpen(next);
    if (next) {
      initializeSelection();
      setOrderSuccess(false);
      return;
    }

    // Refetching the project flips it out of "draft", and the project page
    // only renders this component while the project is a draft — so
    // invalidating any earlier unmounts the dialog on top of the user,
    // including the confirmation they just placed an order.
    if (orderSuccess()) {
      queryClient.invalidateQueries({ queryKey: ["project"] });
    }
  }

  function handlePlaceOrder() {
    placeOrder.mutate(
      {
        projectUuid: props.project.uuid,
        inlayUuids: Array.from(selected()),
      },
      {
        onSuccess() {
          setOrderSuccess(true);
        },
        onError(error) {
          if (isApiError(error)) {
            showToast({
              title: "Failed to place order",
              description: error?.data?.error ?? "Unknown error",
              variant: "error",
            });
          }
        },
      },
    );
  }

  return (
    <Dialog open={open()} onOpenChange={handleOpenChange}>
      <DialogTrigger as={Button} disabled={props.disabled}>
        Place Order
      </DialogTrigger>
      <DialogContent class="max-w-2xl">
        <Show
          when={!orderSuccess()}
          fallback={
            <div class="text-center py-8 space-y-4">
              <div class="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto">
                <IoCheckmarkCircle class="w-10 h-10 text-green-600" />
              </div>
              <h3 class="text-xl font-semibold text-gray-900">
                Order Placed Successfully!
              </h3>
              <p class="text-gray-600">
                Your order for{" "}
                <span class="font-medium">{props.project.name}</span> has been
                submitted. Manufacturing will begin shortly.
              </p>
              <DialogClose as={Button} class="w-full">
                Close
              </DialogClose>
            </div>
          }
        >
          <DialogHeader>
            <DialogTitle>Review your order</DialogTitle>
          </DialogHeader>

          <Show
            when={readyInlays().length > 0}
            fallback={
              <div class="border-2 border-dashed border-gray-200 rounded-lg p-8 text-center text-sm text-gray-500">
                There are no ready-to-order inlays in this project.
              </div>
            }
          >
            <div class="border rounded-lg divide-y max-h-96 overflow-y-auto">
              <For each={readyInlays()}>
                {(inlay) => {
                  const isSelected = () => selected().has(inlay.uuid);
                  const unitDollars = () => dollarsFromCents(inlay.price_cents);
                  const priceFormula = () =>
                    formatPriceFormula(
                      inlay.price_group_name,
                      inlay.price_adjustment_type,
                      inlay.price_adjustment_value,
                    );
                  const isCatalog = () => inlay.type === "catalog";
                  const catalogCode = () =>
                    inlay.catalog_info ? `Catalog #${inlay.catalog_info.catalog_item_id}` : null;
                  const dims = () => {
                    if (inlay.custom_info?.requested_width && inlay.custom_info?.requested_height) {
                      return `${inlay.custom_info.requested_width}" × ${inlay.custom_info.requested_height}"`;
                    }
                    return null;
                  };

                  return (
                    <Checkbox
                      checked={isSelected()}
                      onChange={() => toggleInlay(inlay.uuid)}
                      class="w-full"
                    >
                      <div class="p-3 flex items-center gap-3 w-full hover:bg-gray-50">
                        <CheckboxLabel class="flex flex-1 min-w-0 items-center gap-3 cursor-pointer">
                          <CheckboxControl />
                          <Show
                            when={inlay.preview_url}
                            fallback={
                              <div class="w-10 h-10 bg-gray-100 rounded flex items-center justify-center text-gray-400 text-xs shrink-0">
                                N/A
                              </div>
                            }
                          >
                            <img
                              src={inlay.preview_url}
                              alt={inlay.name}
                              class="w-10 h-10 object-contain rounded shrink-0"
                              draggable={false}
                            />
                          </Show>
                          <div class="flex-1 min-w-0">
                            <div class="flex items-center gap-2 flex-wrap">
                              <p class="text-sm font-medium truncate">
                                {inlay.name}
                              </p>
                              <Show when={inlay.is_customized}>
                                <Badge variant="warning" class="text-xs">
                                  Customized
                                </Badge>
                              </Show>
                              <Show when={!isCatalog()}>
                                <Badge variant="outline" class="text-xs">
                                  Custom
                                </Badge>
                              </Show>
                            </div>
                            <div class="flex items-center gap-2 text-xs text-gray-500 mt-0.5">
                              <Show when={catalogCode()}>
                                <span>{catalogCode()}</span>
                              </Show>
                              <Show when={dims()}>
                                <span>{dims()}</span>
                              </Show>
                            </div>
                          </div>
                        </CheckboxLabel>

                        <div class="flex flex-col items-end shrink-0 text-sm gap-1">
                          <Show when={priceFormula()}>
                            <span class="text-gray-500 text-xs">
                              {priceFormula()}
                            </span>
                          </Show>
                          <span class="font-semibold">
                            {formatMoney(unitDollars())}
                          </span>
                        </div>
                      </div>
                    </Checkbox>
                  );
                }}
              </For>
            </div>

            <label class="mt-3 flex cursor-pointer items-center gap-3 rounded-lg border p-3 hover:bg-gray-50">
              <input
                type="checkbox"
                class="h-4 w-4 shrink-0 accent-gray-900"
                checked={props.project.installation_kit}
                disabled={patchProject.isPending}
                onChange={(e) => handleToggleKit(e.currentTarget.checked)}
              />
              <span class="flex-1 min-w-0">
                <span class="block text-sm font-medium">
                  Include an installation kit
                </span>
                <span class="block text-xs text-gray-500">
                  Glue, grout and consumables for every inlay in this project.
                </span>
              </span>
              <span class="shrink-0 text-sm font-semibold">
                {formatMoney(kitPriceDollars)}
              </span>
            </label>

            <div class="border-t pt-3 mt-3 space-y-1 text-sm">
              <div class="flex justify-between text-gray-600">
                <span>Subtotal ({selectedCount()} item{selectedCount() === 1 ? "" : "s"})</span>
                <span>{formatMoney(subtotalDollars())}</span>
              </div>
              <Show when={props.project.installation_kit}>
                <div class="flex justify-between text-gray-600">
                  <span>Installation kit</span>
                  <span>{formatMoney(kitTotalDollars())}</span>
                </div>
              </Show>
              <div class="flex justify-between text-base font-semibold">
                <span>Total</span>
                <span>{formatMoney(totalDollars())}</span>
              </div>
              <PriceCaveat class="pt-1 text-right" />
            </div>
          </Show>

          <DialogFooter class="flex justify-end gap-3 mt-4">
            <DialogClose
              as={Button}
              variant="outline"
              disabled={placeOrder.isPending}
            >
              Cancel
            </DialogClose>
            <Button
              onClick={handlePlaceOrder}
              disabled={
                placeOrder.isPending ||
                selectedCount() === 0 ||
                readyInlays().length === 0
              }
            >
              {placeOrder.isPending ? "Placing Order..." : "Place Order"}
            </Button>
          </DialogFooter>
        </Show>
      </DialogContent>
    </Dialog>
  );
}
