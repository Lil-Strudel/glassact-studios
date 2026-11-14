import { createFileRoute, Link } from "@tanstack/solid-router";
import { useQuery, useQueryClient } from "@tanstack/solid-query";
import { For, Match, Show, Switch } from "solid-js";
import {
  Alert,
  Badge,
  Breadcrumb,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@glassact/ui";
import { IoAlertCircleOutline } from "solid-icons/io";
import type { ReviewQueueItem } from "@glassact/data";
import { PERMISSION_ACTIONS } from "@glassact/data";
import { getReviewQueueOpts } from "../../queries/review-queue";
import { Can } from "../../components/Can";
import { MirrorBadge } from "../../components/inlay/mirror-badge";
import { mirrorCssTransform } from "../../utils/inlay-mirror";
import ProofActions from "../../components/proof/proof-actions";
import CreateProofDialog from "../../components/proof/create-proof-dialog";
import { PageHeader } from "@glassact/ui";
import { EmptyState } from "@glassact/ui";

export const Route = createFileRoute("/_app/review-queue")({
  component: RouteComponent,
});

function RouteComponent() {
  const query = useQuery(() => getReviewQueueOpts());

  const needsApproval = () =>
    query.isSuccess ? query.data.needs_approval : [];
  const needsProof = () => (query.isSuccess ? query.data.needs_proof : []);
  const isEmpty = () =>
    query.isSuccess &&
    query.data.needs_approval.length === 0 &&
    query.data.needs_proof.length === 0;

  return (
    <Can
      permission={PERMISSION_ACTIONS.INTERNAL_APPROVE_PROOF}
      fallback={() => (
        <EmptyState
          class="mt-8 rounded-lg border border-dashed border-input"
          title="You don't have access to the review queue."
        />
      )}
    >
      <div class="flex flex-col gap-8">
        <div class="space-y-3">
          <Breadcrumb crumbs={[{ title: "Review", to: "/review-queue" }]} />
          <PageHeader
            eyebrow="Design studio"
            title="Review Queue"
            description="Inlays across all projects that need internal action."
          />
        </div>

        <Switch>
          <Match when={query.isLoading}>
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              <For each={[1, 2, 3]}>
                {() => <div class="h-48 animate-pulse rounded-lg bg-card" />}
              </For>
            </div>
          </Match>

          <Match when={query.isError}>
            <Alert variant="destructive">
              <IoAlertCircleOutline size={18} />
              <div>
                <p class="font-medium">Failed to load review queue</p>
                <p class="mt-1 text-muted-foreground">
                  {query.error?.message ?? "An unexpected error occurred."}
                </p>
              </div>
            </Alert>
          </Match>

          <Match when={isEmpty()}>
            <EmptyState
              class="rounded-lg border border-dashed border-input"
              title="All caught up"
              description="Nothing is waiting on internal review right now."
            />
          </Match>

          <Match when={query.isSuccess}>
            <div class="flex flex-col gap-12">
              <Show when={needsApproval().length > 0}>
                <section>
                  <div class="flex items-center gap-3">
                    <h2 class="text-2xl">Needs Approval</h2>
                    <Badge variant="gilt" class="tabular-nums">
                      {needsApproval().length}
                    </Badge>
                  </div>
                  <p class="mt-1 text-sm text-muted-foreground">
                    Customized catalog inlays awaiting your pricing approval.
                  </p>
                  <div class="mt-5 grid grid-cols-1 gap-4 lg:grid-cols-2">
                    <For each={needsApproval()}>
                      {(item) => <ApprovalCard item={item} />}
                    </For>
                  </div>
                </section>
              </Show>

              <Show when={needsProof().length > 0}>
                <section>
                  <div class="flex items-center gap-3">
                    <h2 class="text-2xl">Needs Proof</h2>
                    <Badge variant="outline" class="tabular-nums">
                      {needsProof().length}
                    </Badge>
                  </div>
                  <p class="mt-1 text-sm text-muted-foreground">
                    Custom inlays that still need a proof created.
                  </p>
                  <div class="mt-5 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
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
        <CardTitle class="truncate leading-snug">
          {props.item.inlay.name}
        </CardTitle>
        <MirrorBadge inlay={props.item.inlay} />
        <Link
          to="/projects/$id"
          params={{ id: props.item.project_uuid }}
          class="text-xs text-crimson-light underline-offset-4 hover:underline"
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
        class="flex-shrink-0 text-xs text-faint underline-offset-4 hover:text-foreground hover:underline"
      >
        Open
      </Link>
    </div>
  );
}

function ApprovalCard(props: { item: ReviewQueueItem }) {
  return (
    <Card class="overflow-hidden border-gilt/45">
      <Show when={props.item.pending_proof?.design_asset_url}>
        <div class="flex h-40 items-center justify-center overflow-hidden border-b border-border bg-plaque p-4">
          <img
            src={props.item.pending_proof!.design_asset_url}
            alt={props.item.inlay.name}
            class="max-w-full max-h-full object-contain"
            style={{ transform: mirrorCssTransform(props.item.inlay) }}
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
            <p class="text-sm text-muted-foreground">
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
          <p class="line-clamp-3 text-sm text-muted-foreground">
            {props.item.inlay.custom_info!.description}
          </p>
        </Show>
        <CreateProofDialog
          inlayUuid={props.item.inlay.uuid}
          requestedWidth={props.item.inlay.custom_info?.requested_width}
          requestedHeight={props.item.inlay.custom_info?.requested_height}
          onProofCreated={() =>
            queryClient.invalidateQueries({ queryKey: ["review-queue"] })
          }
        />
      </CardContent>
    </Card>
  );
}
