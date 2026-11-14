import { Show, createMemo } from "solid-js";
import { Badge, ImageLightbox } from "@glassact/ui";
import type { ColorOverrides, InlayDetail } from "@glassact/data";
import { formatMoney } from "../../utils/format-money";
import { PriceCaveat } from "../price-caveat";
import { formatPriceFormula } from "../../utils/format-price-formula";
import {
  INLAY_PHASE_LABELS,
  inlayPhaseBadgeVariant,
  type InlayPhase,
} from "./inlay-phase";
import { InlayGlassSummary } from "./inlay-glass-summary";
import { MirrorBadge } from "./mirror-badge";
import { mirrorCssTransform } from "../../utils/inlay-mirror";

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
    <div class="space-y-4 rounded-lg border border-border bg-card p-4">
      <Show
        when={previewUrl()}
        fallback={
          <div class="came-frame flex h-48 items-center justify-center rounded-arch bg-came text-sm text-faint">
            No preview
          </div>
        }
      >
        <ImageLightbox
          images={[
            {
              src: previewUrl(),
              alt: props.inlay.name,
              transform: mirrorCssTransform(props.inlay),
            },
          ]}
          triggerClass="came-frame block w-full overflow-hidden rounded-arch bg-plaque px-4 pb-4 pt-8"
        >
          <img
            src={previewUrl()}
            alt={props.inlay.name}
            class="mx-auto max-h-56 w-full object-contain"
            style={{ transform: mirrorCssTransform(props.inlay) }}
          />
        </ImageLightbox>
      </Show>

      <div class="space-y-2">
        <h1 class="text-3xl text-foreground">{props.inlay.name}</h1>
        <div class="flex flex-wrap items-center gap-2">
          <Badge variant={inlayPhaseBadgeVariant(props.phase)} lozenge>
            {INLAY_PHASE_LABELS[props.phase]}
          </Badge>
          <Badge variant="outline">
            {props.inlay.type === "catalog" ? "Catalog" : "Custom"}
          </Badge>
          <Show when={props.inlay.is_customized}>
            <Badge variant="outline">Customized</Badge>
          </Show>
          <MirrorBadge inlay={props.inlay} />
        </div>
      </div>

      <Show when={props.inlay.catalog_item}>
        {(catalogItem) => (
          <div class="space-y-1 border-t border-border pt-3">
            <code class="inline-block rounded bg-accent px-2 py-0.5 font-mono text-xs tracking-wide text-muted-foreground">
              {catalogItem().catalog_code}
            </code>
            <p class="text-sm text-foreground">{catalogItem().name}</p>
            <p class="text-xs text-faint">{catalogItem().category}</p>
          </div>
        )}
      </Show>

      <div class="space-y-1 border-t border-border pt-3 text-sm">
        <Show when={dimensions()}>
          {(dims) => (
            <p class="tabular-nums text-muted-foreground">
              {dims().width}" &times; {dims().height}"
            </p>
          )}
        </Show>
        <div class="flex items-baseline justify-between gap-2">
          <span class="font-display text-2xl tabular-nums text-foreground">
            {priceCents() == null ? "—" : formatMoney(priceCents()! / 100)}
          </span>
          <Show when={priceFormula()}>
            <span class="text-xs tabular-nums text-faint">{priceFormula()}</span>
          </Show>
        </div>
        <Show when={snapshot()}>
          <p class="text-xs text-faint">
            Locked in at order time — invoices bill from this.
          </p>
        </Show>
        <PriceCaveat />
      </div>

      <Show when={props.inlay.is_customized && props.inlay.catalog_item}>
        {(catalogItem) => (
          <div class="border-t border-border pt-3">
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
