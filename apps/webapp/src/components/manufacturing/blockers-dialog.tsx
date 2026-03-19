import { createSignal, For, Match, Show, Switch } from "solid-js";
import {
  Badge,
  Button,
  Dialog,
  DialogContent,
  DialogFooter,
  DialogClose,
  DialogHeader,
  DialogTitle,
  Form,
  showToast,
} from "@glassact/ui";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import type { ManufacturingStep } from "@glassact/data";
import {
  getBlockersByInlayOpts,
  postBlockerOpts,
  postResolveBlockerOpts,
} from "../../queries/manufacturing";
import { Can } from "../Can";
import { isApiError } from "../../utils/is-api-error";

interface BlockersDialogProps {
  inlayUuid: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const MANUFACTURING_STEPS: ManufacturingStep[] = [
  "ordered",
  "materials-prep",
  "cutting",
  "fire-polish",
  "packaging",
  "shipped",
  "delivered",
];

const STEP_LABELS: Record<ManufacturingStep, string> = {
  ordered: "Ordered",
  "materials-prep": "Prepping Materials",
  cutting: "Cutting",
  "fire-polish": "Fire Polish",
  packaging: "Packaging",
  shipped: "Shipped",
  delivered: "Delivered",
};

const AddBlockerSchema = z.object({
  blocker_type: z.enum(["soft", "hard"]),
  reason: z.string().min(1, "Reason is required"),
  step_blocked: z.string().min(1, "Step is required"),
});

export function BlockersDialog(props: BlockersDialogProps) {
  const queryClient = useQueryClient();
  const [showAddForm, setShowAddForm] = createSignal(false);

  const blockersQuery = useQuery(() => getBlockersByInlayOpts(props.inlayUuid));
  const addBlocker = useMutation(() => postBlockerOpts());
  const resolveBlocker = useMutation(() => postResolveBlockerOpts());

  const unresolvedBlockers = () =>
    (blockersQuery.data ?? []).filter((b) => b.resolved_at === null);
  const resolvedBlockers = () =>
    (blockersQuery.data ?? []).filter((b) => b.resolved_at !== null);

  const form = createForm(() => ({
    defaultValues: {
      blocker_type: "soft" as "soft" | "hard",
      reason: "",
      step_blocked: "",
    } as z.output<typeof AddBlockerSchema>,
    validators: {
      onSubmit: AddBlockerSchema,
    },
    onSubmit: async ({ value }) => {
      addBlocker.mutate(
        {
          inlayUuid: props.inlayUuid,
          body: {
            blocker_type: value.blocker_type,
            reason: value.reason,
            step_blocked: value.step_blocked,
          },
        },
        {
          onSuccess() {
            form.reset();
            setShowAddForm(false);
            queryClient.invalidateQueries({
              queryKey: ["inlay", props.inlayUuid, "blockers"],
            });
            queryClient.invalidateQueries({ queryKey: ["kanban-inlays"] });
            showToast({
              title: "Blocker added",
              description: "The blocker has been recorded.",
              variant: "success",
            });
          },
          onError(error) {
            if (isApiError(error)) {
              showToast({
                title: "Failed to add blocker",
                description: error?.data?.error ?? "Unknown error",
                variant: "error",
              });
            }
          },
        },
      );
    },
  }));

  function handleResolve(blockerUuid: string) {
    resolveBlocker.mutate(
      { blockerUuid, body: {} },
      {
        onSuccess() {
          queryClient.invalidateQueries({
            queryKey: ["inlay", props.inlayUuid, "blockers"],
          });
          queryClient.invalidateQueries({ queryKey: ["kanban-inlays"] });
          showToast({
            title: "Blocker resolved",
            description: "The blocker has been marked as resolved.",
            variant: "success",
          });
        },
        onError(error) {
          if (isApiError(error)) {
            showToast({
              title: "Failed to resolve blocker",
              description: error?.data?.error ?? "Unknown error",
              variant: "error",
            });
          }
        },
      },
    );
  }

  return (
    <Dialog open={props.open} onOpenChange={props.onOpenChange}>
      <DialogContent class="max-w-lg max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Blockers</DialogTitle>
        </DialogHeader>

        <Switch>
          <Match when={blockersQuery.isLoading}>
            <div class="space-y-2">
              <For each={[1, 2]}>
                {() => <div class="h-16 bg-gray-100 rounded animate-pulse" />}
              </For>
            </div>
          </Match>

          <Match when={blockersQuery.isError}>
            <p class="text-sm text-red-600">Failed to load blockers.</p>
          </Match>

          <Match when={blockersQuery.isSuccess}>
            <div class="space-y-4">
              <Show when={unresolvedBlockers().length > 0}>
                <div class="space-y-2">
                  <p class="text-xs font-semibold uppercase tracking-wide text-gray-500">
                    Active
                  </p>
                  <For each={unresolvedBlockers()}>
                    {(blocker) => (
                      <div class="border border-red-200 bg-red-50 rounded-lg p-3 space-y-2">
                        <div class="flex items-start justify-between gap-2">
                          <div class="space-y-1 min-w-0">
                            <div class="flex items-center gap-2">
                              <Badge
                                variant="secondary"
                                class={`text-xs capitalize ${blocker.blocker_type === "hard" ? "bg-red-100 text-red-700 border-red-200" : ""}`}
                              >
                                {blocker.blocker_type}
                              </Badge>
                              <span class="text-xs text-gray-500">
                                Blocking:{" "}
                                {STEP_LABELS[
                                  blocker.step_blocked as ManufacturingStep
                                ] ?? blocker.step_blocked}
                              </span>
                            </div>
                            <p class="text-sm text-gray-800 break-words">
                              {blocker.reason}
                            </p>
                          </div>
                          <Can permission="create_blocker">
                            <Button
                              size="sm"
                              variant="outline"
                              class="flex-shrink-0 border-red-300 hover:bg-red-100"
                              onClick={() => handleResolve(blocker.uuid)}
                              disabled={resolveBlocker.isPending}
                            >
                              Resolve
                            </Button>
                          </Can>
                        </div>
                      </div>
                    )}
                  </For>
                </div>
              </Show>

              <Show when={resolvedBlockers().length > 0}>
                <div class="space-y-2">
                  <p class="text-xs font-semibold uppercase tracking-wide text-gray-500">
                    Resolved
                  </p>
                  <For each={resolvedBlockers()}>
                    {(blocker) => (
                      <div class="border border-gray-200 bg-gray-50 rounded-lg p-3 space-y-1 opacity-70">
                        <div class="flex items-center gap-2">
                          <Badge variant="outline" class="text-xs capitalize">
                            {blocker.blocker_type}
                          </Badge>
                          <span class="text-xs text-gray-500">
                            {STEP_LABELS[
                              blocker.step_blocked as ManufacturingStep
                            ] ?? blocker.step_blocked}
                          </span>
                        </div>
                        <p class="text-sm text-gray-600 break-words">
                          {blocker.reason}
                        </p>
                        <Show when={blocker.resolution_notes}>
                          <p class="text-xs text-gray-500 italic">
                            Resolution: {blocker.resolution_notes}
                          </p>
                        </Show>
                      </div>
                    )}
                  </For>
                </div>
              </Show>

              <Show
                when={
                  unresolvedBlockers().length === 0 &&
                  resolvedBlockers().length === 0
                }
              >
                <p class="text-sm text-gray-500 text-center py-4">
                  No blockers recorded.
                </p>
              </Show>

              <Can permission="create_blocker">
                <Show
                  when={showAddForm()}
                  fallback={
                    <Button
                      variant="outline"
                      class="w-full"
                      onClick={() => setShowAddForm(true)}
                    >
                      Add Blocker
                    </Button>
                  }
                >
                  <div class="border rounded-lg p-4 space-y-3 bg-gray-50">
                    <p class="text-sm font-medium">Add Blocker</p>
                    <form
                      onSubmit={(e) => {
                        e.preventDefault();
                        form.handleSubmit();
                      }}
                      class="space-y-3"
                    >
                      <form.Field
                        name="blocker_type"
                        children={(field) => (
                          <Form.Select
                            field={field}
                            label="Type"
                            options={[
                              { value: "soft", label: "Soft (Informational)" },
                              {
                                value: "hard",
                                label: "Hard (Blocks Progress)",
                              },
                            ]}
                            placeholder="Select type"
                          />
                        )}
                      />

                      <form.Field
                        name="step_blocked"
                        children={(field) => (
                          <Form.Select
                            field={field}
                            label="Step Blocked"
                            options={MANUFACTURING_STEPS.map((s) => ({
                              value: s,
                              label: STEP_LABELS[s],
                            }))}
                            placeholder="Select step"
                          />
                        )}
                      />

                      <form.Field
                        name="reason"
                        children={(field) => (
                          <Form.TextArea
                            field={field}
                            label="Reason"
                            placeholder="Describe the blocker..."
                          />
                        )}
                      />

                      <div class="flex gap-2">
                        <Button
                          type="button"
                          variant="outline"
                          class="flex-1"
                          onClick={() => {
                            setShowAddForm(false);
                            form.reset();
                          }}
                        >
                          Cancel
                        </Button>
                        <Button
                          type="submit"
                          class="flex-1"
                          disabled={addBlocker.isPending}
                        >
                          {addBlocker.isPending ? "Saving..." : "Save Blocker"}
                        </Button>
                      </div>
                    </form>
                  </div>
                </Show>
              </Can>
            </div>
          </Match>
        </Switch>

        <DialogFooter>
          <DialogClose as={Button} variant="outline">
            Close
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
