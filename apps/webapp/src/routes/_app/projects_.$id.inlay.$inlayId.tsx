import { createFileRoute, Link } from "@tanstack/solid-router";
import { useQuery, useQueryClient } from "@tanstack/solid-query";
import { For, Match, Show, Switch, createMemo, createSignal } from "solid-js";
import { Badge, Breadcrumb, Button, ImageLightbox } from "@glassact/ui";
import type { InlayDetail, ManufacturingStep } from "@glassact/data";
import { PERMISSION_ACTIONS } from "@glassact/data";
import { getInlayOpts } from "../../queries/inlay";
import { getProjectOpts } from "../../queries/project";
import { Can } from "../../components/Can";
import EditCustomInfoForm from "../../components/inlay/edit-custom-info-form";
import { InlayIdentityRail } from "../../components/inlay/inlay-identity-rail";
import { InlayDiscussion } from "../../components/inlay/inlay-discussion";
import { DesignHistory } from "../../components/inlay/design-history";
import { ProofReviewPanel } from "../../components/inlay/proof-review-panel";
import { SandblastFileCard } from "../../components/inlay/sandblast-file-card";
import { InlayTimeline } from "../../components/manufacturing/inlay-timeline";
import { ManufacturingTracker } from "../../components/manufacturing/manufacturing-tracker";
import { STEP_LABELS } from "../../components/manufacturing/steps";
import CreateProofDialog from "../../components/proof/create-proof-dialog";
import { deriveInlayPhase } from "../../components/inlay/inlay-phase";

export const Route = createFileRoute("/_app/projects_/$id/inlay/$inlayId")({
  component: InlayDetailPage,
});

function InlayDetailPage() {
  const params = Route.useParams();
  const queryClient = useQueryClient();

  const projectQuery = useQuery(() => getProjectOpts(params().id));
  const inlayQuery = useQuery(() => getInlayOpts(params().inlayId));

  const inlay = () => (inlayQuery.isSuccess ? inlayQuery.data : null);
  const project = () => (projectQuery.isSuccess ? projectQuery.data : null);

  const phase = createMemo(() => {
    const loadedInlay = inlay();
    const loadedProject = project();
    if (!loadedInlay || !loadedProject) return null;
    return deriveInlayPhase(loadedInlay, loadedProject.status);
  });

  const pendingProof = createMemo(() => {
    const latest = inlay()?.latest_proof;
    return latest?.status === "pending" ? latest : null;
  });

  const isInShop = createMemo(
    () => phase() === "in-production" || phase() === "complete",
  );

  function invalidateInlay() {
    queryClient.invalidateQueries({ queryKey: ["inlay", params().inlayId] });
    queryClient.invalidateQueries({ queryKey: ["project", params().id] });
    queryClient.invalidateQueries({
      queryKey: ["project", params().id, "inlays"],
    });
  }

  return (
    <div class="space-y-6">
      <Breadcrumb
        crumbs={[
          { title: "Projects", to: "/projects" },
          {
            title: project()?.name ?? "Project",
            to: `/projects/${params().id}`,
          },
          {
            title: inlay()?.name ?? "Inlay",
            to: `/projects/${params().id}/inlay/${params().inlayId}`,
          },
        ]}
      />

      <Switch>
        <Match when={inlayQuery.isLoading || projectQuery.isLoading}>
          <div class="grid grid-cols-1 gap-6 lg:grid-cols-[300px_minmax(0,1fr)]">
            <div class="h-96 animate-pulse rounded-lg bg-gray-200" />
            <div class="h-96 animate-pulse rounded-lg bg-gray-200" />
          </div>
        </Match>

        <Match when={inlayQuery.isError || projectQuery.isError}>
          <div class="rounded-xl border-2 border-dashed border-red-300 p-8 text-center">
            <p class="font-medium text-red-600">Failed to load inlay</p>
            <Button
              variant="outline"
              class="mt-4"
              onClick={() => {
                inlayQuery.refetch();
                projectQuery.refetch();
              }}
            >
              Retry
            </Button>
          </div>
        </Match>

        <Match when={inlay() && project() && phase()}>
          <div class="grid grid-cols-1 gap-6 lg:grid-cols-[300px_minmax(0,1fr)]">
            <div class="space-y-4">
              <InlayIdentityRail inlay={inlay()!} phase={phase()!} />
            </div>

            <div class="space-y-6">
              <Switch>
                <Match when={phase() === "awaiting-approval"}>
                  <Show when={pendingProof()}>
                    {(proof) => (
                      <ProofReviewPanel
                        proof={proof()}
                        inlayUuid={params().inlayId}
                      />
                    )}
                  </Show>
                </Match>

                <Match when={isInShop()}>
                  <ManufacturingPanel inlay={inlay()!} />
                </Match>

                <Match when={phase() === "configuring"}>
                  <ConfiguringPanel
                    inlay={inlay()!}
                    inlayUuid={params().inlayId}
                    isDraft={project()!.status === "draft"}
                    onChanged={invalidateInlay}
                  />
                </Match>

                <Match when={phase() === "ready"}>
                  <ReadyPanel
                    inlay={inlay()!}
                    projectUuid={params().id}
                    inlayUuid={params().inlayId}
                  />
                </Match>

                <Match when={phase() === "cancelled"}>
                  <div class="rounded-lg border border-gray-200 bg-gray-50 p-6 text-sm text-gray-600">
                    This project was cancelled, so this inlay will not be
                    manufactured.
                  </div>
                </Match>
              </Switch>

              <Show when={isInShop()}>
                <SandblastFileCard
                  inlayUuid={params().inlayId}
                  inlayName={inlay()!.name}
                  sandblastFileUrl={inlay()!.sandblast_file_url}
                  projectStatus={project()!.status}
                  onUploaded={invalidateInlay}
                  layout="card"
                />
              </Show>

              <DesignHistory
                inlayUuid={params().inlayId}
                excludeProofId={pendingProof()?.id}
              />

              <InlayDiscussion
                inlayUuid={params().inlayId}
                projectUuid={params().id}
                expandByDefault={
                  pendingProof()?.approval_authority === "dealership"
                }
              />
            </div>
          </div>
        </Match>

        <Match when={!inlay()}>
          <div class="rounded-xl border-2 border-dashed border-gray-300 p-8 text-center">
            <p class="text-gray-400">Inlay not found</p>
          </div>
        </Match>
      </Switch>
    </div>
  );
}

function ManufacturingPanel(props: { inlay: InlayDetail }) {
  const step = createMemo(
    () => props.inlay.manufacturing_step as ManufacturingStep | null,
  );

  return (
    <div class="space-y-6">
      <Show when={step()}>
        {(currentStep) => (
          <div class="space-y-3 rounded-lg border p-4">
            <div class="flex items-center justify-between">
              <h2 class="text-base font-semibold text-gray-900">
                {STEP_LABELS[currentStep()]}
              </h2>
              <Badge variant="secondary">In Production</Badge>
            </div>
            <ManufacturingTracker currentStep={currentStep()} />
          </div>
        )}
      </Show>

      <InlayTimeline inlayUuid={props.inlay.uuid} />
    </div>
  );
}

function ConfiguringPanel(props: {
  inlay: InlayDetail;
  inlayUuid: string;
  isDraft: boolean;
  onChanged: () => void;
}) {
  const [isEditing, setIsEditing] = createSignal(false);
  const customInfo = createMemo(() => props.inlay.custom_info);

  return (
    <div class="space-y-6">
      <Show when={customInfo()}>
        {(info) => (
          <div class="space-y-4 rounded-lg border p-4">
            <div class="flex items-start justify-between gap-2">
              <h2 class="text-base font-semibold text-gray-900">
                Design request
              </h2>
              <Show when={props.isDraft && !isEditing()}>
                <Can permission={PERMISSION_ACTIONS.MANAGE_PROJECT}>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => setIsEditing(true)}
                  >
                    Edit details
                  </Button>
                </Can>
              </Show>
            </div>

            <Show
              when={!isEditing()}
              fallback={
                <EditCustomInfoForm
                  inlayUuid={props.inlayUuid}
                  description={info().description}
                  imageUrls={(info().reference_images ?? []).map(
                    (image) => image.image_url,
                  )}
                  onDone={() => {
                    setIsEditing(false);
                    props.onChanged();
                  }}
                />
              }
            >
              <div class="space-y-3">
                <p class="text-sm text-gray-700">{info().description}</p>
                <Show
                  when={info().requested_width && info().requested_height}
                >
                  <p class="text-sm text-gray-500">
                    Requested: {info().requested_width}" &times;{" "}
                    {info().requested_height}"
                  </p>
                </Show>

                <Show when={(info().reference_images ?? []).length > 0}>
                  <div class="space-y-1.5">
                    <p class="text-xs font-medium uppercase tracking-wide text-gray-500">
                      Reference pictures
                    </p>
                    <div class="grid grid-cols-3 gap-2 sm:grid-cols-4">
                      <For each={info().reference_images}>
                        {(image, imageIndex) => (
                          <ImageLightbox
                            images={info().reference_images.map((ref, i) => ({
                              src: ref.image_url,
                              alt: `Reference picture ${i + 1}`,
                            }))}
                            index={imageIndex()}
                            triggerClass="block aspect-square w-full overflow-hidden rounded-md border bg-gray-50"
                          >
                            <img
                              src={image.image_url}
                              alt={`Reference picture ${imageIndex() + 1}`}
                              class="h-full w-full object-cover"
                            />
                          </ImageLightbox>
                        )}
                      </For>
                    </div>
                  </div>
                </Show>
              </div>
            </Show>
          </div>
        )}
      </Show>

      <div class="rounded-lg border border-dashed border-gray-300 p-6 text-center">
        <p class="text-sm font-medium text-gray-500">
          Waiting on a proof from GlassAct
        </p>
        <p class="mt-1 text-sm text-gray-400">
          A designer will send a proof for approval.
        </p>
        <Can permission={PERMISSION_ACTIONS.CREATE_PROOF}>
          <div class="mt-4 flex justify-center">
            <CreateProofDialog
              inlayUuid={props.inlayUuid}
              onProofCreated={props.onChanged}
            />
          </div>
        </Can>
      </div>
    </div>
  );
}

function ReadyPanel(props: {
  inlay: InlayDetail;
  projectUuid: string;
  inlayUuid: string;
}) {
  const canRecustomize = createMemo(
    () => props.inlay.type === "catalog" && props.inlay.is_customized,
  );

  return (
    <div class="space-y-4 rounded-lg border p-4">
      <div class="flex items-center justify-between">
        <h2 class="text-base font-semibold text-gray-900">Ready to order</h2>
        <Badge>Ready</Badge>
      </div>
      <p class="text-sm text-gray-600">
        This inlay is approved and will be included when the order is placed
        from the project page.
      </p>
      <div class="flex flex-wrap gap-2">
        <Button as={Link} to="/projects/$id" params={{ id: props.projectUuid }} variant="outline" size="sm">
          Back to project
        </Button>
        <Show when={canRecustomize()}>
          <Can permission={PERMISSION_ACTIONS.MANAGE_PROJECT}>
            <Button
              as={Link}
              to="/projects/$id/inlay/$inlayId/recustomize"
              params={{ id: props.projectUuid, inlayId: props.inlayUuid }}
              size="sm"
            >
              Adjust design
            </Button>
          </Can>
        </Show>
      </div>
    </div>
  );
}
