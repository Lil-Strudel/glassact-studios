import { Button } from "@glassact/ui";
import type { GET, InlayProof } from "@glassact/data";
import { useQuery } from "@tanstack/solid-query";
import { For, Show, type Component } from "solid-js";
import { IoDownloadOutline } from "solid-icons/io";
import { getProofsByInlayOpts } from "../../queries/proof";
import { ProofStatusBadge } from "./proof-status-badge";

interface ProofHistoryProps {
  inlayUuid: string;
}

const ProofHistory: Component<ProofHistoryProps> = (props) => {
  const query = useQuery(() => getProofsByInlayOpts(props.inlayUuid));

  return (
    <div class="space-y-3">
      <h3 class="text-sm font-semibold text-gray-700">Proof History</h3>
      <Show
        when={!query.isLoading}
        fallback={<p class="text-sm text-gray-500">Loading proofs...</p>}
      >
        <Show
          when={(query.data ?? []).length > 0}
          fallback={<p class="text-sm text-gray-500">No proofs yet.</p>}
        >
          <div class="space-y-3">
            <For each={[...(query.data ?? [])].reverse()}>
              {(proof: GET<InlayProof>) => (
                <div class="border rounded-lg overflow-hidden bg-white shadow-sm">
                  <Show when={proof.design_asset_url}>
                    <div class="bg-gray-50 p-3 flex items-center justify-center border-b">
                      <img
                        src={proof.design_asset_url}
                        alt={`Proof v${proof.version_number}`}
                        class="max-w-full max-h-32 object-contain rounded"
                      />
                    </div>
                  </Show>
                  <div class="p-3 space-y-3">
                    <div class="flex items-center justify-between">
                      <span class="text-sm font-semibold text-gray-900">
                        Version {proof.version_number}
                      </span>
                      <ProofStatusBadge status={proof.status} />
                    </div>

                    <div class="text-xs text-gray-600 space-y-1">
                      <p class="font-medium">
                        {proof.width}" x {proof.height}"
                      </p>
                      <Show when={proof.approved_at}>
                        <p class="text-green-600">
                          Approved:{" "}
                          {new Date(proof.approved_at!).toLocaleDateString()}
                        </p>
                      </Show>
                      <Show when={proof.declined_at}>
                        <p class="text-red-600">
                          Declined:{" "}
                          {new Date(proof.declined_at!).toLocaleDateString()}
                        </p>
                      </Show>
                      <Show when={proof.decline_reason}>
                        <p class="text-red-600 bg-red-50 p-2 rounded border border-red-200">
                          Reason: {proof.decline_reason}
                        </p>
                      </Show>
                      <p class="text-gray-400">
                        Created: {new Date(proof.created_at).toLocaleString()}
                      </p>
                    </div>

                    <Show when={proof.design_asset_url}>
                      <Button
                        variant="outline"
                        size="sm"
                        as="a"
                        href={proof.design_asset_url}
                        download
                        class="w-full"
                      >
                        <IoDownloadOutline class="mr-2" size={16} />
                        Download Design
                      </Button>
                    </Show>
                  </div>
                </div>
              )}
            </For>
          </div>
        </Show>
      </Show>
    </div>
  );
};

export default ProofHistory;
