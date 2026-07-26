import { For, Show, createMemo, createSignal } from "solid-js";
import { Button } from "@glassact/ui";
import { useQuery } from "@tanstack/solid-query";
import { IoChevronDown, IoChevronForward, IoDownloadOutline } from "solid-icons/io";
import type { GET, InlayProof } from "@glassact/data";
import { getProofsByInlayOpts } from "../../queries/proof";
import { ProofStatusBadge } from "../proof/proof-status-badge";

interface DesignHistoryProps {
  inlayUuid: string;
  // The pending proof already gets a full panel of its own, so skip it here
  // rather than rendering the same design twice.
  excludeProofId?: number;
}

function ProofRow(props: { proof: GET<InlayProof>; defaultOpen: boolean }) {
  const [isOpen, setIsOpen] = createSignal(props.defaultOpen);

  return (
    <li class="overflow-hidden rounded-lg border">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen())}
        class="flex w-full items-center gap-2 px-3 py-2 text-left transition-colors hover:bg-gray-50"
      >
        <span class="text-gray-400">
          <Show when={isOpen()} fallback={<IoChevronForward size={14} />}>
            <IoChevronDown size={14} />
          </Show>
        </span>
        <span class="text-sm font-medium text-gray-900">
          Version {props.proof.version_number}
        </span>
        <span class="flex-1" />
        <ProofStatusBadge status={props.proof.status} />
      </button>

      <Show when={isOpen()}>
        <div class="space-y-3 border-t p-3">
          <Show when={props.proof.design_asset_url}>
            <div class="flex items-center justify-center rounded bg-gray-50 p-3">
              <img
                src={props.proof.design_asset_url}
                alt={`Proof v${props.proof.version_number}`}
                class="max-h-40 max-w-full rounded object-contain"
              />
            </div>
          </Show>

          <div class="space-y-1 text-xs text-gray-600">
            <p class="font-medium">
              {props.proof.width}" &times; {props.proof.height}"
            </p>
            <Show when={props.proof.approved_at}>
              <p class="text-green-600">
                Approved {new Date(props.proof.approved_at!).toLocaleDateString()}
              </p>
            </Show>
            <Show when={props.proof.declined_at}>
              <p class="text-red-600">
                Declined {new Date(props.proof.declined_at!).toLocaleDateString()}
              </p>
            </Show>
            <Show when={props.proof.decline_reason}>
              <p class="rounded border border-red-200 bg-red-50 p-2 text-red-600">
                Reason: {props.proof.decline_reason}
              </p>
            </Show>
            <p class="text-gray-400">
              Created {new Date(props.proof.created_at).toLocaleString()}
            </p>
          </div>

          <Show when={props.proof.design_asset_url}>
            <Button
              variant="outline"
              size="sm"
              as="a"
              href={props.proof.design_asset_url}
              download
            >
              <IoDownloadOutline class="mr-2" size={16} />
              Download Design
            </Button>
          </Show>
        </div>
      </Show>
    </li>
  );
}

// Every design version this inlay has had, newest first. A stock or
// once-approved customized inlay has nothing interesting here, so the whole
// section collapses to a single line in that case.
export function DesignHistory(props: DesignHistoryProps) {
  const query = useQuery(() => getProofsByInlayOpts(props.inlayUuid));

  const proofs = createMemo(() =>
    [...(query.data ?? [])]
      .filter((proof) => proof.id !== props.excludeProofId)
      .reverse(),
  );

  const [isExpanded, setIsExpanded] = createSignal(false);
  const isWorthExpanding = createMemo(() => proofs().length > 1);

  return (
    <Show when={proofs().length > 0}>
      <section class="space-y-3">
        <button
          type="button"
          onClick={() => setIsExpanded(!isExpanded())}
          class="flex w-full items-center gap-2 text-left"
        >
          <span class="text-gray-400">
            <Show when={isExpanded()} fallback={<IoChevronForward size={14} />}>
              <IoChevronDown size={14} />
            </Show>
          </span>
          <h2 class="text-sm font-semibold text-gray-700">
            Design history ({proofs().length})
          </h2>
        </button>

        <Show when={isExpanded()}>
          <ul class="space-y-2">
            <For each={proofs()}>
              {(proof, index) => (
                <ProofRow
                  proof={proof}
                  defaultOpen={index() === 0 && isWorthExpanding()}
                />
              )}
            </For>
          </ul>
        </Show>
      </section>
    </Show>
  );
}
