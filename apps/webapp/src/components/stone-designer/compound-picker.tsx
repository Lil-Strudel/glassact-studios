import { For, Show, createSignal } from "solid-js";
import { useQuery } from "@tanstack/solid-query";
import { Button, cn } from "@glassact/ui";
import { browseCompoundsOpts } from "../../queries/compound";
import { CompoundPreview } from "./compound-preview";

interface CompoundPickerProps {
  onAdd: (compoundUuid: string) => Promise<void>;
  class?: string;
}

export function CompoundPicker(props: CompoundPickerProps) {
  const compoundsQuery = useQuery(() => browseCompoundsOpts());
  const [addingUuid, setAddingUuid] = createSignal<string | null>(null);

  async function add(uuid: string) {
    setAddingUuid(uuid);
    try {
      await props.onAdd(uuid);
    } finally {
      setAddingUuid(null);
    }
  }

  return (
    <div class={cn("flex flex-col gap-3", props.class)}>
      <Show when={compoundsQuery.isLoading}>
        <div class="h-24 animate-pulse rounded-md bg-gray-100" />
      </Show>
      <Show when={compoundsQuery.isSuccess && compoundsQuery.data.length === 0}>
        <p class="text-xs text-gray-500">No compound inlays are offered yet.</p>
      </Show>
      <For each={compoundsQuery.data ?? []}>
        {(compound) => (
          <div class="rounded-md border border-gray-200 p-2">
            <CompoundPreview members={compound.members} class="h-28 w-full" />
            <div class="mt-2 flex items-center justify-between gap-2">
              <div class="min-w-0">
                <p class="truncate text-sm font-medium text-gray-900">{compound.name}</p>
                <p class="text-xs text-gray-500">{compound.members.length} inlays</p>
              </div>
              <Button
                size="sm"
                variant="outline"
                disabled={addingUuid() !== null}
                onClick={() => add(compound.uuid)}
              >
                {addingUuid() === compound.uuid ? "Adding..." : "Add all"}
              </Button>
            </div>
          </div>
        )}
      </For>
    </div>
  );
}
