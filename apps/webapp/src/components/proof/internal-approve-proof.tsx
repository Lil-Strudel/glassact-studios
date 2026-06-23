import { Show, createMemo } from "solid-js";
import { Button, Form, showToast } from "@glassact/ui";
import { createForm } from "@tanstack/solid-form";
import { useMutation, useQuery } from "@tanstack/solid-query";
import { z } from "zod";
import type { GET, InlayProof, PriceAdjustmentType, PriceGroup } from "@glassact/data";
import { postApproveProofOpts } from "../../queries/proof";
import { getPriceGroupsOpts } from "../../queries/price-group";
import { isApiError } from "../../utils/is-api-error";
import { formatMoney } from "../../utils/format-money";
import { formatPriceFormula } from "../../utils/format-price-formula";
import {
  PRICE_ADJUSTMENT_OPTIONS,
  adjustmentValueFromCents,
  adjustmentValueToCents,
  computeAdjustedPriceCents,
} from "../../utils/proof-pricing";
import PriceGroupCombobox from "../price-group-combobox";

interface InternalApproveProofProps {
  proof: GET<InlayProof>;
  onApproved: () => void;
}

const InternalApproveSchema = z.object({
  price_group_id: z.number().int().optional(),
  price_adjustment_type: z.enum(["none", "percent", "fixed"]),
  price_adjustment_value: z.number().optional(),
});

// Internal review of a customized-catalog proof: the designer/admin picks the
// price group and optionally applies a percent or fixed adjustment before
// approving. The final price is locked into the order snapshot at order time.
export default function InternalApproveProof(props: InternalApproveProofProps) {
  const approveMutation = useMutation(() => postApproveProofOpts());
  const priceGroupsQuery = useQuery(() => getPriceGroupsOpts());

  const form = createForm(() => ({
    defaultValues: {
      price_group_id: props.proof.price_group_id ?? undefined,
      price_adjustment_type: props.proof.price_adjustment_type ?? "none",
      price_adjustment_value:
        props.proof.price_adjustment_type === "none"
          ? undefined
          : adjustmentValueFromCents(
              props.proof.price_adjustment_type,
              props.proof.price_adjustment_value,
            ),
    } as z.output<typeof InternalApproveSchema>,
    validators: {
      onSubmit: InternalApproveSchema,
    },
    onSubmit: async ({ value }) => {
      approveMutation.mutate(
        {
          proofUuid: props.proof.uuid,
          body: {
            price_group_id: value.price_group_id ?? null,
            price_adjustment_type: value.price_adjustment_type,
            price_adjustment_value:
              value.price_adjustment_type === "none"
                ? 0
                : adjustmentValueToCents(
                    value.price_adjustment_type,
                    value.price_adjustment_value ?? 0,
                  ),
          },
        },
        {
          onSuccess() {
            showToast({ title: "Proof approved", variant: "success" });
            props.onApproved();
          },
          onError(error) {
            showToast({
              title: "Failed to approve",
              description: isApiError(error)
                ? error.data?.error
                : "Unknown error",
              variant: "error",
            });
          },
        },
      );
    },
  }));

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
      class="flex flex-col gap-3 border-t pt-3"
    >
      <form.Field
        name="price_group_id"
        children={(field) => <PriceGroupCombobox field={field} />}
      />

      <form.Field
        name="price_adjustment_type"
        children={(typeField) => (
          <div class="flex flex-col gap-3">
            <Form.Select
              field={typeField}
              label="Price Adjustment"
              options={PRICE_ADJUSTMENT_OPTIONS}
            />
            <Show when={typeField().state.value !== "none"}>
              <form.Field
                name="price_adjustment_value"
                children={(valueField) => (
                  <Form.NumberField
                    field={valueField}
                    label={
                      typeField().state.value === "percent"
                        ? "Adjustment (%)"
                        : "Adjustment ($)"
                    }
                    placeholder={
                      typeField().state.value === "percent"
                        ? "e.g., 20 or -10"
                        : "e.g., 12.21"
                    }
                    decimalPlaces={2}
                    allowNegative
                  />
                )}
              />
            </Show>
            <PricePreview
              priceGroupId={form.state.values.price_group_id}
              adjustmentType={typeField().state.value}
              adjustmentValue={form.state.values.price_adjustment_value}
              priceGroups={priceGroupsQuery.data?.items ?? []}
            />
          </div>
        )}
      />

      <Button type="submit" size="sm" disabled={approveMutation.isPending}>
        <Show when={approveMutation.isPending} fallback="Approve">
          Approving...
        </Show>
      </Button>
    </form>
  );
}

interface PricePreviewProps {
  priceGroupId: number | undefined;
  adjustmentType: PriceAdjustmentType;
  adjustmentValue: number | undefined;
  priceGroups: GET<PriceGroup>[];
}

function PricePreview(props: PricePreviewProps) {
  const group = createMemo(() =>
    props.priceGroups.find((pg) => pg.id === props.priceGroupId),
  );

  // The form stores the adjustment in display units; convert to stored units
  // (cents for "fixed") to match the data model and shared compute helper.
  const storedValue = createMemo(() =>
    adjustmentValueToCents(props.adjustmentType, props.adjustmentValue ?? 0),
  );

  const finalCents = createMemo(() => {
    const base = group()?.base_price_cents;
    if (base == null) return null;
    return computeAdjustedPriceCents(base, props.adjustmentType, storedValue());
  });

  const formula = createMemo(() =>
    formatPriceFormula(group()?.name ?? null, props.adjustmentType, storedValue()),
  );

  return (
    <Show when={group()}>
      <div class="rounded-md bg-muted/50 px-3 py-2">
        <Show when={formula()}>
          <p class="text-xs text-muted-foreground">{formula()}</p>
        </Show>
        <p class="text-sm font-medium">
          {finalCents() != null ? formatMoney(finalCents()! / 100) : "—"}
        </p>
      </div>
    </Show>
  );
}
