import { createFileRoute, Link } from "@tanstack/solid-router";
import { useQuery, useQueryClient } from "@tanstack/solid-query";
import { For, Match, Show, Switch } from "solid-js";
import {
  Badge,
  Breadcrumb,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@glassact/ui";
import type { ReviewQueueItem } from "@glassact/data";
import { PERMISSION_ACTIONS } from "@glassact/data";
import { getReviewQueueOpts } from "../../queries/review-queue";
import { Can } from "../../components/Can";
import ProofActions from "../../components/proof/proof-actions";
import CreateProofDialog from "../../components/proof/create-proof-dialog";

export const Route = createFileRoute("/_app/review-queue")({
  component: RouteComponent,
});

function RouteComponent() {
  const query = useQuery(() => getReviewQueueOpts());

  const needsApproval = () => (query.isSuccess ? query.data.needs_approval : []);
  const needsProof = () => (query.isSuccess ? query.data.needs_proof : []);
  const isEmpty = () =>
    query.isSuccess &&
    query.data.needs_approval.length === 0 &&
    query.data.needs_proof.length === 0;

  return (
    <Can
      permission={PERMISSION_ACTIONS.INTERNAL_APPROVE_PROOF}
      fallback={() => (
        <div class="border-2 border-dashed border-gray-300 rounded-xl p-8 text-center mt-8">
          <p class="text-gray-500">You don't have access to the review queue.</p>
        </div>
      )}
    >
      <div>
        <Breadcrumb crumbs={[{ title: "Review", to: "/review-queue" }]} />
        <h1 class="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
          Review Queue
        </h1>
        <p class="mt-1 text-sm text-gray-500">
          Inlays across all projects that need internal action.
        </p>

        <Switch>
          <Match when={query.isLoading}>
            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-8">
              <For each={[1, 2, 3]}>
                {() => <div class="h-48 bg-gray-200 rounded-lg animate-pulse" />}
              </For>
            </div>
          </Match>

          <Match when={query.isError}>
            <div class="mt-8 border-2 border-dashed border-red-300 rounded-xl p-8 text-center">
              <p class="text-red-600 font-medium">Failed to load review queue</p>
              <p class="text-gray-500 text-sm mt-1">
                {query.error?.message ?? "An unexpected error occurred."}
              </p>
            </div>
          </Match>

          <Match when={isEmpty()}>
            <div class="mt-8 border-2 border-dashed border-gray-300 rounded-xl p-8 text-center">
              <p class="text-gray-400 text-lg font-medium">All caught up</p>
              <p class="text-gray-400 text-sm mt-2">
                Nothing is waiting on internal review right now.
              </p>
            </div>
          </Match>

          <Match when={query.isSuccess}>
            <div class="flex flex-col gap-12 mt-8">
              <Show when={needsApproval().length > 0}>
                <section>
                  <div class="flex items-center gap-3">
                    <h2 class="text-xl font-bold tracking-tight text-gray-900">
                      Needs Approval
                    </h2>
                    <Badge variant="warning">{needsApproval().length}</Badge>
                  </div>
                  <p class="mt-1 text-sm text-gray-500">
                    Customized catalog inlays awaiting your pricing approval.
                  </p>
                  <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 mt-4">
                    <For each={needsApproval()}>
                      {(item) => <ApprovalCard item={item} />}
                    </For>
                  </div>
                </section>
              </Show>

              <Show when={needsProof().length > 0}>
                <section>
                  <div class="flex items-center gap-3">
                    <h2 class="text-xl font-bold tracking-tight text-gray-900">
                      Needs Proof
                    </h2>
                    <Badge variant="outline">{needsProof().length}</Badge>
                  </div>
                  <p class="mt-1 text-sm text-gray-500">
                    Custom inlays that still need a proof created.
                  </p>
                  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-4">
                    <For each={needsProof()}>
                      {(item) => <NeedsProofCard item={item} />}
                    </For>
                  </div>
                </section>
              </Show>
            </div>
          </Match>
        </Switch>
      </div>
    </Can>
  );
}

function ItemHeader(props: { item: ReviewQueueItem }) {
  return (
    <div class="flex items-start justify-between gap-2">
      <div class="min-w-0">
        <CardTitle class="text-sm truncate">{props.item.inlay.name}</CardTitle>
        <Link
          to="/projects/$id"
          params={{ id: props.item.project_uuid }}
          class="text-xs text-primary hover:underline"
        >
          {props.item.project_name}
        </Link>
      </div>
      <Link
        to="/projects/$id/inlay/$inlayId"
        params={{
          id: props.item.project_uuid,
          inlayId: props.item.inlay.uuid,
        }}
        class="text-xs text-gray-500 hover:underline flex-shrink-0"
      >
        Open
      </Link>
    </div>
  );
}

function ApprovalCard(props: { item: ReviewQueueItem }) {
  return (
    <Card class="overflow-hidden">
      <Show when={props.item.pending_proof?.design_asset_url}>
        <div class="bg-gray-50 p-4 flex items-center justify-center h-40 overflow-hidden border-b">
          <img
            src={props.item.pending_proof!.design_asset_url}
            alt={props.item.inlay.name}
            class="max-w-full max-h-full object-contain"
          />
        </div>
      </Show>
      <CardHeader>
        <ItemHeader item={props.item} />
      </CardHeader>
      <CardContent>
        <Show
          when={props.item.pending_proof}
          fallback={
            <p class="text-sm text-gray-500">
              No pending proof found for this inlay.
            </p>
          }
        >
          {(proof) => (
            <ProofActions proof={proof()} inlayUuid={props.item.inlay.uuid} />
          )}
        </Show>
      </CardContent>
    </Card>
  );
}

function NeedsProofCard(props: { item: ReviewQueueItem }) {
  const queryClient = useQueryClient();

  return (
    <Card class="overflow-hidden">
      <CardHeader>
        <ItemHeader item={props.item} />
      </CardHeader>
      <CardContent class="space-y-3">
        <Show when={props.item.inlay.custom_info?.description}>
          <p class="text-xs text-gray-600 line-clamp-3">
            {props.item.inlay.custom_info!.description}
          </p>
        </Show>
        <CreateProofDialog
          inlayUuid={props.item.inlay.uuid}
          onProofCreated={() =>
            queryClient.invalidateQueries({ queryKey: ["review-queue"] })
          }
        />
      </CardContent>
    </Card>
  );
}
