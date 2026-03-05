import { createFileRoute } from "@tanstack/solid-router";
import { useQuery, useQueryClient } from "@tanstack/solid-query";
import { Show, createMemo } from "solid-js";
import { Badge, Breadcrumb, Button } from "@glassact/ui";
import { IoDownloadOutline } from "solid-icons/io";
import { getInlayOpts } from "../../queries/inlay";
import { getProjectOpts } from "../../queries/project";
import { getProofsByInlayOpts } from "../../queries/proof";
import { Can } from "../../components/Can";
import ChatThread from "../../components/chat/chat-thread";
import ChatInput from "../../components/chat/chat-input";
import ProofHistory from "../../components/proof/proof-history";
import ProofActions from "../../components/proof/proof-actions";
import CreateProofDialog from "../../components/proof/create-proof-dialog";
import { ProofStatusBadge } from "../../components/proof/proof-status-badge";

export const Route = createFileRoute("/_app/projects_/$id/inlay/$inlayId")({
  component: InlayDetailPage,
});

function InlayDetailPage() {
  const params = Route.useParams();
  const queryClient = useQueryClient();

  const projectQuery = useQuery(() => getProjectOpts(params().id));
  const inlayQuery = useQuery(() => getInlayOpts(params().inlayId));
  const proofsQuery = useQuery(() => getProofsByInlayOpts(params().inlayId));

  const latestPendingProof = createMemo(() => {
    const proofs = proofsQuery.isSuccess ? proofsQuery.data : [];
    const pending = proofs.filter((p) => p.status === "pending");
    return pending.length > 0 ? pending[pending.length - 1] : null;
  });

  const handleProofCreated = () => {
    queryClient.invalidateQueries({ queryKey: ["inlay", params().inlayId] });
    queryClient.invalidateQueries({ queryKey: ["project", params().id] });
  };

  const inlay = () => inlayQuery.isSuccess && inlayQuery.data;

  return (
    <div class="space-y-6">
      <Breadcrumb
        crumbs={[
          { title: "Projects", to: "/projects" },
          {
            title: projectQuery.isSuccess ? projectQuery.data.name : "Project",
            to: `/projects/${params().id}`,
          },
          {
            title: inlayQuery.isSuccess ? inlayQuery.data.name : "Inlay",
            to: `/projects/${params().id}/inlay/${params().inlayId}`,
          },
        ]}
      />

      <Show
        when={inlay()}
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
                  <ProofStatusBadge status="approved" />
                </Show>
              </div>

              <Show when={latestPendingProof()}>
                {(proof) => (
                  <div class="border rounded-lg overflow-hidden bg-white shadow-sm">
                    <Show when={proof().design_asset_url}>
                      <div class="bg-gray-50 p-4 flex items-center justify-center border-b">
                        <img
                          src={proof().design_asset_url}
                          alt={`Proof v${proof().version_number}`}
                          class="max-w-full max-h-48 object-contain rounded"
                        />
                      </div>
                    </Show>
                    <div class="p-4 space-y-3">
                      <div class="flex items-center justify-between">
                        <h3 class="text-sm font-semibold text-gray-900">
                          Pending Proof (v{proof().version_number})
                        </h3>
                        <ProofStatusBadge status="pending" />
                      </div>
                      <p class="text-sm text-gray-600">
                        {proof().width}" x {proof().height}"
                      </p>
                      <Show when={proof().design_asset_url}>
                        <Button
                          variant="outline"
                          size="sm"
                          as="a"
                          href={proof().design_asset_url}
                          download
                          class="w-full"
                        >
                          <IoDownloadOutline class="mr-2" size={16} />
                          Download Design
                        </Button>
                      </Show>
                      <Can permission="approve_proof">
                        <ProofActions
                          proof={proof()}
                          inlayUuid={params().inlayId}
                        />
                      </Can>
                    </div>
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
