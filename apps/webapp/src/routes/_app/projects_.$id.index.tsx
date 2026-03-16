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
  DialogFooter,
  showToast,
  Checkbox,
  CheckboxControl,
} from "@glassact/ui";
import { createMemo, createSignal, For, Match, Show, Switch } from "solid-js";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import {
  getProjectOpts,
  deleteProjectOpts,
  postSubmitProjectOpts,
} from "../../queries/project";
import {
  getInlaysByProjectOpts,
  deleteInlayOpts,
  patchExcludeInlayOpts,
} from "../../queries/inlay";
import { postPlaceOrderOpts } from "../../queries/order";
import type { ProjectStatus, InlayWithInfo } from "@glassact/data";
import { ProjectStatusBadge } from "../../components/project/status-badge";
import { ProofStatusBadge } from "../../components/proof/proof-status-badge";
import { Can } from "../../components/Can";
import { useUserContext } from "../../providers/user";
import { isApiError } from "../../utils/is-api-error";
import { IoTrashOutline, IoAddCircleOutline, IoCheckmarkCircle } from "solid-icons/io";

export const Route = createFileRoute("/_app/projects_/$id/")({
  component: RouteComponent,
});

const STATUS_STEPS: ProjectStatus[] = [
  "draft",
  "designing",
  "pending-approval",
  "approved",
  "ordered",
  "in-production",
  "shipped",
  "delivered",
  "invoiced",
  "completed",
];

const STATUS_LABELS: Record<string, string> = {
  draft: "Draft",
  designing: "Designing",
  "pending-approval": "Pending Approval",
  approved: "Approved",
  ordered: "Ordered",
  "in-production": "In Production",
  shipped: "Shipped",
  delivered: "Delivered",
  invoiced: "Invoiced",
  completed: "Completed",
};

const PRE_ORDERED_STATUSES: ProjectStatus[] = [
  "draft",
  "designing",
  "pending-approval",
  "approved",
];

const EDITABLE_STATUSES: ProjectStatus[] = ["draft", "designing"];

function RouteComponent() {
  const params = Route.useParams();
  const queryClient = useQueryClient();

  const projectQuery = useQuery(() => getProjectOpts(params().id));
  const inlaysQuery = useQuery(() => getInlaysByProjectOpts(params().id));
  const cancelProject = useMutation(deleteProjectOpts);
  const removeInlay = useMutation(deleteInlayOpts);
  const submitProject = useMutation(postSubmitProjectOpts);
  const excludeInlay = useMutation(patchExcludeInlayOpts);

  const inlays = () => (inlaysQuery.isSuccess ? inlaysQuery.data : []);

  const canCancel = createMemo(() => {
    if (!projectQuery.isSuccess) return false;
    return PRE_ORDERED_STATUSES.includes(projectQuery.data.status);
  });

  const canEditInlays = createMemo(() => {
    if (!projectQuery.isSuccess) return false;
    return EDITABLE_STATUSES.includes(projectQuery.data.status);
  });

  const canExcludeInlays = createMemo(() => {
    if (!projectQuery.isSuccess) return false;
    return PRE_ORDERED_STATUSES.includes(projectQuery.data.status);
  });

  const includedInlays = createMemo(() => {
    return inlays().filter((inlay) => !inlay.excluded_from_order);
  });

  const canSubmit = createMemo(() => {
    if (!projectQuery.isSuccess) return false;
    if (projectQuery.data.status !== "draft") return false;
    return includedInlays().length > 0;
  });



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

  function handleSubmit() {
    if (!projectQuery.isSuccess) return;
    submitProject.mutate(projectQuery.data.uuid, {
      onSuccess() {
        showToast({
          title: "Project submitted",
          description: `${projectQuery.data!.name} has been submitted for design.`,
          variant: "success",
        });
        queryClient.invalidateQueries({ queryKey: ["project"] });
      },
      onError(error) {
        if (isApiError(error)) {
          showToast({
            title: "Failed to submit project",
            description: error?.data?.error ?? "Unknown error",
            variant: "error",
          });
        }
      },
    });
  }

  function handleExcludeInlay(inlay: InlayWithInfo, excluded: boolean) {
    excludeInlay.mutate(
      { uuid: inlay.uuid, excluded },
      {
        onSuccess() {
          showToast({
            title: excluded ? "Inlay excluded" : "Inlay included",
            description: excluded
              ? `${inlay.name} has been excluded from the order.`
              : `${inlay.name} has been included in the order.`,
            variant: "success",
          });
          queryClient.invalidateQueries({
            queryKey: ["project", params().id, "inlays"],
          });
        },
        onError(error) {
          if (isApiError(error)) {
            showToast({
              title: "Failed to update inlay",
              description: error?.data?.error ?? "Unknown error",
              variant: "error",
            });
          }
        },
      },
    );
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

          <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
            <div class="flex items-center gap-3">
              <h1 class="text-2xl font-bold text-gray-900">
                {projectQuery.data!.name}
              </h1>
              <ProjectStatusBadge status={projectQuery.data!.status} />
            </div>
            <div class="flex gap-3">
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
              <Can permission="create_project">
                <Show when={projectQuery.data!.status === "draft"}>
                  <Dialog>
                    <DialogTrigger as={Button} disabled={!canSubmit()}>
                      Submit for Design
                    </DialogTrigger>
                    <DialogContent>
                      <DialogHeader>
                        <DialogTitle>Submit for Design</DialogTitle>
                      </DialogHeader>
                      <div class="space-y-3">
                        <p class="text-sm text-gray-600">
                          Submit{" "}
                          <span class="font-semibold">
                            {projectQuery.data!.name}
                          </span>{" "}
                          for design? The GlassAct team will begin creating
                          proofs for your inlays.
                        </p>
                        <p class="text-sm text-gray-500">
                          {includedInlays().length} inlay
                          {includedInlays().length !== 1 ? "s" : ""} will be
                          submitted.
                        </p>
                      </div>
                      <DialogFooter class="flex justify-end gap-3 mt-4">
                        <DialogClose
                          as={Button}
                          variant="outline"
                          disabled={submitProject.isPending}
                        >
                          Cancel
                        </DialogClose>
                        <Button
                          onClick={handleSubmit}
                          disabled={submitProject.isPending}
                        >
                          {submitProject.isPending
                            ? "Submitting..."
                            : "Submit for Design"}
                        </Button>
                      </DialogFooter>
                    </DialogContent>
                  </Dialog>
                </Show>
              </Can>
              <Can permission="place_order">
                <Show when={projectQuery.data!.status === "approved"}>
                  <PlaceOrderDialog
                    project={projectQuery.data!}
                    inlays={inlays()}
                    onSuccess={() => {
                      queryClient.invalidateQueries({ queryKey: ["project"] });
                      queryClient.invalidateQueries({
                        queryKey: ["project", params().id, "inlays"],
                      });
                    }}
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
                <Can permission="create_project">
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
                    <Can permission="create_project">
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
                        canDelete={canEditInlays()}
                        onDelete={() => handleDeleteInlay(inlay)}
                        isDeleting={removeInlay.isPending}
                        canExclude={canExcludeInlays()}
                        onToggleExclude={(excluded) =>
                          handleExcludeInlay(inlay, excluded)
                        }
                        isExcluding={excludeInlay.isPending}
                      />
                    )}
                  </For>
                </div>
              </Match>
            </Switch>
          </div>
        </div>
      </Match>
    </Switch>
  );
}

interface InlayCardProps {
  inlay: InlayWithInfo;
  projectId: string;
  canDelete: boolean;
  onDelete: () => void;
  isDeleting: boolean;
  canExclude: boolean;
  onToggleExclude: (excluded: boolean) => void;
  isExcluding: boolean;
}

function InlayCard(props: InlayCardProps) {
  const { isDealership, isInternal } = useUserContext();

  const description = () => {
    if (props.inlay.type === "catalog" && props.inlay.catalog_info) {
      return props.inlay.catalog_info.customization_notes || null;
    }
    if (props.inlay.type === "custom" && props.inlay.custom_info) {
      return props.inlay.custom_info.description || null;
    }
    return null;
  };

  const isExcluded = () => props.inlay.excluded_from_order;

  const needsAction = createMemo(() => {
    if (isExcluded()) return false;

    if (isDealership()) {
      return props.inlay.has_pending_proof === true;
    }

    if (isInternal()) {
      return (
        !props.inlay.approved_proof_id && !props.inlay.has_pending_proof
      );
    }

    return false;
  });

  return (
    <Card
      class={`overflow-hidden transition-opacity relative ${isExcluded() ? "opacity-50" : ""}`}
    >
      <Show when={needsAction()}>
        <span
          class="absolute top-2 right-2 z-10 w-3 h-3 bg-orange-500 rounded-full border-2 border-white shadow-sm"
          title={
            isDealership() ? "Proof awaiting approval" : "Needs proof"
          }
        />
      </Show>
      <Link
        to="/projects/$id/inlay/$inlayId"
        params={{ id: props.projectId, inlayId: props.inlay.uuid }}
        class="block hover:bg-gray-50/50 transition-colors"
      >
        <Show when={props.inlay.preview_url}>
          <div class="bg-gray-50 p-4 flex items-center justify-center h-40 overflow-hidden relative">
            <img
              src={props.inlay.preview_url}
              alt={props.inlay.name}
              class={`max-w-full max-h-full object-contain ${isExcluded() ? "grayscale" : ""}`}
            />
            <Show when={isExcluded()}>
              <div class="absolute inset-0 flex items-center justify-center">
                <Badge variant="secondary" class="text-xs">
                  Excluded
                </Badge>
              </div>
            </Show>
          </div>
        </Show>
        <Show when={!props.inlay.preview_url}>
          <div class="bg-gray-100 p-4 flex items-center justify-center h-40">
            <p class="text-gray-400 text-sm">
              {isExcluded() ? "Excluded" : "No preview"}
            </p>
          </div>
        </Show>
        <CardHeader class="space-y-2">
          <div class="flex items-start justify-between gap-2">
            <CardTitle
              class={`text-sm truncate ${isExcluded() ? "line-through text-gray-400" : ""}`}
            >
              {props.inlay.name}
            </CardTitle>
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
          <Show when={props.inlay.approved_proof_id}>
            <ProofStatusBadge status="approved" class="text-xs" />
          </Show>
          <Show
            when={
              !props.inlay.approved_proof_id && props.inlay.has_pending_proof
            }
          >
            <ProofStatusBadge status="pending" class="text-xs" />
          </Show>
        </CardHeader>
      </Link>
      <Show when={props.canDelete || props.canExclude}>
        <div class="px-6 pb-4 flex flex-col gap-2">
          <Show when={props.canExclude}>
            <Button
              variant="ghost"
              size="sm"
              class={
                isExcluded()
                  ? "text-green-600 hover:text-green-700 hover:bg-green-50 w-full"
                  : "text-gray-600 hover:text-gray-700 hover:bg-gray-50 w-full"
              }
              onClick={(e: MouseEvent) => {
                e.stopPropagation();
                props.onToggleExclude(!isExcluded());
              }}
              disabled={props.isExcluding}
            >
              {isExcluded() ? "Include in Order" : "Exclude from Order"}
            </Button>
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

interface PlaceOrderDialogProps {
  project: { uuid: string; name: string };
  inlays: InlayWithInfo[];
  onSuccess: () => void;
}

function PlaceOrderDialog(props: PlaceOrderDialogProps) {
  const [selectedInlays, setSelectedInlays] = createSignal<Set<string>>(
    new Set(),
  );
  const [orderSuccess, setOrderSuccess] = createSignal(false);
  const placeOrder = useMutation(() => postPlaceOrderOpts());

  const initializeSelection = () => {
    const eligibleInlays = props.inlays
      .filter((inlay) => inlay.approved_proof_id && !inlay.excluded_from_order)
      .map((inlay) => inlay.uuid);
    setSelectedInlays(new Set(eligibleInlays));
  };

  const toggleInlay = (uuid: string) => {
    const inlay = props.inlays.find((i) => i.uuid === uuid);
    if (!inlay?.approved_proof_id) return;

    const current = selectedInlays();
    const updated = new Set(current);
    if (updated.has(uuid)) {
      updated.delete(uuid);
    } else {
      updated.add(uuid);
    }
    setSelectedInlays(updated);
  };

  const selectedCount = () => selectedInlays().size;

  const totalPriceCents = createMemo(() => {
    return props.inlays
      .filter((inlay) => selectedInlays().has(inlay.uuid))
      .reduce((sum, inlay) => sum + (inlay.approved_proof_price_cents ?? 0), 0);
  });

  const formatPrice = (cents: number) => {
    return `$${(cents / 100).toFixed(2)}`;
  };

  const canConfirmOrder = () => selectedCount() > 0;

  const handlePlaceOrder = () => {
    placeOrder.mutate(
      {
        projectUuid: props.project.uuid,
        inlayUuids: Array.from(selectedInlays()),
      },
      {
        onSuccess() {
          setOrderSuccess(true);
          props.onSuccess();
        },
        onError(error) {
          if (isApiError(error)) {
            showToast({
              title: "Failed to place order",
              description: error?.data?.error ?? "Unknown error",
              variant: "error",
            });
          }
        },
      },
    );
  };

  const handleOpenChange = (open: boolean) => {
    if (open) {
      initializeSelection();
      setOrderSuccess(false);
    }
  };

  return (
    <Dialog onOpenChange={handleOpenChange}>
      <DialogTrigger as={Button}>Place Order</DialogTrigger>
      <DialogContent class="max-w-lg">
        <Show
          when={!orderSuccess()}
          fallback={
            <div class="text-center py-8 space-y-4">
              <div class="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto">
                <IoCheckmarkCircle class="w-10 h-10 text-green-600" />
              </div>
              <h3 class="text-xl font-semibold text-gray-900">
                Order Placed Successfully!
              </h3>
              <p class="text-gray-600">
                Your order for{" "}
                <span class="font-medium">{props.project.name}</span> has been
                submitted. Manufacturing will begin shortly.
              </p>
              <DialogClose as={Button} class="w-full">
                Close
              </DialogClose>
            </div>
          }
        >
          <DialogHeader>
            <DialogTitle>Place Order</DialogTitle>
          </DialogHeader>
          <div class="space-y-4">
            <p class="text-sm text-gray-600">
              Select the inlays to include in this order. Only inlays with
              approved proofs can be ordered.
            </p>

            <div class="border rounded-lg divide-y max-h-64 overflow-y-auto">
              <For each={props.inlays}>
                {(inlay) => {
                  const isEligible = () => !!inlay.approved_proof_id;
                  const isSelected = () => selectedInlays().has(inlay.uuid);

                  return (
                    <Checkbox
                      checked={isSelected()}
                      onChange={() => toggleInlay(inlay.uuid)}
                      disabled={!isEligible()}
                      class="w-full"
                    >
                      <label
                        class={`p-3 flex items-center gap-3 cursor-pointer hover:bg-gray-50 w-full ${
                          !isEligible() ? "opacity-50 cursor-not-allowed" : ""
                        }`}
                      >
                        <CheckboxControl />
                        <Show
                          when={inlay.preview_url}
                          fallback={
                            <div class="w-10 h-10 bg-gray-100 rounded flex items-center justify-center text-gray-400 text-xs">
                              N/A
                            </div>
                          }
                        >
                          <img
                            src={inlay.preview_url}
                            alt={inlay.name}
                            class="w-10 h-10 object-contain rounded"
                          />
                        </Show>
                        <div class="flex-1 min-w-0">
                          <p class="text-sm font-medium truncate">
                            {inlay.name}
                          </p>
                          <Show when={inlay.approved_proof_price_group_name}>
                            <p class="text-xs text-gray-500">
                              {inlay.approved_proof_price_group_name}
                            </p>
                          </Show>
                        </div>
                        <Show
                          when={isEligible()}
                          fallback={
                            <Badge variant="outline" class="text-xs shrink-0">
                              No Proof
                            </Badge>
                          }
                        >
                          <div class="flex items-center gap-2 shrink-0">
                            <Show when={inlay.approved_proof_price_cents}>
                              <span class="text-sm font-medium">
                                {formatPrice(
                                  inlay.approved_proof_price_cents!,
                                )}
                              </span>
                            </Show>
                            <ProofStatusBadge status="approved" class="text-xs" />
                          </div>
                        </Show>
                      </label>
                    </Checkbox>
                  );
                }}
              </For>
            </div>

            <div class="flex justify-between items-center pt-2 border-t">
              <span class="text-sm text-gray-600">
                {selectedCount()} inlay{selectedCount() !== 1 ? "s" : ""}{" "}
                selected
              </span>
              <Show when={totalPriceCents() > 0}>
                <span class="text-lg font-semibold">
                  {formatPrice(totalPriceCents())}
                </span>
              </Show>
            </div>
          </div>
          <DialogFooter class="flex justify-end gap-3 mt-4">
            <DialogClose
              as={Button}
              variant="outline"
              disabled={placeOrder.isPending}
            >
              Cancel
            </DialogClose>
            <Button
              onClick={handlePlaceOrder}
              disabled={placeOrder.isPending || !canConfirmOrder()}
            >
              {placeOrder.isPending ? "Placing Order..." : "Confirm Order"}
            </Button>
          </DialogFooter>
        </Show>
      </DialogContent>
    </Dialog>
  );
}
