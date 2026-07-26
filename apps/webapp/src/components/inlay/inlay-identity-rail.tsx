import { Show, createMemo } from "solid-js";
import { Badge, ImageLightbox } from "@glassact/ui";
import type { ColorOverrides, InlayDetail } from "@glassact/data";
import { INSTALLATION_KIT_PRICE_CENTS } from "@glassact/data";
import { formatMoney } from "../../utils/format-money";
import { formatPriceFormula } from "../../utils/format-price-formula";
import {
  INLAY_PHASE_LABELS,
  inlayPhaseBadgeVariant,
  type InlayPhase,
} from "./inlay-phase";
import { InlayGlassSummary } from "./inlay-glass-summary";

interface InlayIdentityRailProps {
  inlay: InlayDetail;
  phase: InlayPhase;
}

// Everything that identifies an inlay and what it costs, in a column that stays
// put across every phase of the inlay's life.
export function InlayIdentityRail(props: InlayIdentityRailProps) {
  // The order snapshot is authoritative once placed — invoices bill from it, so
  // the page must not show a price that drifted away from what was ordered.
  const snapshot = createMemo(() => props.inlay.order_snapshot);

  const designProof = createMemo(
    () => props.inlay.approved_proof ?? props.inlay.latest_proof,
  );

  const previewUrl = createMemo(
    () => designProof()?.design_asset_url || props.inlay.preview_url,
  );

  const dimensions = createMemo(() => {
    const locked = snapshot();
    if (locked) return { width: locked.width, height: locked.height };
    const proof = designProof();
    if (proof) return { width: proof.width, height: proof.height };
    const catalogItem = props.inlay.catalog_item;
    if (catalogItem) {
      return {
        width: catalogItem.default_width,
        height: catalogItem.default_height,
      };
    }
    const custom = props.inlay.custom_info;
    if (custom?.requested_width && custom?.requested_height) {
      return { width: custom.requested_width, height: custom.requested_height };
    }
    return null;
  });

  const priceCents = createMemo(
    () => snapshot()?.price_cents ?? props.inlay.price_cents,
  );

  const priceFormula = createMemo(() => {
    const locked = snapshot();
    if (locked) {
      return formatPriceFormula(
        props.inlay.price_group_name,
        locked.price_adjustment_type,
        locked.price_adjustment_value,
      );
    }
    return formatPriceFormula(
      props.inlay.price_group_name,
      props.inlay.price_adjustment_type,
      props.inlay.price_adjustment_value,
    );
  });

  const colorOverrides = createMemo(
    () => (designProof()?.color_overrides ?? {}) as ColorOverrides,
  );

  return (
    <div class="space-y-4 rounded-lg border p-4">
      <Show
        when={previewUrl()}
        fallback={
          <div class="flex h-40 items-center justify-center rounded-lg bg-gray-100 text-sm text-gray-400">
            No preview
          </div>
        }
      >
        <ImageLightbox
          images={[{ src: previewUrl(), alt: props.inlay.name }]}
          triggerClass="block w-full overflow-hidden rounded-lg border bg-gray-50 p-2"
        >
          <img
            src={previewUrl()}
            alt={props.inlay.name}
            class="mx-auto max-h-56 w-full object-contain"
          />
        </ImageLightbox>
      </Show>

      <div class="space-y-2">
        <h1 class="text-xl font-semibold text-gray-900">{props.inlay.name}</h1>
        <div class="flex flex-wrap items-center gap-2">
          <Badge variant={inlayPhaseBadgeVariant(props.phase)}>
            {INLAY_PHASE_LABELS[props.phase]}
          </Badge>
          <Badge variant="outline">
            {props.inlay.type === "catalog" ? "Catalog" : "Custom"}
          </Badge>
          <Show when={props.inlay.is_customized}>
            <Badge variant="warning">Customized</Badge>
          </Show>
        </div>
      </div>

      <Show when={props.inlay.catalog_item}>
        {(catalogItem) => (
          <div class="space-y-1 border-t pt-3">
            <code class="inline-block rounded bg-gray-100 px-2 py-0.5 font-mono text-xs">
              {catalogItem().catalog_code}
            </code>
            <p class="text-sm text-gray-800">{catalogItem().name}</p>
            <p class="text-xs text-gray-500">{catalogItem().category}</p>
          </div>
        )}
      </Show>

      <div class="space-y-1 border-t pt-3 text-sm">
        <Show when={dimensions()}>
          {(dims) => (
            <p class="text-gray-800">
              {dims().width}" &times; {dims().height}"
            </p>
          )}
        </Show>
        <div class="flex items-baseline justify-between gap-2">
          <span class="font-semibold text-gray-900">
            {priceCents() == null ? "—" : formatMoney(priceCents()! / 100)}
          </span>
          <Show when={priceFormula()}>
            <span class="text-xs text-gray-500">{priceFormula()}</span>
          </Show>
        </div>
        <Show when={props.inlay.installation_kit}>
          <p class="text-xs text-green-700">
            + Installation kit (
            {formatMoney(INSTALLATION_KIT_PRICE_CENTS / 100)})
          </p>
        </Show>
        <Show when={snapshot()}>
          <p class="text-xs text-gray-400">
            Locked in at order time — invoices bill from this.
          </p>
        </Show>
      </div>

      <Show when={props.inlay.is_customized && props.inlay.catalog_item}>
        {(catalogItem) => (
          <div class="border-t pt-3">
            <InlayGlassSummary
              manifest={catalogItem().manifest}
              colorOverrides={colorOverrides()}
            />
          </div>
        )}
      </Show>
    </div>
  );
}
