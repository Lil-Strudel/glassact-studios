import { For, ParentProps, Show } from "solid-js";
import type { GET, SandblastBackground } from "@glassact/data";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@glassact/ui";
import { DownloadSandblastButton } from "./download-sandblast-button";

interface BackgroundDetailModalProps {
  background: GET<SandblastBackground>;
}

export function BackgroundDetailModal(props: ParentProps<BackgroundDetailModalProps>) {
  return (
    <Dialog>
      <DialogTrigger as="div" class="contents">
        {props.children}
      </DialogTrigger>
      <DialogContent class="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{props.background.name}</DialogTitle>
        </DialogHeader>

        <div class="flex flex-col gap-6">
          <div class="flex min-h-[320px] items-center justify-center rounded-md bg-gray-50 p-4">
            <img
              src={props.background.svg_url}
              alt={props.background.name}
              class="max-h-[400px] max-w-full object-contain brightness-0"
            />
          </div>

          <div class="flex flex-wrap items-center justify-between gap-3">
            <code class="rounded bg-gray-100 px-2 py-1 text-sm">{props.background.code}</code>
            <DownloadSandblastButton backgroundUuid={props.background.uuid} />
          </div>

          <Show when={props.background.description}>
            {(description) => <p class="text-sm text-gray-600">{description()}</p>}
          </Show>

          <p class="text-sm text-gray-600">
            Place it on a stone in a project's designer, or download the file
            for your sandblaster.
          </p>

          <Show when={props.background.related_items.length > 0}>
            <div class="flex flex-col gap-2 border-t border-gray-200 pt-4">
              <h3 class="text-sm font-medium text-gray-900">Inlays that go with it</h3>
              <div class="grid grid-cols-3 gap-3 sm:grid-cols-4">
                <For each={props.background.related_items}>
                  {(item) => (
                    <div class="flex flex-col items-center gap-1 rounded-md border border-gray-200 p-2">
                      <img src={item.svg_url} alt={item.name} loading="lazy" class="h-16 w-full object-contain" />
                      <span class="w-full truncate text-center text-xs text-gray-600" title={item.name}>
                        {item.catalog_code}
                      </span>
                    </div>
                  )}
                </For>
              </div>
            </div>
          </Show>
        </div>
      </DialogContent>
    </Dialog>
  );
}
