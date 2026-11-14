import { createFileRoute, Link, useNavigate } from "@tanstack/solid-router";
import { useQuery, useQueryClient } from "@tanstack/solid-query";
import { For, Match, Show, Switch, createMemo, createSignal } from "solid-js";
import { Badge, Breadcrumb, Button, ImageLightbox } from "@glassact/ui";
import type { InlayDetail, ManufacturingStep } from "@glassact/data";
import { PERMISSION_ACTIONS } from "@glassact/data";
import {
  IoAlertCircleOutline,
  IoChevronBack,
  IoChevronForward,
} from "solid-icons/io";
import { getInlayOpts, getInlaysByProjectOpts } from "../../queries/inlay";
import { getProjectOpts } from "../../queries/project";
import { Can } from "../../components/Can";
import EditCustomInfoForm from "../../components/inlay/edit-custom-info-form";
import { InlayIdentityRail } from "../../components/inlay/inlay-identity-rail";
import { ProjectDiscussion } from "../../components/chat/project-discussion";
import { DesignHistory } from "../../components/inlay/design-history";
import { ProofReviewPanel } from "../../components/inlay/proof-review-panel";
import { SandblastFileCard } from "../../components/inlay/sandblast-file-card";
import { InlayTimeline } from "../../components/manufacturing/inlay-timeline";
import { ManufacturingTracker } from "../../components/manufacturing/manufacturing-tracker";
import { STEP_LABELS } from "../../components/manufacturing/steps";
import CreateProofDialog from "../../components/proof/create-proof-dialog";
import { deriveInlayPhase } from "../../components/inlay/inlay-phase";
import { InlayArrangeActions } from "../../components/inlay/inlay-arrange-actions";

export const Route = createFileRoute("/_app/projects_/$id/inlay/$inlayId")({
  component: InlayDetailPage,
});

function InlayDetailPage() {
  const params = Route.useParams();
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  const projectQuery = useQuery(() => getProjectOpts(params().id));
  const inlayQuery = useQuery(() => getInlayOpts(params().inlayId));
  const siblingsQuery = useQuery(() => getInlaysByProjectOpts(params().id));

  const inlay = () => (inlayQuery.isSuccess ? inlayQuery.data : null);
  const project = () => (projectQuery.isSuccess ? projectQuery.data : null);

  // Projects routinely carry a dozen inlays and reviewing them one by one via
  // the project page means a round trip per inlay.
  const siblings = createMemo(() => siblingsQuery.data ?? []);
  const currentIndex = createMemo(() =>
    siblings().findIndex((s) => s.uuid === params().inlayId),
  );
  const previousInlay = createMemo(() =>
    currentIndex() > 0 ? siblings()[currentIndex() - 1] : null,
  );
  const nextInlay = createMemo(() => {
    const index = currentIndex();
    return index >= 0 && index < siblings().length - 1
      ? siblings()[index + 1]
      : null;
  });

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

  // Any catalog inlay can be (re)coloured right up until the order is placed,
  // whatever phase it is in — doing so resets it to awaiting internal pricing.
  const canCustomize = createMemo(
    () => inlay()?.type === "catalog" && project()?.status === "draft",
  );

  function invalidateInlay() {
    queryClient.invalidateQueries({ queryKey: ["inlay", params().inlayId] });
    queryClient.invalidateQueries({ queryKey: ["project", params().id] });
    queryClient.invalidateQueries({
      queryKey: ["project", params().id, "inlays"],
    });
  }

  return (
    <div class="grid gap-6 lg:h-[var(--app-pane-height)] lg:grid-cols-[minmax(0,1fr)_380px] lg:overflow-hidden">
      <div class="min-w-0 space-y-6 lg:overflow-y-auto lg:pr-2">
        <div class="flex flex-wrap items-center justify-between gap-3">
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

          <div class="flex items-center gap-2">
            <Show when={siblings().length > 1}>
              {/* Rendered as a plain disabled button at the ends — an anchor
                  ignores `disabled` and would still navigate. */}
              <Show
                when={previousInlay()}
                fallback={
                  <Button variant="outline" size="sm" disabled>
                    <IoChevronBack size={16} />
                  </Button>
                }
              >
                {(previous) => (
                  <Button
                    as={Link}
                    to="/projects/$id/inlay/$inlayId"
                    params={{ id: params().id, inlayId: previous().uuid }}
                    variant="outline"
                    size="sm"
                    aria-label="Previous inlay"
                  >
                    <IoChevronBack size={16} />
                  </Button>
                )}
              </Show>

              <span class="text-sm tabular-nums text-faint">
                {currentIndex() + 1} of {siblings().length}
              </span>

              <Show
                when={nextInlay()}
                fallback={
                  <Button variant="outline" size="sm" disabled>
                    <IoChevronForward size={16} />
                  </Button>
                }
              >
                {(next) => (
                  <Button
                    as={Link}
                    to="/projects/$id/inlay/$inlayId"
                    params={{ id: params().id, inlayId: next().uuid }}
                    variant="outline"
                    size="sm"
                    aria-label="Next inlay"
                  >
                    <IoChevronForward size={16} />
                  </Button>
                )}
              </Show>
            </Show>

            <Show when={canCustomize()}>
              <Can permission={PERMISSION_ACTIONS.MANAGE_PROJECT}>
                <Button
                  as={Link}
                  to="/projects/$id/inlay/$inlayId/recustomize"
                  params={{ id: params().id, inlayId: params().inlayId }}
                  size="sm"
                >
                  {inlay()?.is_customized ? "Adjust design" : "Customize design"}
                </Button>
              </Can>
            </Show>

            <Button
              as={Link}
              to="/projects/$id"
              params={{ id: params().id }}
              variant="outline"
              size="sm"
            >
              Back to project
            </Button>
          </div>
        </div>

        <Switch>
          <Match when={inlayQuery.isLoading || projectQuery.isLoading}>
            <div class="grid grid-cols-1 gap-6 lg:grid-cols-[300px_minmax(0,1fr)]">
              <div class="h-96 animate-pulse rounded-lg border border-border bg-card" />
              <div class="h-96 animate-pulse rounded-lg border border-border bg-card" />
            </div>
          </Match>

          <Match when={inlayQuery.isError || projectQuery.isError}>
            <div class="flex flex-col items-center gap-2 rounded-lg border border-dashed border-error-foreground/40 bg-error p-8 text-center">
              <p class="flex items-center gap-2 font-medium text-error-foreground">
                <IoAlertCircleOutline size={18} aria-hidden="true" />
                Couldn't load this inlay
              </p>
              <Button
                variant="secondary"
                class="mt-2"
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
                <Show when={project()!.status === "draft"}>
                  <InlayArrangeActions
                    projectUuid={params().id}
                    inlay={inlay()!}
                    duplicateHooks={{
                      after: (duplicate) =>
                        navigate({
                          to: "/projects/$id/inlay/$inlayId",
                          params: { id: params().id, inlayId: duplicate.uuid },
                        }),
                    }}
                  />
                </Show>
              </div>

              <div class="space-y-6">
                <Switch>
                  <Match when={phase() === "awaiting-approval"}>
                    <Show when={pendingProof()}>
                      {(proof) => (
                        <ProofReviewPanel
                          proof={proof()}
                          inlayUuid={params().inlayId}
                          mirror={inlay()!}
                        />
                      )}
                    </Show>
                  </Match>

                  <Match when={isInShop()}>
                    <ManufacturingPanel
                      inlay={inlay()!}
                      orderedAt={project()?.ordered_at ?? null}
                    />
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
                    <ReadyPanel />
                  </Match>

                  <Match when={phase() === "cancelled"}>
                    <div class="rounded-lg border border-border bg-card p-6 text-sm text-muted-foreground">
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
                    sandblastFileFormat={project()!.sandblast_file_format}
                    projectStatus={project()!.status}
                    onUploaded={invalidateInlay}
                    layout="card"
                  />
                </Show>

                <DesignHistory
                  inlayUuid={params().inlayId}
                  mirror={inlay()!}
                  excludeProofId={pendingProof()?.id}
                />
              </div>
            </div>
          </Match>

          <Match when={!inlay()}>
            <div class="rounded-lg border border-dashed border-input p-8 text-center">
              <p class="text-faint">Inlay not found</p>
            </div>
          </Match>
        </Switch>
      </div>

      {/* The project's one thread, in full. Messages sent from here are tagged
          with this inlay; the tag on each message is what tells them apart. */}
      <ProjectDiscussion
        projectUuid={params().id}
        tagInlayUuid={params().inlayId}
        class="min-h-[24rem] lg:h-full lg:min-h-0"
      />
    </div>
  );
}

function ManufacturingPanel(props: {
  inlay: InlayDetail;
  orderedAt: string | null;
}) {
  const step = createMemo(
    () => props.inlay.manufacturing_step as ManufacturingStep | null,
  );

  return (
    <div class="space-y-6">
      <Show when={step()}>
        {(currentStep) => (
          <div class="space-y-4 rounded-lg border border-border bg-card p-5">
            <div class="flex items-center justify-between gap-3">
              <div class="flex flex-col gap-1">
                <p class="eyebrow">Manufacturing</p>
                <h2 class="text-xl text-foreground">
                  {STEP_LABELS[currentStep()]}
                </h2>
              </div>
              <Badge variant="production" lozenge>
                In production
              </Badge>
            </div>
            <ManufacturingTracker
              currentStep={currentStep()}
              orderedAt={props.orderedAt}
            />
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
          <div class="space-y-4 rounded-lg border border-border bg-card p-5">
            <div class="flex items-start justify-between gap-2">
              <h2 class="text-xl text-foreground">Design request</h2>
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
                <p class="max-w-prose whitespace-pre-wrap leading-relaxed text-foreground">
                  {info().description}
                </p>
                <Show
                  when={info().requested_width && info().requested_height}
                >
                  <p class="text-sm tabular-nums text-muted-foreground">
                    Requested: {info().requested_width}" &times;{" "}
                    {info().requested_height}"
                  </p>
                </Show>

                <Show when={(info().reference_images ?? []).length > 0}>
                  <div class="space-y-1.5">
                    <p class="eyebrow">
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
                            triggerClass="came-frame block aspect-square w-full overflow-hidden rounded-md bg-came"
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

      <div class="flex flex-col items-center gap-3 rounded-lg border border-dashed border-input px-6 py-10 text-center">
        <div class="ornament-rule w-32" aria-hidden="true">
          <span class="lozenge" />
        </div>
        <p class="font-display text-xl text-foreground">
          Waiting on a proof from GlassAct
        </p>
        <p class="text-sm text-muted-foreground">
          A designer will send a proof for approval.
        </p>
        <Can permission={PERMISSION_ACTIONS.CREATE_PROOF}>
          <div class="mt-1 flex justify-center">
            <CreateProofDialog
              inlayUuid={props.inlayUuid}
              requestedWidth={props.inlay.custom_info?.requested_width}
              requestedHeight={props.inlay.custom_info?.requested_height}
              onProofCreated={props.onChanged}
            />
          </div>
        </Can>
      </div>
    </div>
  );
}

function ReadyPanel() {
  return (
    <div class="space-y-3 rounded-lg border border-success-foreground/30 bg-card p-5">
      <div class="flex items-center justify-between gap-3">
        <h2 class="text-xl text-foreground">Ready to order</h2>
        <Badge variant="success" lozenge>
          Ready
        </Badge>
      </div>
      <p class="text-sm text-muted-foreground">
        This inlay is approved and will be included when the order is placed
        from the project page.
      </p>
    </div>
  );
}
