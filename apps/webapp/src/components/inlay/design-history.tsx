import { For, Show, createMemo, createSignal } from "solid-js";
import { ImageLightbox } from "@glassact/ui";
import { useQuery } from "@tanstack/solid-query";
import { IoChevronDown, IoChevronForward } from "solid-icons/io";
import type { GET, InlayProof } from "@glassact/data";
import { getProofsByInlayOpts } from "../../queries/proof";
import { ProofStatusBadge } from "../proof/proof-status-badge";
import { DownloadDesignButton } from "../proof/download-design-button";
import { mirrorCssTransform, type MirrorFlags } from "../../utils/inlay-mirror";

interface DesignHistoryProps {
  inlayUuid: string;
  mirror: MirrorFlags;
  // The pending proof already gets a full panel of its own, so skip it here
  // rather than rendering the same design twice.
  excludeProofId?: number;
}

function ProofRow(props: {
  proof: GET<InlayProof>;
  defaultOpen: boolean;
  mirror: MirrorFlags;
}) {
  const [isOpen, setIsOpen] = createSignal(props.defaultOpen);

  return (
    <li class="overflow-hidden rounded-lg border border-border bg-card">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen())}
        class="flex w-full items-center gap-2 px-3 py-2.5 text-left transition-colors hover:bg-accent/60"
      >
        <span class="text-faint">
          <Show when={isOpen()} fallback={<IoChevronForward size={14} />}>
            <IoChevronDown size={14} />
          </Show>
        </span>
        <span class="text-sm font-medium text-foreground">
          Version {props.proof.version_number}
        </span>
        <span class="flex-1" />
        <ProofStatusBadge status={props.proof.status} />
      </button>

      <Show when={isOpen()}>
        <div class="space-y-3 border-t border-border p-3">
          <Show when={props.proof.design_asset_url}>
            {(assetUrl) => (
              <ImageLightbox
                images={[
                  {
                    src: assetUrl(),
                    alt: `Proof v${props.proof.version_number}`,
                    transform: mirrorCssTransform(props.mirror),
                  },
                ]}
                triggerClass="flex w-full items-center justify-center rounded bg-muted p-3"
              >
                <img
                  src={assetUrl()}
                  alt={`Proof v${props.proof.version_number}`}
                  class="came-frame max-h-40 max-w-full rounded-md bg-plaque object-contain p-2"
                  style={{ transform: mirrorCssTransform(props.mirror) }}
                />
              </ImageLightbox>
            )}
          </Show>

          <div class="space-y-1 text-xs text-muted-foreground">
            <p class="tabular-nums">
              {props.proof.width}" &times; {props.proof.height}"
            </p>
            <Show when={props.proof.approved_at}>
              <p class="text-success-foreground">
                Approved {new Date(props.proof.approved_at!).toLocaleDateString()}
              </p>
            </Show>
            <Show when={props.proof.declined_at}>
              <p class="text-error-foreground">
                Declined {new Date(props.proof.declined_at!).toLocaleDateString()}
              </p>
            </Show>
            <Show when={props.proof.decline_reason}>
              <p class="rounded border border-error-foreground/30 bg-error p-2 text-error-foreground">
                Reason: {props.proof.decline_reason}
              </p>
            </Show>
            <p class="text-faint">
              Created {new Date(props.proof.created_at).toLocaleString()}
            </p>
          </div>

          <Show when={props.proof.design_asset_url}>
            <DownloadDesignButton proofUuid={props.proof.uuid} />
          </Show>
        </div>
      </Show>
    </li>
  );
}

// Every design version this inlay has had, newest first. A stock or
// once-approved customized inlay has nothing interesting here, so the whole
// section collapses to a single line in that case.
export function DesignHistory(props: DesignHistoryProps) {
  const query = useQuery(() => getProofsByInlayOpts(props.inlayUuid));

  const proofs = createMemo(() =>
    [...(query.data ?? [])]
      .filter((proof) => proof.id !== props.excludeProofId)
      .reverse(),
  );

  const [isExpanded, setIsExpanded] = createSignal(false);
  const isWorthExpanding = createMemo(() => proofs().length > 1);

  return (
    <Show when={proofs().length > 0}>
      <section class="space-y-3">
        <button
          type="button"
          onClick={() => setIsExpanded(!isExpanded())}
          class="flex w-full items-center gap-2 text-left"
        >
          <span class="text-faint">
            <Show when={isExpanded()} fallback={<IoChevronForward size={14} />}>
              <IoChevronDown size={14} />
            </Show>
          </span>
          <h2 class="text-lg text-foreground">
            Design history ({proofs().length})
          </h2>
        </button>

        <Show when={isExpanded()}>
          <ul class="space-y-2">
            <For each={proofs()}>
              {(proof, index) => (
                <ProofRow
                  proof={proof}
                  defaultOpen={index() === 0 && isWorthExpanding()}
                  mirror={props.mirror}
                />
              )}
            </For>
          </ul>
        </Show>
      </section>
    </Show>
  );
}
