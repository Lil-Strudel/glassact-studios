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
} from "@glassact/ui";
import { createMemo, For, Match, Show, Switch } from "solid-js";
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
import { Can } from "../../components/Can";
import { isApiError } from "../../utils/is-api-error";
import { IoTrashOutline, IoAddCircleOutline } from "solid-icons/io";

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
  const placeOrder = useMutation(() => postPlaceOrderOpts());
  const submitProject = useMutation(postSubmitProjectOpts);
  const excludeInlay = useMutation(patchExcludeInlayOpts);

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
    const inlays = inlaysQuery.data ?? [];
    return inlays.filter((inlay) => !inlay.excluded_from_order);
  });

  const canSubmit = createMemo(() => {
    if (!projectQuery.isSuccess) return false;
    if (projectQuery.data.status !== "draft") return false;
    return includedInlays().length > 0;
  });

  const canPlaceOrder = createMemo(() => {
    if (!projectQuery.isSuccess) return false;
    if (projectQuery.data.status !== "approved") return false;
    const included = includedInlays();
    if (included.length === 0) return false;
    return included.every((inlay) => inlay.approved_proof_id !== null);
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

  function handlePlaceOrder() {
    if (!projectQuery.isSuccess) return;
    placeOrder.mutate(projectQuery.data.uuid, {
      onSuccess() {
        showToast({
          title: "Order placed",
          description: `Order for ${projectQuery.data!.name} has been placed successfully.`,
          variant: "success",
        });
        queryClient.invalidateQueries({ queryKey: ["project"] });
        queryClient.invalidateQueries({ queryKey: ["project", params().id, "inlays"] });
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
                <Dialog>
                  <DialogTrigger as={Button} disabled={!canPlaceOrder()}>
                    Place Order
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>Place Order</DialogTitle>
                    </DialogHeader>
                    <div class="space-y-4">
                      <p class="text-sm text-gray-600">
                        You are about to place an order for{" "}
                        <span class="font-semibold">{projectQuery.data!.name}</span>.
                        This will lock pricing and begin manufacturing.
                      </p>
                      <Show
                        when={
                          (inlaysQuery.data ?? []).length !==
                          includedInlays().length
                        }
                      >
                        <p class="text-sm text-gray-500">
                          Ordering {includedInlays().length} of{" "}
                          {(inlaysQuery.data ?? []).length} inlays (
                          {(inlaysQuery.data ?? []).length -
                            includedInlays().length}{" "}
                          excluded).
                        </p>
                      </Show>
                      <div class="border rounded-lg divide-y">
                        <For each={includedInlays()}>
                          {(inlay) => (
                            <div class="p-3 flex items-center justify-between">
                              <div class="flex items-center gap-2">
                                <Show when={inlay.preview_url}>
                                  <img
                                    src={inlay.preview_url}
                                    alt={inlay.name}
                                    class="w-8 h-8 object-contain rounded"
                                  />
                                </Show>
                                <span class="text-sm font-medium">{inlay.name}</span>
                              </div>
                              <Show when={inlay.approved_proof_id}>
                                <Badge variant="outline" class="bg-green-50 text-green-700 border-green-200 text-xs">
                                  Approved
                                </Badge>
                              </Show>
                            </div>
                          )}
                        </For>
                      </div>
                    </div>
                    <DialogFooter class="flex justify-end gap-3 mt-4">
                      <DialogClose as={Button} variant="outline" disabled={placeOrder.isPending}>
                        Cancel
                      </DialogClose>
                      <Button
                        onClick={handlePlaceOrder}
                        disabled={placeOrder.isPending}
                      >
                        {placeOrder.isPending ? "Placing Order..." : "Confirm Order"}
                      </Button>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
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
  const description = () => {
    if (props.inlay.type === "catalog" && props.inlay.catalog_info) {
      return props.inlay.catalog_info.customization_notes || null;
    }
    if (props.inlay.type === "custom" && props.inlay.custom_info) {
      return props.inlay.custom_info.description || null;
    }
    return null;
  };

  const proofStatusBadge = () => {
    if (props.inlay.approved_proof_id) {
      return { label: "Proof Approved", class: "bg-green-50 text-green-700 border-green-200" };
    }
    return null;
  };

  const isExcluded = () => props.inlay.excluded_from_order;

  return (
    <Card
      class={`overflow-hidden transition-opacity ${isExcluded() ? "opacity-50" : ""}`}
    >
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
          <Show when={proofStatusBadge()}>
            {(badge) => (
              <Badge variant="outline" class={`text-xs ${badge().class}`}>
                {badge().label}
              </Badge>
            )}
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
