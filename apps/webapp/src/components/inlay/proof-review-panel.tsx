import { Show } from "solid-js";
import { ImageLightbox } from "@glassact/ui";
import type { GET, InlayProof } from "@glassact/data";
import { PERMISSION_ACTIONS } from "@glassact/data";
import { Can } from "../Can";
import ProofActions from "../proof/proof-actions";
import { ProofStatusBadge } from "../proof/proof-status-badge";
import { DownloadDesignButton } from "../proof/download-design-button";

interface ProofReviewPanelProps {
  proof: GET<InlayProof>;
  inlayUuid: string;
}

// The pending proof, presented as the thing the page is about. Which side may
// act is decided by the proof's own approval authority: customizer-baked proofs
// are an internal pricing decision, designer proofs are the customer's call.
export function ProofReviewPanel(props: ProofReviewPanelProps) {
  const isInternalAuthority = () => props.proof.approval_authority === "internal";

  return (
    <div class="overflow-hidden rounded-lg border bg-white shadow-sm">
      <div class="flex items-center justify-between border-b px-4 py-3">
        <div>
          <h2 class="text-base font-semibold text-gray-900">
            Proof v{props.proof.version_number}
          </h2>
          <p class="text-xs text-gray-500">
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
              },
            ]}
            triggerClass="flex w-full items-center justify-center border-b bg-gray-50 p-6"
          >
            <img
              src={assetUrl()}
              alt={`Proof v${props.proof.version_number}`}
              class="max-h-96 max-w-full rounded object-contain"
            />
          </ImageLightbox>
        )}
      </Show>

      <div class="space-y-4 p-4">
        <p class="text-sm text-gray-600">
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
