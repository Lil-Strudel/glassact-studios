import { Show } from "solid-js";
import { ImageLightbox } from "@glassact/ui";
import type { GET, InlayProof } from "@glassact/data";
import { PERMISSION_ACTIONS } from "@glassact/data";
import { Can } from "../Can";
import ProofActions from "../proof/proof-actions";
import { ProofStatusBadge } from "../proof/proof-status-badge";
import { DownloadDesignButton } from "../proof/download-design-button";
import { mirrorCssTransform, type MirrorFlags } from "../../utils/inlay-mirror";

interface ProofReviewPanelProps {
  proof: GET<InlayProof>;
  inlayUuid: string;
  mirror: MirrorFlags;
}

// The pending proof, presented as the thing the page is about. Which side may
// act is decided by the proof's own approval authority: customizer-baked proofs
// are an internal pricing decision, designer proofs are the customer's call.
export function ProofReviewPanel(props: ProofReviewPanelProps) {
  const isInternalAuthority = () => props.proof.approval_authority === "internal";

  return (
    <div
      class={`overflow-hidden rounded-lg border bg-card ${
        isInternalAuthority() ? "border-border" : "border-gilt/45"
      }`}
    >
      <div class="flex items-center justify-between gap-3 border-b border-border px-5 py-4">
        <div class="flex flex-col gap-0.5">
          <h2 class="text-xl text-foreground">
            Proof v{props.proof.version_number}
          </h2>
          <p class="text-sm text-muted-foreground">
            {isInternalAuthority()
              ? "Awaiting internal pricing approval"
              : "Awaiting your approval"}
          </p>
        </div>
        <ProofStatusBadge status={props.proof.status} />
      </div>

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
            triggerClass="flex w-full items-center justify-center border-b border-border bg-muted p-6"
          >
            <img
              src={assetUrl()}
              alt={`Proof v${props.proof.version_number}`}
              class="came-frame max-h-96 max-w-full rounded-lg bg-plaque object-contain p-6"
              style={{ transform: mirrorCssTransform(props.mirror) }}
            />
          </ImageLightbox>
        )}
      </Show>

      <div class="space-y-4 p-5">
        <p class="text-sm tabular-nums text-muted-foreground">
          {props.proof.width}" &times; {props.proof.height}"
        </p>

        <Show when={props.proof.design_asset_url}>
          <DownloadDesignButton proofUuid={props.proof.uuid} />
        </Show>

        <Show
          when={isInternalAuthority()}
          fallback={
            <Can permission={PERMISSION_ACTIONS.APPROVE_PROOF}>
              <ProofActions proof={props.proof} inlayUuid={props.inlayUuid} />
            </Can>
          }
        >
          <Can permission={PERMISSION_ACTIONS.INTERNAL_APPROVE_PROOF}>
            <ProofActions proof={props.proof} inlayUuid={props.inlayUuid} />
          </Can>
        </Show>
      </div>
    </div>
  );
}
