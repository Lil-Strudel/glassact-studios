import { createFileRoute, Link } from "@tanstack/solid-router";
import {
  Breadcrumb,
  Button,
  Badge,
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  Dialog,
  DialogClose,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  showToast,
} from "@glassact/ui";
import { createMemo, createSignal, For, Match, Show, Switch } from "solid-js";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import {
  getProjectOpts,
  deleteProjectOpts,
  patchProjectOpts,
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
} from "@glassact/data";
import { PERMISSION_ACTIONS } from "@glassact/data";
import { ProjectStatusBadge } from "../../components/project/status-badge";
import { Can } from "../../components/Can";
import { useUserContext } from "../../providers/user";
import { isApiError } from "../../utils/is-api-error";
import { formatMoney } from "../../utils/format-money";
import { formatPriceFormula } from "../../utils/format-price-formula";
import { PlaceOrderCart } from "../../components/place-order-cart";
import {
  IoTrashOutline,
  IoAddCircleOutline,
  IoPencilOutline,
} from "solid-icons/io";
import { ManufacturingTracker } from "../../components/manufacturing/manufacturing-tracker";

export const Route = createFileRoute("/_app/projects_/$id/")({
  component: RouteComponent,
});

const STATUS_STEPS: ProjectStatus[] = [
  "draft",
  "ordered",
  "in-production",
  "shipped",
  "delivered",
  "invoiced",
  "completed",
];

const STATUS_LABELS: Record<ProjectStatus, string> = {
  draft: "Draft",
  ordered: "Ordered",
  "in-production": "In Production",
  shipped: "Shipped",
  delivered: "Delivered",
  invoiced: "Invoiced",
  completed: "Completed",
  cancelled: "Cancelled",
};

const EDITABLE_STATUSES: ProjectStatus[] = ["draft"];
const CANCELLABLE_STATUSES: ProjectStatus[] = ["draft", "ordered"];
const MANUFACTURING_STATUSES: ProjectStatus[] = [
  "ordered",
  "in-production",
  "shipped",
  "delivered",
];

function isManufacturingStatus(status: ProjectStatus): boolean {
  return MANUFACTURING_STATUSES.includes(status);
}

function RouteComponent() {
  const params = Route.useParams();
  const queryClient = useQueryClient();

  const projectQuery = useQuery(() => getProjectOpts(params().id));
  const inlaysQuery = useQuery(() => getInlaysByProjectOpts(params().id));
  const cancelProject = useMutation(deleteProjectOpts);
  const removeInlay = useMutation(deleteInlayOpts);

  const INVOICE_STATUSES: ProjectStatus[] = [
    "delivered",
    "invoiced",
    "completed",
  ];
  const showInvoiceSection = createMemo(() => {
    if (!projectQuery.isSuccess) return false;
    return INVOICE_STATUSES.includes(projectQuery.data.status);
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

  const totalPriceCents = createMemo(() =>
    inlays().reduce((sum, inlay) => sum + (inlay.price_cents ?? 0), 0),
  );

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
        <div class="space-y-6">
          <div class="h-5 w-64 bg-gray-200 rounded animate-pulse" />
          <div class="h-8 w-48 bg-gray-200 rounded animate-pulse" />
          <div class="h-4 w-32 bg-gray-100 rounded animate-pulse" />
          <div class="flex gap-2 overflow-x-auto py-4">
            <For each={Array.from({ length: 10 })}>
              {() => (
                <div class="h-8 w-24 bg-gray-200 rounded-full animate-pulse flex-shrink-0" />
              )}
            </For>
          </div>
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4 mt-8">
            <For each={[1, 2, 3]}>
              {() => <div class="h-48 bg-gray-200 rounded-lg animate-pulse" />}
            </For>
          </div>
        </div>
      </Match>

      <Match when={projectQuery.isError}>
        <div class="border-2 border-dashed border-red-300 rounded-xl p-8 text-center">
          <p class="text-red-600 font-medium">Failed to load project</p>
          <p class="text-gray-500 text-sm mt-1">
            {projectQuery.error?.message ?? "An unexpected error occurred."}
          </p>
          <Button
            variant="outline"
            class="mt-4"
            onClick={() => projectQuery.refetch()}
          >
            Retry
          </Button>
        </div>
      </Match>

      <Match when={projectQuery.isSuccess}>
        <div>
          <Breadcrumb
            crumbs={[
              { title: "Projects", to: "/projects" },
              {
                title: projectQuery.data!.name,
                to: `/projects/${projectQuery.data!.uuid}`,
              },
            ]}
          />

          <div class="flex flex-col sm:flex-row sm:items-start sm:justify-between gap-4">
            <div class="flex flex-col gap-2 min-w-0">
              <div class="flex items-center gap-3 flex-wrap">
                <h1 class="text-2xl font-bold text-gray-900">
                  {projectQuery.data!.name}
                </h1>
                <ProjectStatusBadge status={projectQuery.data!.status} />
              </div>
              <InternalReferenceField
                projectUuid={projectQuery.data!.uuid}
                value={projectQuery.data!.internal_reference}
              />
              <p class="text-sm text-gray-500">
                Project total:{" "}
                <span class="font-semibold text-gray-900">
                  {formatMoney(totalPriceCents() / 100)}
                </span>
              </p>
            </div>
            <div class="flex gap-3 flex-wrap">
              <Show when={canCancel()}>
                <Dialog>
                  <DialogTrigger as={Button} variant="outline">
                    Cancel Project
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>Cancel Project</DialogTitle>
                    </DialogHeader>
                    <p class="text-sm text-gray-600">
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
                          : "Cancel Project"}
                      </Button>
                    </div>
                  </DialogContent>
                </Dialog>
              </Show>
              <Can permission={PERMISSION_ACTIONS.PLACE_ORDER}>
                <Show when={projectQuery.data!.status === "draft"}>
                  <PlaceOrderCart
                    project={projectQuery.data!}
                    inlays={inlays()}
                    disabled={!canPlaceOrder()}
                  />
                </Show>
              </Can>
            </div>
          </div>

          <Show when={projectQuery.data!.status !== "cancelled"}>
            <div class="mt-6 overflow-x-auto">
              <div class="flex items-center gap-1 min-w-max">
                <For each={STATUS_STEPS}>
                  {(step, index) => {
                    const isCurrent = () => index() === currentStepIndex();
                    const isComplete = () => index() < currentStepIndex();

                    return (
                      <div class="flex items-center">
                        <Show when={index() > 0}>
                          <div
                            class={`w-6 h-0.5 ${isComplete() ? "bg-primary" : "bg-gray-300"}`}
                          />
                        </Show>
                        <div
                          class={`px-3 py-1 rounded-full text-xs font-medium whitespace-nowrap ${
                            isCurrent()
                              ? "bg-primary text-white"
                              : isComplete()
                                ? "bg-primary/10 text-primary"
                                : "bg-gray-100 text-gray-500"
                          }`}
                        >
                          {STATUS_LABELS[step]}
                        </div>
                      </div>
                    );
                  }}
                </For>
              </div>
            </div>
          </Show>

          <div class="mt-8">
            <div class="flex items-center justify-between mb-4">
              <h2 class="text-lg font-semibold text-gray-900">Inlays</h2>
              <Show when={canEditInlays()}>
                <Can permission={PERMISSION_ACTIONS.CREATE_PROJECT}>
                  <Button
                    as={Link}
                    to={`/projects/${params().id}/add-inlay`}
                    variant="outline"
                    size="sm"
                  >
                    <IoAddCircleOutline size={18} class="mr-1" />
                    Add Inlay
                  </Button>
                </Can>
              </Show>
            </div>

            <Switch>
              <Match when={inlaysQuery.isLoading}>
                <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                  <For each={[1, 2, 3]}>
                    {() => (
                      <div class="h-48 bg-gray-200 rounded-lg animate-pulse" />
                    )}
                  </For>
                </div>
              </Match>

              <Match when={inlaysQuery.isError}>
                <div class="border-2 border-dashed border-red-300 rounded-xl p-8 text-center">
                  <p class="text-red-600 font-medium">Failed to load inlays</p>
                  <Button
                    variant="outline"
                    class="mt-4"
                    onClick={() => inlaysQuery.refetch()}
                  >
                    Retry
                  </Button>
                </div>
              </Match>

              <Match
                when={inlaysQuery.isSuccess && inlaysQuery.data!.length === 0}
              >
                <div class="border-2 border-dashed border-gray-300 rounded-xl p-8 text-center">
                  <p class="text-gray-400 text-lg font-medium">No inlays yet</p>
                  <p class="text-gray-400 text-sm mt-2">
                    Add inlays to this project to get started.
                  </p>
                  <Show when={canEditInlays()}>
                    <Can permission={PERMISSION_ACTIONS.CREATE_PROJECT}>
                      <Button
                        as={Link}
                        to={`/projects/${params().id}/add-inlay`}
                        variant="outline"
                        class="mt-4"
                      >
                        Add Inlay
                      </Button>
                    </Can>
                  </Show>
                </div>
              </Match>

              <Match
                when={inlaysQuery.isSuccess && inlaysQuery.data!.length > 0}
              >
                <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                  <For each={inlaysQuery.data!}>
                    {(inlay) => (
                      <InlayCard
                        inlay={inlay}
                        projectId={params().id}
                        projectStatus={projectQuery.data!.status}
                        canDelete={
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
          </div>

          <Show when={showInvoiceSection()}>
            <div class="mt-8">
              <h2 class="text-lg font-semibold text-gray-900 mb-4">Invoice</h2>
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
            </div>
          </Show>
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
    <div class="flex items-center gap-2 text-sm text-gray-600">
      <span class="font-medium">PO / Reference:</span>
      <Show
        when={isEditing()}
        fallback={
          <>
            <span class="text-gray-900">{props.value || "—"}</span>
            <Show when={canEdit()}>
              <Button
                size="sm"
                variant="ghost"
                onClick={startEdit}
                class="h-7 px-2"
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
          class="border rounded-md px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
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
  canDelete: boolean;
  onDelete: () => void;
  isDeleting: boolean;
}

type ReadinessBadge =
  | { label: string; variant: "default" | "warning" | "outline" }
  | null;

function readinessBadge(inlay: InlayWithInfo): ReadinessBadge {
  if (inlay.is_ready) {
    return { label: "Ready", variant: "default" };
  }

  const hasPending = inlay.has_pending_proof === true;

  if (inlay.is_customized && inlay.type === "catalog" && hasPending) {
    return { label: "Needs Internal Review", variant: "warning" };
  }

  if (inlay.type === "custom" && hasPending) {
    return { label: "Needs Customer Approval", variant: "warning" };
  }

  if (inlay.type === "custom" && !hasPending) {
    return { label: "Awaiting Proof", variant: "outline" };
  }

  return { label: "Awaiting Proof", variant: "outline" };
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

  return (
    <Card class="overflow-hidden">
      <Link
        to="/projects/$id/inlay/$inlayId"
        params={{ id: props.projectId, inlayId: props.inlay.uuid }}
        class="block hover:bg-gray-50/50 transition-colors"
      >
        <Show when={props.inlay.preview_url}>
          <div class="bg-gray-50 p-4 flex items-center justify-center h-40 overflow-hidden">
            <img
              src={props.inlay.preview_url}
              alt={props.inlay.name}
              class="max-w-full max-h-full object-contain"
            />
          </div>
        </Show>
        <Show when={!props.inlay.preview_url}>
          <div class="bg-gray-100 p-4 flex items-center justify-center h-40">
            <p class="text-gray-400 text-sm">No preview</p>
          </div>
        </Show>
        <CardHeader class="space-y-2">
          <div class="flex items-start justify-between gap-2">
            <CardTitle class="text-sm truncate">{props.inlay.name}</CardTitle>
            <Badge variant="outline" class="text-xs flex-shrink-0">
              {props.inlay.type === "catalog" ? "Catalog" : "Custom"}
            </Badge>
          </div>
          <Show when={description()}>
            {(desc) => (
              <CardDescription class="text-xs line-clamp-2">
                {desc()}
              </CardDescription>
            )}
          </Show>
          <div class="flex items-center gap-2 flex-wrap">
            <Show when={badge()}>
              {(b) => (
                <Badge variant={b().variant} class="text-xs">
                  {b().label}
                </Badge>
              )}
            </Show>
            <Show when={props.inlay.is_customized}>
              <Badge variant="warning" class="text-xs">
                Customized
              </Badge>
            </Show>
          </div>
          <div class="flex items-end justify-between text-xs text-gray-600">
            <span>{priceLabel()}</span>
            <Show when={priceFormula()}>
              <span class="text-gray-500">{priceFormula()}</span>
            </Show>
          </div>
          <Show when={showManufacturingTracker()}>
            <div class="pt-1">
              <ManufacturingTracker
                currentStep={
                  props.inlay.manufacturing_step as ManufacturingStep
                }
                hasBlocker={props.inlay.has_active_blocker}
              />
            </div>
          </Show>
        </CardHeader>
      </Link>
      <Show when={props.canDelete || showInternalApprove()}>
        <div class="px-6 pb-4 flex flex-col gap-2">
          <Show when={showInternalApprove()}>
            <Can permission={PERMISSION_ACTIONS.INTERNAL_APPROVE_PROOF}>
              {/* TODO: implement internal review dialog with price-group
                  override. The pending proof UUID is not exposed on the
                  inlay list endpoint; for now route to the inlay detail
                  page where the standard approve flow lives. */}
              <Button
                as={Link}
                to="/projects/$id/inlay/$inlayId"
                params={{
                  id: props.projectId,
                  inlayId: props.inlay.uuid,
                }}
                variant="outline"
                size="sm"
                class="w-full"
              >
                Review &amp; Approve
              </Button>
            </Can>
          </Show>
          <Show when={props.canDelete}>
            <Dialog>
              <DialogTrigger
                as={Button}
                variant="ghost"
                size="sm"
                class="text-red-600 hover:text-red-700 hover:bg-red-50 w-full"
              >
                <IoTrashOutline size={16} class="mr-1" />
                Remove
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Remove Inlay</DialogTitle>
                </DialogHeader>
                <p class="text-sm text-gray-600">
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
                    {props.isDeleting ? "Removing..." : "Remove Inlay"}
                  </Button>
                </div>
              </DialogContent>
            </Dialog>
          </Show>
        </div>
      </Show>
    </Card>
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
    <div class="border rounded-lg p-6 space-y-4">
      <Switch>
        <Match when={props.isLoading}>
          <div class="h-16 bg-gray-100 rounded animate-pulse" />
        </Match>

        <Match when={!hasActiveInvoice() && can(PERMISSION_ACTIONS.CREATE_INVOICE)}>
          <div class="space-y-3">
            <p class="text-sm text-gray-600">
              Paste the invoice link from your billing platform to attach it to
              this project.
            </p>
            <div class="flex gap-2">
              <input
                type="url"
                placeholder="https://..."
                value={invoiceUrl()}
                onInput={(e) => setInvoiceUrl(e.currentTarget.value)}
                class="flex-1 border rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              />
              <Button
                onClick={handleAttach}
                disabled={attachInvoice.isPending || invoiceUrl().trim() === ""}
              >
                {attachInvoice.isPending ? "Attaching..." : "Attach Invoice"}
              </Button>
            </div>
          </div>
        </Match>

        <Match when={!hasActiveInvoice() && !can(PERMISSION_ACTIONS.CREATE_INVOICE)}>
          <div class="text-center py-4">
            <p class="text-gray-500 text-sm">
              Invoice not yet available. Check back soon.
            </p>
          </div>
        </Match>

        <Match when={hasActiveInvoice()}>
          <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div class="space-y-1">
              <div class="flex items-center gap-2">
                <span class="text-sm font-medium text-gray-700">
                  Invoice Status:
                </span>
                <Badge
                  variant={
                    props.invoice!.status === "paid"
                      ? "default"
                      : props.invoice!.status === "sent"
                        ? "secondary"
                        : "outline"
                  }
                  class="text-xs capitalize"
                >
                  {props.invoice!.status}
                </Badge>
              </div>
              <Show when={props.invoice!.paid_at}>
                <p class="text-xs text-gray-500">
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
                variant="outline"
              >
                View Invoice
              </Button>
              <Show when={can(PERMISSION_ACTIONS.CREATE_INVOICE) && props.invoice!.status === "sent"}>
                <Button onClick={handleMarkPaid} disabled={markPaid.isPending}>
                  {markPaid.isPending ? "Saving..." : "Mark as Paid"}
                </Button>
              </Show>
              <Show when={can(PERMISSION_ACTIONS.CREATE_INVOICE) && props.invoice!.status !== "paid"}>
                <Dialog open={isVoidDialogOpen()} onOpenChange={setIsVoidDialogOpen}>
                  <DialogTrigger as={Button} variant="outline">
                    Void Invoice
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>Void Invoice</DialogTitle>
                    </DialogHeader>
                    <p class="text-sm text-gray-600">
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
                        {voidInvoice.isPending ? "Voiding..." : "Void Invoice"}
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
