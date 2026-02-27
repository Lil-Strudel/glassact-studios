import { createFileRoute, Link } from "@tanstack/solid-router";
import { useQuery, useQueryClient } from "@tanstack/solid-query";
import { Show, createMemo, type Component } from "solid-js";
import { Badge, Button, Breadcrumb } from "@glassact/ui";
import type { ProofStatus } from "@glassact/data";
import { getInlayOpts } from "../../queries/inlay";
import { getProjectOpts } from "../../queries/project";
import { getProofsByInlayOpts } from "../../queries/proof";
import { useUserContext } from "../../providers/user";
import { Can } from "../../components/Can";
import ChatThread from "../../components/chat/chat-thread";
import ChatInput from "../../components/chat/chat-input";
import ProofHistory from "../../components/proof/proof-history";
import ProofActions from "../../components/proof/proof-actions";
import CreateProofDialog from "../../components/proof/create-proof-dialog";

export const Route = createFileRoute("/_app/projects_/$id/inlay/$inlayId")({
  component: InlayDetailPage,
});

function proofStatusColor(status: ProofStatus): string {
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

function InlayDetailPage() {
  const params = Route.useParams();
  const userContext = useUserContext();
  const queryClient = useQueryClient();

  const projectQuery = useQuery(() => getProjectOpts(params().id));
  const inlayQuery = useQuery(() => getInlayOpts(params().inlayId));
  const proofsQuery = useQuery(() => getProofsByInlayOpts(params().inlayId));

  const latestPendingProof = createMemo(() => {
    const proofs = proofsQuery.data ?? [];
    const pending = proofs.filter((p) => p.status === "pending");
    return pending.length > 0 ? pending[pending.length - 1] : null;
  });

  const handleProofCreated = () => {
    queryClient.invalidateQueries({ queryKey: ["inlay", params().inlayId] });
    queryClient.invalidateQueries({ queryKey: ["project", params().id] });
  };

  return (
    <div class="space-y-6">
      <Breadcrumb
        crumbs={[
          { title: "Projects", to: "/projects" },
          {
            title: projectQuery.data?.name ?? "Project",
            to: `/projects/${params().id}`,
          },
          {
            title: inlayQuery.data?.name ?? "Inlay",
            to: `/projects/${params().id}/inlay/${params().inlayId}`,
          },
        ]}
      />

      <Show
        when={inlayQuery.data}
        fallback={
          <div class="text-gray-500">
            {inlayQuery.isLoading ? "Loading..." : "Inlay not found"}
          </div>
        }
      >
        {(inlay) => (
          <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
            <div class="lg:col-span-1 space-y-6">
              <div class="border rounded-lg p-4 space-y-4">
                <div class="flex items-center justify-between">
                  <h1 class="text-xl font-semibold">{inlay().name}</h1>
                  <Badge variant="outline">{inlay().type}</Badge>
                </div>

                <Show when={inlay().preview_url}>
                  <div class="border rounded-lg overflow-hidden bg-gray-50 p-2">
                    <img
                      src={inlay().preview_url}
                      alt={inlay().name}
                      class="w-full h-auto max-h-48 object-contain"
                    />
                  </div>
                </Show>

                <Show when={inlay().catalog_info}>
                  {(catalogInfo) => (
                    <div class="text-sm space-y-1">
                      <p class="text-gray-500">
                        Catalog Item ID: {catalogInfo().catalog_item_id}
                      </p>
                      <Show when={catalogInfo().customization_notes}>
                        <p class="text-gray-600">
                          Notes: {catalogInfo().customization_notes}
                        </p>
                      </Show>
                    </div>
                  )}
                </Show>

                <Show when={inlay().custom_info}>
                  {(customInfo) => (
                    <div class="text-sm space-y-1">
                      <p class="text-gray-600">{customInfo().description}</p>
                      <Show
                        when={
                          customInfo().requested_width &&
                          customInfo().requested_height
                        }
                      >
                        <p class="text-gray-500">
                          Requested: {customInfo().requested_width}" x{" "}
                          {customInfo().requested_height}"
                        </p>
                      </Show>
                    </div>
                  )}
                </Show>

                <Show when={inlay().approved_proof_id}>
                  <Badge
                    variant="outline"
                    class="bg-green-50 text-green-700 border-green-200"
                  >
                    Proof Approved
                  </Badge>
                </Show>
              </div>

              <Show when={latestPendingProof()}>
                {(proof) => (
                  <div class="border rounded-lg p-4 space-y-3">
                    <h3 class="text-sm font-semibold">
                      Pending Proof (v{proof().version_number})
                    </h3>
                    <div class="text-xs text-gray-500">
                      <p>
                        {proof().width}" x {proof().height}"
                      </p>
                      <Show when={proof().design_asset_url}>
                        <a
                          href={proof().design_asset_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          class="text-blue-600 underline"
                        >
                          View Design
                        </a>
                      </Show>
                    </div>
                    <Can permission="approve_proof">
                      <ProofActions
                        proof={proof()}
                        inlayUuid={params().inlayId}
                      />
                    </Can>
                  </div>
                )}
              </Show>

              <ProofHistory inlayUuid={params().inlayId} />
            </div>

            <div
              class="lg:col-span-2 border rounded-lg flex flex-col"
              style={{ "min-height": "500px" }}
            >
              <div class="flex items-center justify-between p-4 border-b">
                <h2 class="text-lg font-semibold">Chat</h2>
                <Can permission="create_proof">
                  <CreateProofDialog
                    inlayUuid={params().inlayId}
                    onProofCreated={handleProofCreated}
                  />
                </Can>
              </div>

              <ChatThread
                inlayUuid={params().inlayId}
                projectUuid={params().id}
              />

              <ChatInput inlayUuid={params().inlayId} />
            </div>
          </div>
        )}
      </Show>
    </div>
  );
}
