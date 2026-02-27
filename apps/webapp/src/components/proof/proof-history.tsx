import { Badge } from "@glassact/ui";
import type { GET, InlayProof, ProofStatus } from "@glassact/data";
import { useQuery } from "@tanstack/solid-query";
import { For, Show, type Component } from "solid-js";
import { getProofsByInlayOpts } from "../../queries/proof";

interface ProofHistoryProps {
  inlayUuid: string;
}

function statusBadgeVariant(status: ProofStatus): "default" | "secondary" | "outline" {
  switch (status) {
    case "approved":
      return "default";
    case "declined":
      return "secondary";
    default:
      return "outline";
  }
}

function statusColor(status: ProofStatus): string {
  switch (status) {
    case "pending":
      return "bg-yellow-50 text-yellow-700 border-yellow-200";
    case "approved":
      return "bg-green-50 text-green-700 border-green-200";
    case "declined":
      return "bg-red-50 text-red-700 border-red-200";
    case "superseded":
      return "bg-gray-50 text-gray-500 border-gray-200";
    default:
      return "";
  }
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
          fallback={
            <p class="text-sm text-gray-500">No proofs yet.</p>
          }
        >
          <div class="space-y-2">
            <For each={[...(query.data ?? [])].reverse()}>
              {(proof: GET<InlayProof>) => (
                <div class="border rounded-lg p-3 space-y-2">
                  <div class="flex items-center justify-between">
                    <span class="text-sm font-medium">
                      v{proof.version_number}
                    </span>
                    <Badge variant="outline" class={statusColor(proof.status)}>
                      {proof.status}
                    </Badge>
                  </div>
                  <div class="text-xs text-gray-500 space-y-1">
                    <p>
                      {proof.width}" x {proof.height}"
                    </p>
                    <Show when={proof.design_asset_url}>
                      <a
                        href={proof.design_asset_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        class="text-blue-600 underline"
                      >
                        View Design
                      </a>
                    </Show>
                    <Show when={proof.approved_at}>
                      <p>Approved: {new Date(proof.approved_at!).toLocaleDateString()}</p>
                    </Show>
                    <Show when={proof.declined_at}>
                      <p>Declined: {new Date(proof.declined_at!).toLocaleDateString()}</p>
                    </Show>
                    <Show when={proof.decline_reason}>
                      <p class="text-red-600">Reason: {proof.decline_reason}</p>
                    </Show>
                    <p class="text-gray-400">
                      Created: {new Date(proof.created_at).toLocaleString()}
                    </p>
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
