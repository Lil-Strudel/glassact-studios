import { createFileRoute, Link } from "@tanstack/solid-router";
import {
  Breadcrumb,
  Button,
  Badge,
  type BadgeVariant,
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  Tooltip,
  TooltipContent,
  TooltipTrigger,
  showToast,
  cn,
} from "@glassact/ui";
import { createMemo, createSignal, For, Match, Show, Switch } from "solid-js";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import {
  getProjectOpts,
  deleteProjectOpts,
  patchProjectOpts,
  postMarkProjectShippedOpts,
  postMarkProjectDeliveredOpts,
} from "../../queries/project";
import {
  getInlaysByProjectOpts,
  deleteInlayOpts,
} from "../../queries/inlay";
import {
  getProjectInvoiceOpts,
  postProjectInvoiceOpts,
  postMarkInvoicePaidOpts,
  postVoidInvoiceOpts,
} from "../../queries/invoice";
import type {
  ManufacturingStep,
  ProjectStatus,
  InlayWithInfo,
  SandblastFileFormat,
} from "@glassact/data";
import { PERMISSION_ACTIONS, INSTALLATION_KIT_PRICE_CENTS } from "@glassact/data";
import { ProjectStatusBadge } from "../../components/project/status-badge";
import { WatchButton } from "../../components/project/watch-button";
import { inlayDeleteBlockedMessage } from "../../utils/inlay-delete";
import { Can } from "../../components/Can";
import { useUserContext } from "../../providers/user";
import { isApiError } from "../../utils/is-api-error";
import { formatMoney } from "../../utils/format-money";
import { formatPriceFormula } from "../../utils/format-price-formula";
import { PlaceOrderCart } from "../../components/place-order-cart";
import { StoneDesignSection } from "../../components/stone-designer/stone-design-section";
import {
  IoTrashOutline,
  IoAddCircleOutline,
  IoAlertCircleOutline,
  IoPencilOutline,
} from "solid-icons/io";
import { LozengeStepper } from "../../components/project/lozenge-stepper";
import { ManufacturingTracker } from "../../components/manufacturing/manufacturing-tracker";
import { SandblastFileCard } from "../../components/inlay/sandblast-file-card";
import { MirrorBadge } from "../../components/inlay/mirror-badge";
import { InlayArrangeActions } from "../../components/inlay/inlay-arrange-actions";
import { mirrorCssTransform } from "../../utils/inlay-mirror";
import { ProjectDiscussion } from "../../components/chat/project-discussion";
import { PriceCaveat } from "../../components/price-caveat";
import {
  isManufacturingStatus,
  PROJECT_STATUS_LABELS,
} from "../../utils/project-status";
import { paymentTimingNotice } from "../../utils/payment-timing";

export const Route = createFileRoute("/_app/projects_/$id/")({
  component: RouteComponent,
});

const STATUS_STEPS: ProjectStatus[] = [
  "draft",
  "ordered",
  "in-production",
  "shipped",
  "invoiced",
  "completed",
];

const STATUS_STEP_LABELS = STATUS_STEPS.map(
  (status) => PROJECT_STATUS_LABELS[status],
);

const EDITABLE_STATUSES: ProjectStatus[] = ["draft"];
const CANCELLABLE_STATUSES: ProjectStatus[] = ["draft", "ordered"];
// Statuses from which an internal user may record a shipment.
const SHIPPABLE_STATUSES: ProjectStatus[] = ["ordered", "in-production"];

function RouteComponent() {
  const params = Route.useParams();
  const queryClient = useQueryClient();
  const userContext = useUserContext();

  const projectQuery = useQuery(() => getProjectOpts(params().id));
  const inlaysQuery = useQuery(() => getInlaysByProjectOpts(params().id));
  const cancelProject = useMutation(deleteProjectOpts);
  const removeInlay = useMutation(deleteInlayOpts);
  const toggleKit = useMutation(patchProjectOpts);
  const markShipped = useMutation(postMarkProjectShippedOpts);
  const markDelivered = useMutation(postMarkProjectDeliveredOpts);

  const [trackingNumber, setTrackingNumber] = createSignal("");

  // Invoices can be attached at any point once an order exists, so surface the
  // section for every non-draft, non-cancelled project. Attaching is gated by
  // the CREATE_INVOICE permission inside InvoiceSection itself.
  const showInvoiceSection = createMemo(() => {
    if (!projectQuery.isSuccess) return false;
    const status = projectQuery.data.status;
    return status !== "draft" && status !== "cancelled";
  });
  const invoiceQuery = useQuery(() => ({
    ...getProjectInvoiceOpts(params().id),
    retry: false,
    enabled: showInvoiceSection(),
  }));

  const inlays = () => (inlaysQuery.isSuccess ? inlaysQuery.data : []);

  const canCancel = createMemo(() => {
    if (!projectQuery.isSuccess) return false;
    return CANCELLABLE_STATUSES.includes(projectQuery.data.status);
  });

  const canEditInlays = createMemo(() => {
    if (!projectQuery.isSuccess) return false;
    return EDITABLE_STATUSES.includes(projectQuery.data.status);
  });

  const canPlaceOrder = createMemo(() => {
    if (!projectQuery.isSuccess) return false;
    if (projectQuery.data.status !== "draft") return false;
    const list = inlays();
    if (list.length === 0) return false;
    return list.every((inlay) => inlay.is_ready);
  });

  // One kit covers the whole project, so it is a flat add-on rather than a
  // per-inlay multiple.
  const hasInstallationKit = createMemo(
    () => projectQuery.data?.installation_kit ?? false,
  );

  const notReadyCount = createMemo(
    () => inlays().filter((inlay) => !inlay.is_ready).length,
  );

  const placeOrderBlockedReason = createMemo(() => {
    if (inlays().length === 0) return "Add an inlay to place an order.";
    const count = notReadyCount();
    if (count === 0) return null;
    return `${count} ${count === 1 ? "inlay is" : "inlays are"} not ready yet.`;
  });

  const totalPriceCents = createMemo(
    () =>
      inlays().reduce((sum, inlay) => sum + (inlay.price_cents ?? 0), 0) +
      (hasInstallationKit() ? INSTALLATION_KIT_PRICE_CENTS : 0),
  );

  function handleToggleKit(next: boolean) {
    toggleKit.mutate(
      { uuid: params().id, body: { installation_kit: next } },
      {
        onSuccess() {
          queryClient.invalidateQueries({
            queryKey: ["project", params().id],
          });
        },
        onError(error) {
          if (isApiError(error)) {
            showToast({
              title: "Failed to update installation kit",
              description: error?.data?.error ?? "Unknown error",
              variant: "error",
            });
          }
        },
      },
    );
  }

  const currentStepIndex = createMemo(() => {
    if (!projectQuery.isSuccess) return -1;
    if (projectQuery.data.status === "cancelled") return -1;
    return STATUS_STEPS.indexOf(projectQuery.data.status);
  });

  function handleCancel() {
    if (!projectQuery.isSuccess) return;
    cancelProject.mutate(projectQuery.data.uuid, {
      onSuccess() {
        showToast({
          title: "Project cancelled",
          description: `${projectQuery.data!.name} has been cancelled.`,
          variant: "success",
        });
        queryClient.invalidateQueries({ queryKey: ["project"] });
      },
      onError(error) {
        if (isApiError(error)) {
          showToast({
            title: "Failed to cancel project",
            description: error?.data?.error ?? "Unknown error",
            variant: "error",
          });
        }
      },
    });
  }

  const canShip = createMemo(() => {
    if (!projectQuery.isSuccess) return false;
    return SHIPPABLE_STATUSES.includes(projectQuery.data.status);
  });

  const canDeliver = createMemo(() => {
    if (!projectQuery.isSuccess) return false;
    return projectQuery.data.status === "shipped";
  });

  const paymentNotice = createMemo(() => {
    if (!projectQuery.isSuccess) return null;
    return paymentTimingNotice({
      timing: projectQuery.data.payment_timing,
      awaitingPayment: projectQuery.data.awaiting_payment ?? false,
      isInternal: userContext.isInternal(),
    });
  });

  function handleMarkShipped() {
    if (!projectQuery.isSuccess) return;
    const tracking = trackingNumber().trim();
    if (tracking === "") return;
    markShipped.mutate(
      { uuid: projectQuery.data.uuid, trackingNumber: tracking },
      {
        onSuccess() {
          setTrackingNumber("");
          showToast({
            title: "Project marked shipped",
            description: `${projectQuery.data!.name} is on its way.`,
            variant: "success",
          });
          queryClient.invalidateQueries({ queryKey: ["project"] });
        },
        onError(error) {
          if (isApiError(error)) {
            showToast({
              title: "Failed to mark shipped",
              description: error?.data?.error ?? "Unknown error",
              variant: "error",
            });
          }
        },
      },
    );
  }

  function handleMarkDelivered() {
    if (!projectQuery.isSuccess) return;
    markDelivered.mutate(projectQuery.data.uuid, {
      onSuccess() {
        showToast({
          title: "Project marked delivered",
          description: `${projectQuery.data!.name} has been delivered.`,
          variant: "success",
        });
        queryClient.invalidateQueries({ queryKey: ["project"] });
      },
      onError(error) {
        if (isApiError(error)) {
          showToast({
            title: "Failed to mark delivered",
            description: error?.data?.error ?? "Unknown error",
            variant: "error",
          });
        }
      },
    });
  }

  function handleDeleteInlay(inlay: InlayWithInfo) {
    removeInlay.mutate(inlay.uuid, {
      onSuccess() {
        showToast({
          title: "Inlay removed",
          description: `${inlay.name} has been removed from the project.`,
          variant: "success",
        });
        queryClient.invalidateQueries({
          queryKey: ["project", params().id, "inlays"],
        });
      },
      onError(error) {
        if (isApiError(error)) {
          showToast({
            title: "Failed to remove inlay",
            description: error?.data?.error ?? "Unknown error",
            variant: "error",
          });
        }
      },
    });
  }

  return (
    <Switch>
      <Match when={projectQuery.isLoading}>
        <div class="flex flex-col gap-6">
          <div class="h-4 w-48 animate-pulse rounded bg-muted" />
          <div class="h-10 w-72 animate-pulse rounded bg-muted" />
          <div class="h-12 animate-pulse rounded bg-muted" />
          <div class="h-28 animate-pulse rounded-lg border border-border bg-card" />
          <For each={[1, 2, 3]}>
            {() => (
              <div class="h-28 animate-pulse rounded-lg border border-border bg-card" />
            )}
          </For>
        </div>
      </Match>

      <Match when={projectQuery.isError}>
        <div class="flex flex-col items-center gap-2 rounded-lg border border-dashed border-error-foreground/40 bg-error p-8 text-center">
          <p class="flex items-center gap-2 font-medium text-error-foreground">
            <IoAlertCircleOutline size={18} aria-hidden="true" />
            Couldn't load this project
          </p>
          <p class="text-sm text-muted-foreground">
            {projectQuery.error?.message ?? "An unexpected error occurred."}
          </p>
          <Button
            variant="secondary"
            class="mt-2"
            onClick={() => projectQuery.refetch()}
          >
            Retry
          </Button>
        </div>
      </Match>

      <Match when={projectQuery.isSuccess}>
        <div class="grid gap-6 lg:h-[var(--app-pane-height)] lg:grid-cols-[minmax(0,1fr)_380px] lg:overflow-hidden">
          <div class="flex min-w-0 flex-col gap-6 lg:overflow-y-auto lg:pr-2 [&>*]:shrink-0">
          <Breadcrumb
            crumbs={[
              { title: "Projects", to: "/projects" },
              {
                title: projectQuery.data!.name,
                to: `/projects/${projectQuery.data!.uuid}`,
              },
            ]}
          />

          <header class="flex flex-col gap-4 border-b border-border pb-5 sm:flex-row sm:items-end sm:justify-between">
            <div class="flex min-w-0 flex-col gap-2">
              <Show when={projectQuery.data!.dealership_name}>
                {(name) => <p class="eyebrow">{name()}</p>}
              </Show>
              <h1 class="text-3xl sm:text-4xl">{projectQuery.data!.name}</h1>
              <div class="flex flex-wrap items-center gap-x-4 gap-y-2">
                <ProjectStatusBadge status={projectQuery.data!.status} />
                <InternalReferenceField
                  projectUuid={projectQuery.data!.uuid}
                  value={projectQuery.data!.internal_reference}
                />
              </div>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <WatchButton
                projectUuid={params().id}
                isWatching={projectQuery.data!.is_watching}
                watcherCount={projectQuery.data!.watcher_count}
              />
              <Show when={canCancel()}>
                <Dialog>
                  <DialogTrigger as={Button} variant="destructive">
                    Cancel project
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>Cancel project</DialogTitle>
                    </DialogHeader>
                    <p class="text-sm text-muted-foreground">
                      Are you sure you want to cancel{" "}
                      <span class="font-semibold">
                        {projectQuery.data!.name}
                      </span>
                      ? This action cannot be undone.
                    </p>
                    <div class="flex justify-end gap-3 mt-4">
                      <DialogClose
                        as={Button}
                        variant="outline"
                        disabled={cancelProject.isPending}
                      >
                        Close
                      </DialogClose>
                      <Button
                        variant="destructive"
                        onClick={handleCancel}
                        disabled={cancelProject.isPending}
                      >
                        {cancelProject.isPending
                          ? "Cancelling..."
                          : "Cancel project"}
                      </Button>
                    </div>
                  </DialogContent>
                </Dialog>
              </Show>
              <Can permission={PERMISSION_ACTIONS.MANAGE_SHIPPING}>
                <Show when={canShip()}>
                  <Dialog>
                    <DialogTrigger as={Button}>Mark shipped</DialogTrigger>
                    <DialogContent>
                      <DialogHeader>
                        <DialogTitle>Mark project shipped</DialogTitle>
                      </DialogHeader>
                      <p class="text-sm text-muted-foreground">
                        Enter the tracking number for this shipment. It will be
                        shared with the dealership so they can look it up.
                      </p>
                      <input
                        type="text"
                        placeholder="Tracking number"
                        value={trackingNumber()}
                        onInput={(e) =>
                          setTrackingNumber(e.currentTarget.value)
                        }
                        class="mt-3 h-9 w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-faint focus-visible:border-gilt focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-gilt/25"
                      />
                      <div class="flex justify-end gap-3 mt-4">
                        <DialogClose as={Button} variant="outline">
                          Cancel
                        </DialogClose>
                        <Button
                          onClick={handleMarkShipped}
                          disabled={
                            markShipped.isPending ||
                            trackingNumber().trim() === ""
                          }
                        >
                          {markShipped.isPending
                            ? "Marking..."
                            : "Mark shipped"}
                        </Button>
                      </div>
                    </DialogContent>
                  </Dialog>
                </Show>
                <Show when={canDeliver()}>
                  <Button
                    onClick={handleMarkDelivered}
                    disabled={markDelivered.isPending}
                  >
                    {markDelivered.isPending ? "Marking..." : "Mark delivered"}
                  </Button>
                </Show>
              </Can>
            </div>
          </header>

          <Show when={projectQuery.data!.status !== "cancelled"}>
            <LozengeStepper
              labels={STATUS_STEP_LABELS}
              currentIndex={currentStepIndex()}
            />
          </Show>

          <Show when={paymentNotice()}>
            {(notice) => (
              <div
                class={`rounded-lg border px-4 py-3 text-sm ${
                  notice().tone === "warning"
                    ? "border-warning-foreground/30 bg-warning text-warning-foreground"
                    : "border-info-foreground/30 bg-info text-info-foreground"
                }`}
              >
                {notice().message}
              </div>
            )}
          </Show>

          <Show when={projectQuery.data!.tracking_number}>
            {(tracking) => (
              <div class="flex flex-wrap items-baseline gap-x-3 gap-y-1 rounded-lg border border-border bg-card px-4 py-3 text-sm">
                <span class="eyebrow">Tracking number</span>
                <span class="font-mono tabular-nums text-foreground">
                  {tracking()}
                </span>
              </div>
            )}
          </Show>

          <section
            aria-label="Order summary"
            class="relative overflow-hidden rounded-lg border border-input bg-card"
          >
            <div
              aria-hidden="true"
              class="pointer-events-none absolute inset-0 bg-gradient-to-r from-crimson-deep/40 via-crimson-deep/10 to-transparent"
            />
            <div class="relative grid gap-5 p-5 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center">
              <div class="flex flex-col gap-1">
                <p class="eyebrow">
                  {projectQuery.data!.status === "draft"
                    ? "Order total so far"
                    : "Project total"}
                </p>
                <p class="font-display text-3xl tabular-nums text-foreground">
                  {formatMoney(totalPriceCents() / 100)}
                </p>
                <PriceCaveat />
                <div class="pt-2">
                  <Show
                    when={projectQuery.data!.status === "draft"}
                    fallback={
                      <p class="text-sm text-muted-foreground">
                        Installation kit:{" "}
                        <span class="font-medium text-foreground">
                          {hasInstallationKit() ? "Included" : "Not included"}
                        </span>
                      </p>
                    }
                  >
                    {/* Matches the PATCH /project route's gate, which internal
                        staff pass but create_project would exclude them from. */}
                    <Can permission={PERMISSION_ACTIONS.MANAGE_PROJECT}>
                      <label class="flex w-fit cursor-pointer items-start gap-2.5 text-sm text-foreground">
                        <input
                          type="checkbox"
                          class="mt-0.5 h-4 w-4 accent-primary"
                          checked={hasInstallationKit()}
                          disabled={toggleKit.isPending}
                          onChange={(e) => handleToggleKit(e.currentTarget.checked)}
                        />
                        <span class="flex flex-col">
                          <span>
                            Include installation kit{" "}
                            <span class="tabular-nums text-muted-foreground">
                              ({formatMoney(INSTALLATION_KIT_PRICE_CENTS / 100)})
                            </span>
                          </span>
                          <span class="text-xs text-faint">
                            One kit covers every inlay in this project.
                          </span>
                        </span>
                      </label>
                    </Can>
                  </Show>
                </div>
              </div>
              <Can permission={PERMISSION_ACTIONS.PLACE_ORDER}>
                <Show when={projectQuery.data!.status === "draft"}>
                  <div class="flex flex-col items-start gap-1.5 sm:items-end">
                    <PlaceOrderCart
                      project={projectQuery.data!}
                      inlays={inlays()}
                      disabled={!canPlaceOrder()}
                    />
                    <Show when={placeOrderBlockedReason()}>
                      {(reason) => (
                        <p class="text-xs text-faint">{reason()}</p>
                      )}
                    </Show>
                  </div>
                </Show>
              </Can>
            </div>
          </section>

          <StoneDesignSection
            projectUuid={params().id}
            inlays={inlays()}
            isDraft={projectQuery.data!.status === "draft"}
          />

          <section class="flex flex-col gap-4">
            <div class="flex items-center justify-between gap-4">
              <h2 class="text-2xl">Inlays</h2>
              <Show when={canEditInlays()}>
                <Can permission={PERMISSION_ACTIONS.CREATE_PROJECT}>
                  <Button
                    as={Link}
                    to={`/projects/${params().id}/add-inlay`}
                    variant="outline"
                    size="sm"
                  >
                    <IoAddCircleOutline size={18} class="mr-1.5" />
                    Add inlay
                  </Button>
                </Can>
              </Show>
            </div>

            <Switch>
              <Match when={inlaysQuery.isLoading}>
                <div class="flex flex-col gap-3">
                  <For each={[1, 2, 3]}>
                    {() => (
                      <div class="h-28 animate-pulse rounded-lg border border-border bg-card" />
                    )}
                  </For>
                </div>
              </Match>

              <Match when={inlaysQuery.isError}>
                <div class="flex flex-col items-center gap-2 rounded-lg border border-dashed border-error-foreground/40 bg-error p-8 text-center">
                  <p class="flex items-center gap-2 font-medium text-error-foreground">
                    <IoAlertCircleOutline size={18} aria-hidden="true" />
                    Couldn't load inlays
                  </p>
                  <Button
                    variant="secondary"
                    class="mt-2"
                    onClick={() => inlaysQuery.refetch()}
                  >
                    Retry
                  </Button>
                </div>
              </Match>

              <Match
                when={inlaysQuery.isSuccess && inlaysQuery.data!.length === 0}
              >
                <div class="flex flex-col items-center gap-3 rounded-lg border border-dashed border-input px-6 py-12 text-center">
                  <div class="ornament-rule w-40" aria-hidden="true">
                    <span class="lozenge" />
                  </div>
                  <p class="font-display text-xl text-foreground">
                    Inlays on this project will appear here
                  </p>
                  <p class="max-w-prose text-sm text-muted-foreground">
                    Add a catalog inlay or request a custom design to get
                    started.
                  </p>
                  <Show when={canEditInlays()}>
                    <Can permission={PERMISSION_ACTIONS.CREATE_PROJECT}>
                      <Button
                        as={Link}
                        to={`/projects/${params().id}/add-inlay`}
                        class="mt-1"
                      >
                        Add inlay
                      </Button>
                    </Can>
                  </Show>
                </div>
              </Match>

              <Match
                when={inlaysQuery.isSuccess && inlaysQuery.data!.length > 0}
              >
                <div class="flex flex-col gap-3">
                  <For each={inlaysQuery.data!}>
                    {(inlay) => (
                      <InlayCard
                        inlay={inlay}
                        projectId={params().id}
                        projectStatus={projectQuery.data!.status}
                        sandblastFileFormat={
                          projectQuery.data!.sandblast_file_format
                        }
                        canDelete={
                          projectQuery.data!.status === "draft"
                        }
                        canArrange={
                          projectQuery.data!.status === "draft"
                        }
                        onDelete={() => handleDeleteInlay(inlay)}
                        isDeleting={removeInlay.isPending}
                      />
                    )}
                  </For>
                </div>
              </Match>
            </Switch>
          </section>

          <Show when={showInvoiceSection()}>
            <section class="flex flex-col gap-4">
              <h2 class="text-2xl">Invoice</h2>
              <InvoiceSection
                projectUuid={params().id}
                invoice={invoiceQuery.data ?? null}
                isLoading={invoiceQuery.isLoading}
                onInvoiceChange={() => {
                  queryClient.invalidateQueries({
                    queryKey: ["project", params().id, "invoice"],
                  });
                  queryClient.invalidateQueries({ queryKey: ["project"] });
                }}
              />
            </section>
          </Show>
          </div>

          {/* Pinned pane: the conversation fills the window and never moves
              while you scroll the inlay list. On narrow screens it stacks
              underneath instead. */}
          <ProjectDiscussion
            projectUuid={params().id}
            class="min-h-[24rem] lg:h-full lg:min-h-0"
          />
        </div>
      </Match>
    </Switch>
  );
}

interface InternalReferenceFieldProps {
  projectUuid: string;
  value: string | null;
}

function InternalReferenceField(props: InternalReferenceFieldProps) {
  const { can } = useUserContext();
  const queryClient = useQueryClient();
  const patchMutation = useMutation(patchProjectOpts);

  const [isEditing, setIsEditing] = createSignal(false);
  const [draft, setDraft] = createSignal("");

  const canEdit = () => can(PERMISSION_ACTIONS.MANAGE_PROJECT);

  function startEdit() {
    setDraft(props.value ?? "");
    setIsEditing(true);
  }

  function cancelEdit() {
    setDraft(props.value ?? "");
    setIsEditing(false);
  }

  function save() {
    const trimmed = draft().trim();
    patchMutation.mutate(
      {
        uuid: props.projectUuid,
        body: { internal_reference: trimmed === "" ? null : trimmed },
      },
      {
        onSuccess() {
          showToast({
            title: "Reference updated",
            variant: "success",
          });
          queryClient.invalidateQueries({
            queryKey: ["project", props.projectUuid],
          });
          queryClient.invalidateQueries({ queryKey: ["project"] });
          setIsEditing(false);
        },
        onError(error) {
          if (isApiError(error)) {
            showToast({
              title: "Failed to update reference",
              description: error?.data?.error ?? "Unknown error",
              variant: "error",
            });
          }
        },
      },
    );
  }

  return (
    <div class="flex items-center gap-2 text-sm text-muted-foreground">
      <span class="eyebrow text-faint">PO / Ref</span>
      <Show
        when={isEditing()}
        fallback={
          <>
            <span class="tabular-nums text-foreground">
              {props.value || "—"}
            </span>
            <Show when={canEdit()}>
              <Button
                size="sm"
                variant="ghost"
                onClick={startEdit}
                class="h-7 px-2"
                aria-label="Edit PO / reference"
              >
                <IoPencilOutline size={14} />
              </Button>
            </Show>
          </>
        }
      >
        <input
          type="text"
          value={draft()}
          onInput={(e) => setDraft(e.currentTarget.value)}
          placeholder="e.g. PO-2025-0142"
          class="h-8 rounded-md border border-input bg-background px-2 py-1 text-sm text-foreground placeholder:text-faint focus-visible:border-gilt focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-gilt/25"
          disabled={patchMutation.isPending}
        />
        <Button
          size="sm"
          onClick={save}
          disabled={patchMutation.isPending}
        >
          {patchMutation.isPending ? "Saving..." : "Save"}
        </Button>
        <Button
          size="sm"
          variant="outline"
          onClick={cancelEdit}
          disabled={patchMutation.isPending}
        >
          Cancel
        </Button>
      </Show>
    </div>
  );
}

interface InlayCardProps {
  inlay: InlayWithInfo;
  projectId: string;
  projectStatus: ProjectStatus;
  sandblastFileFormat?: SandblastFileFormat;
  canDelete: boolean;
  canArrange: boolean;
  onDelete: () => void;
  isDeleting: boolean;
}

type ReadinessBadge = { label: string; variant: BadgeVariant };

function readinessBadge(inlay: InlayWithInfo): ReadinessBadge {
  if (inlay.is_ready) {
    return { label: "Ready", variant: "success" };
  }

  const hasPending = inlay.has_pending_proof === true;

  if (inlay.is_customized && inlay.type === "catalog" && hasPending) {
    return { label: "Needs internal review", variant: "info" };
  }

  if (inlay.type === "custom" && hasPending) {
    return { label: "Needs customer approval", variant: "gilt" };
  }

  return { label: "Awaiting proof", variant: "outline" };
}

function InlayCard(props: InlayCardProps) {
  const showManufacturingTracker = () =>
    isManufacturingStatus(props.projectStatus) &&
    props.inlay.manufacturing_step != null;

  const description = () => {
    if (props.inlay.type === "catalog" && props.inlay.catalog_info) {
      return props.inlay.catalog_info.customization_notes || null;
    }
    if (props.inlay.type === "custom" && props.inlay.custom_info) {
      return props.inlay.custom_info.description || null;
    }
    return null;
  };

  const badge = createMemo(() => readinessBadge(props.inlay));
  const priceLabel = createMemo(() => {
    if (props.inlay.price_cents == null && !props.inlay.price_group_name) {
      return "—";
    }
    const dollars = (props.inlay.price_cents ?? 0) / 100;
    return formatMoney(dollars);
  });
  const priceFormula = createMemo(() =>
    formatPriceFormula(
      props.inlay.price_group_name,
      props.inlay.price_adjustment_type,
      props.inlay.price_adjustment_value,
    ),
  );

  const showInternalApprove = createMemo(
    () =>
      props.inlay.is_customized &&
      props.inlay.type === "catalog" &&
      !props.inlay.approved_proof_id &&
      props.inlay.has_pending_proof === true,
  );

  const userContext = useUserContext();
  const queryClient = useQueryClient();

  const needsYourAttention = createMemo(() => {
    if (props.inlay.is_ready || props.inlay.has_pending_proof !== true) {
      return false;
    }
    return userContext.isDealership()
      ? props.inlay.type === "custom"
      : showInternalApprove();
  });

  const typeLabel = () =>
    [
      props.inlay.type === "catalog" ? "Catalog" : "Custom design",
      props.inlay.is_customized ? "Customized" : null,
    ]
      .filter(Boolean)
      .join(" · ");

  const hasSandblast = () => !!props.inlay.sandblast_file_url;
  const canUploadSandblast = () =>
    props.projectStatus !== "draft" &&
    userContext.can(PERMISSION_ACTIONS.MANAGE_KANBAN);

  return (
    <article
      class={cn(
        "overflow-hidden rounded-lg border bg-card",
        needsYourAttention() ? "border-gilt/45" : "border-border",
      )}
    >
      <Link
        to="/projects/$id/inlay/$inlayId"
        params={{ id: props.projectId, inlayId: props.inlay.uuid }}
        class="group grid grid-cols-[72px_minmax(0,1fr)] items-start gap-x-4 gap-y-3 p-3 transition-colors hover:bg-accent/40 sm:grid-cols-[88px_minmax(0,1fr)_auto] sm:items-center"
      >
        <Show
          when={props.inlay.preview_url}
          fallback={
            <div class="came-frame flex aspect-square items-center justify-center rounded-lg bg-came">
              <span class="text-[10px] text-faint">No preview</span>
            </div>
          }
        >
          <div class="came-frame flex aspect-square items-center justify-center overflow-hidden rounded-lg bg-plaque p-1.5">
            <img
              src={props.inlay.preview_url}
              alt={props.inlay.name}
              class="max-h-full max-w-full object-contain"
              style={{ transform: mirrorCssTransform(props.inlay) }}
            />
          </div>
        </Show>
        <div class="flex min-w-0 flex-col gap-1">
          <h3 class="truncate text-lg text-foreground group-hover:text-crimson-light">
            {props.inlay.name}
          </h3>
          <p class="text-xs text-faint">{typeLabel()}</p>
          <Show when={description()}>
            {(desc) => (
              <p class="line-clamp-2 text-sm text-muted-foreground">{desc()}</p>
            )}
          </Show>
          <div class="flex flex-wrap items-center gap-1.5 pt-1">
            <Badge variant={badge().variant} lozenge>
              {badge().label}
            </Badge>
            <MirrorBadge inlay={props.inlay} />
          </div>
        </div>
        <div class="col-span-2 flex items-baseline justify-between gap-3 sm:col-span-1 sm:flex-col sm:items-end sm:gap-1 sm:text-right">
          <span class="font-medium tabular-nums text-foreground">
            {priceLabel()}
          </span>
          <Show when={priceFormula()}>
            {(formula) => (
              <span class="text-xs tabular-nums text-faint">{formula()}</span>
            )}
          </Show>
        </div>
      </Link>
      <Show when={showManufacturingTracker()}>
        <div class="border-t border-border px-4 py-3">
          <ManufacturingTracker
            currentStep={props.inlay.manufacturing_step as ManufacturingStep}
          />
        </div>
      </Show>
      <Show
        when={
          props.canDelete ||
          props.canArrange ||
          showInternalApprove() ||
          hasSandblast() ||
          canUploadSandblast()
        }
      >
        <div class="flex flex-col gap-2 border-t border-border bg-muted/60 px-3 py-2.5 sm:flex-row sm:flex-wrap sm:items-start">
          <SandblastFileCard
            inlayUuid={props.inlay.uuid}
            inlayName={props.inlay.name}
            sandblastFileUrl={props.inlay.sandblast_file_url}
            sandblastFileFormat={props.sandblastFileFormat}
            projectStatus={props.projectStatus}
            onUploaded={() =>
              queryClient.invalidateQueries({
                queryKey: ["project", props.projectId, "inlays"],
              })
            }
          />
          <Show when={showInternalApprove()}>
            <Can permission={PERMISSION_ACTIONS.INTERNAL_APPROVE_PROOF}>
              <Button
                as={Link}
                to="/projects/$id/inlay/$inlayId"
                params={{
                  id: props.projectId,
                  inlayId: props.inlay.uuid,
                }}
                size="sm"
              >
                Review &amp; approve
              </Button>
            </Can>
          </Show>
          <Show when={props.canArrange}>
            <InlayArrangeActions
              projectUuid={props.projectId}
              inlay={props.inlay}
            />
          </Show>
          <Show when={props.canDelete}>
            <div class="sm:ml-auto">
              <Show
                when={props.inlay.can_delete}
                fallback={<BlockedRemoveButton inlay={props.inlay} />}
              >
                <Dialog>
                  <DialogTrigger
                    as={Button}
                    variant="destructive"
                    size="sm"
                  >
                    <IoTrashOutline size={16} class="mr-1" />
                    Remove
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>Remove inlay</DialogTitle>
                    </DialogHeader>
                    <p class="text-sm text-muted-foreground">
                      Are you sure you want to remove{" "}
                      <span class="font-semibold">{props.inlay.name}</span> from
                      this project?
                    </p>
                    <div class="flex justify-end gap-3 mt-4">
                      <DialogClose
                        as={Button}
                        variant="outline"
                        disabled={props.isDeleting}
                      >
                        Close
                      </DialogClose>
                      <Button
                        variant="destructive"
                        onClick={props.onDelete}
                        disabled={props.isDeleting}
                      >
                        {props.isDeleting ? "Removing..." : "Remove inlay"}
                      </Button>
                    </div>
                  </DialogContent>
                </Dialog>
              </Show>
            </div>
          </Show>
        </div>
      </Show>
    </article>
  );
}

// A disabled Remove control for inlays the API reports as undeletable. The
// button is wrapped in a span because a disabled button emits no pointer
// events, which would otherwise swallow the tooltip that explains why.
function BlockedRemoveButton(props: { inlay: InlayWithInfo }) {
  const message = createMemo(() =>
    inlayDeleteBlockedMessage(props.inlay.delete_blockers ?? []),
  );

  return (
    <Tooltip>
      <TooltipTrigger as="span">
        <Button
          variant="destructive"
          size="sm"
          disabled
          aria-label={`Cannot remove ${props.inlay.name}`}
        >
          <IoTrashOutline size={16} class="mr-1" />
          Remove
        </Button>
      </TooltipTrigger>
      <TooltipContent class="max-w-xs text-left">{message()}</TooltipContent>
    </Tooltip>
  );
}

interface InvoiceSectionProps {
  projectUuid: string;
  invoice:
    | import("@glassact/data").GET<import("@glassact/data").Invoice>
    | null;
  isLoading: boolean;
  onInvoiceChange: () => void;
}

function InvoiceSection(props: InvoiceSectionProps) {
  const { can } = useUserContext();
  const queryClient = useQueryClient();
  const attachInvoice = useMutation(() => postProjectInvoiceOpts());
  const markPaid = useMutation(() => postMarkInvoicePaidOpts());
  const voidInvoice = useMutation(() => postVoidInvoiceOpts());

  const [invoiceUrl, setInvoiceUrl] = createSignal("");
  const [isVoidDialogOpen, setIsVoidDialogOpen] = createSignal(false);

  const handleAttach = () => {
    attachInvoice.mutate(
      { projectUuid: props.projectUuid, invoiceUrl: invoiceUrl() },
      {
        onSuccess() {
          setInvoiceUrl("");
          showToast({ title: "Invoice attached", variant: "success" });
          props.onInvoiceChange();
        },
        onError(error) {
          if (isApiError(error)) {
            showToast({
              title: "Failed to attach invoice",
              description: error?.data?.error ?? "Unknown error",
              variant: "error",
            });
          }
        },
      },
    );
  };

  const handleMarkPaid = () => {
    if (!props.invoice) return;
    markPaid.mutate(props.invoice.uuid, {
      onSuccess() {
        showToast({ title: "Invoice marked as paid", variant: "success" });
        props.onInvoiceChange();
      },
      onError(error) {
        if (isApiError(error)) {
          showToast({
            title: "Failed to mark invoice paid",
            description: error?.data?.error ?? "Unknown error",
            variant: "error",
          });
        }
      },
    });
  };

  const handleVoid = () => {
    if (!props.invoice) return;
    voidInvoice.mutate(props.invoice.uuid, {
      onSuccess(data) {
        showToast({ title: "Invoice voided", variant: "success" });
        queryClient.setQueryData(
          ["project", props.projectUuid, "invoice"],
          data,
        );
        props.onInvoiceChange();
        setIsVoidDialogOpen(false);
      },
      onError(error) {
        if (isApiError(error)) {
          showToast({
            title: "Failed to void invoice",
            description: error?.data?.error ?? "Unknown error",
            variant: "error",
          });
        }
      },
    });
  };

  const hasActiveInvoice = () =>
    props.invoice !== null && props.invoice.status !== "void";

  return (
    <div class="space-y-4 rounded-lg border border-border bg-card p-5">
      <Switch>
        <Match when={props.isLoading}>
          <div class="h-16 animate-pulse rounded bg-muted" />
        </Match>

        <Match when={!hasActiveInvoice() && can(PERMISSION_ACTIONS.CREATE_INVOICE)}>
          <div class="space-y-3">
            <p class="text-sm text-muted-foreground">
              Paste the invoice link from your billing platform to attach it to
              this project.
            </p>
            <div class="flex gap-2">
              <input
                type="url"
                placeholder="https://..."
                value={invoiceUrl()}
                onInput={(e) => setInvoiceUrl(e.currentTarget.value)}
                class="h-9 flex-1 rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-faint focus-visible:border-gilt focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-gilt/25"
              />
              <Button
                onClick={handleAttach}
                disabled={attachInvoice.isPending || invoiceUrl().trim() === ""}
              >
                {attachInvoice.isPending ? "Attaching..." : "Attach invoice"}
              </Button>
            </div>
          </div>
        </Match>

        <Match when={!hasActiveInvoice() && !can(PERMISSION_ACTIONS.CREATE_INVOICE)}>
          <div class="text-center py-4">
            <p class="text-sm text-faint">
              Invoice not yet available. Check back soon.
            </p>
          </div>
        </Match>

        <Match when={hasActiveInvoice()}>
          <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div class="space-y-1">
              <div class="flex items-center gap-2">
                <span class="eyebrow">Invoice</span>
                <Badge
                  variant={
                    props.invoice!.status === "paid"
                      ? "success"
                      : props.invoice!.status === "sent"
                        ? "invoiced"
                        : "outline"
                  }
                  lozenge
                  class="capitalize"
                >
                  {props.invoice!.status}
                </Badge>
              </div>
              <Show when={props.invoice!.paid_at}>
                <p class="text-xs tabular-nums text-faint">
                  Paid on{" "}
                  {new Date(props.invoice!.paid_at!).toLocaleDateString()}
                </p>
              </Show>
            </div>
            <div class="flex flex-wrap gap-2">
              <Button
                as="a"
                href={props.invoice!.invoice_url!}
                target="_blank"
                rel="noopener noreferrer"
                variant="secondary"
              >
                View invoice
              </Button>
              <Show when={can(PERMISSION_ACTIONS.CREATE_INVOICE) && props.invoice!.status === "sent"}>
                <Button onClick={handleMarkPaid} disabled={markPaid.isPending}>
                  {markPaid.isPending ? "Saving..." : "Mark as paid"}
                </Button>
              </Show>
              <Show when={can(PERMISSION_ACTIONS.CREATE_INVOICE) && props.invoice!.status !== "paid"}>
                <Dialog open={isVoidDialogOpen()} onOpenChange={setIsVoidDialogOpen}>
                  <DialogTrigger as={Button} variant="destructive">
                    Void invoice
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>Void invoice</DialogTitle>
                    </DialogHeader>
                    <p class="text-sm text-muted-foreground">
                      Are you sure you want to void this invoice? The project
                      status will remain unchanged. You can attach a new invoice
                      afterwards.
                    </p>
                    <div class="flex justify-end gap-3 mt-4">
                      <DialogClose
                        as={Button}
                        variant="outline"
                        disabled={voidInvoice.isPending}
                      >
                        Cancel
                      </DialogClose>
                      <Button
                        variant="destructive"
                        onClick={handleVoid}
                        disabled={voidInvoice.isPending}
                      >
                        {voidInvoice.isPending ? "Voiding..." : "Void invoice"}
                      </Button>
                    </div>
                  </DialogContent>
                </Dialog>
              </Show>
            </div>
          </div>
        </Match>
      </Switch>
    </div>
  );
}
