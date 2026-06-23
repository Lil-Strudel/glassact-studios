import { createMemo, createSignal, For, Show } from "solid-js";
import { cn } from "./cn";
import { textfieldLabel } from "./textfield";
import { Badge } from "./badge";

// Freesolo (creatable) comboboxes: the user may pick from a suggestion list OR
// type an arbitrary value that is not in the list. Built on a plain input plus a
// suggestion popover rather than Kobalte's Combobox, because Kobalte's combobox
// is fixed-option (its value must be one of `options`) and fights a creatable
// flow. These keep the same idiom/styling as the rest of libs/ui.

const inputClass =
  "flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-shadow placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-[2px] focus-visible:ring-primary disabled:cursor-not-allowed disabled:opacity-50";

const suggestionListClass =
  "absolute left-0 right-0 top-full z-50 mt-1 max-h-60 overflow-y-auto rounded-md border border-input bg-popover text-popover-foreground shadow-md";

const suggestionItemClass =
  "w-full cursor-default select-none px-3 py-2 text-left text-sm outline-none hover:bg-accent hover:text-accent-foreground";

interface ComboboxFreeProps {
  /** Suggestions offered in the dropdown; the value is NOT restricted to these. */
  options: string[];
  /** Current committed value. Empty string means no value. */
  value: string;
  /** Fires with the new value on selection from the list or commit of typed text. */
  onValueChange: (value: string) => void;
  placeholder?: string;
  label?: string;
  description?: string;
  class?: string;
  disabled?: boolean;
}

/**
 * Single-value freesolo combobox. Selecting a suggestion or pressing Enter on
 * typed text commits the value. The displayed text always reflects `props.value`
 * unless the user is actively typing.
 */
export function ComboboxFree(props: ComboboxFreeProps) {
  const [draft, setDraft] = createSignal<string | null>(null);
  const [open, setOpen] = createSignal(false);

  const text = () => draft() ?? props.value;

  const filtered = createMemo(() => {
    const q = text().trim().toLowerCase();
    const opts = props.options;
    if (!q) return opts;
    return opts.filter((o) => o.toLowerCase().includes(q));
  });

  function commit(value: string) {
    props.onValueChange(value);
    setDraft(null);
    setOpen(false);
  }

  return (
    <div class={cn("relative flex flex-col gap-1", props.class)}>
      <Show when={props.label}>
        <label class={textfieldLabel()}>{props.label}</label>
      </Show>
      <div class="relative">
        <input
          type="text"
          disabled={props.disabled}
          value={text()}
          placeholder={props.placeholder}
          class={inputClass}
          onInput={(e) => {
            setDraft(e.currentTarget.value);
            setOpen(true);
          }}
          onFocus={() => setOpen(true)}
          onBlur={() => {
            // Commit whatever is typed; delay so a suggestion click lands first.
            setTimeout(() => {
              const d = draft();
              if (d !== null) commit(d.trim());
              setOpen(false);
            }, 150);
          }}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              e.preventDefault();
              commit(text().trim());
            } else if (e.key === "Escape") {
              setDraft(null);
              setOpen(false);
            }
          }}
        />
        <Show when={open() && filtered().length > 0}>
          <div class={suggestionListClass}>
            <For each={filtered()}>
              {(option) => (
                <button
                  type="button"
                  class={suggestionItemClass}
                  onMouseDown={(e) => {
                    e.preventDefault();
                    commit(option);
                  }}
                >
                  {option}
                </button>
              )}
            </For>
          </div>
        </Show>
      </div>
      <Show when={props.description}>
        <p class={textfieldLabel({ description: true, label: false })}>
          {props.description}
        </p>
      </Show>
    </div>
  );
}

interface ComboboxFreeMultiProps {
  /** Suggestions offered in the dropdown; values are NOT restricted to these. */
  options: string[];
  /** Current committed tags. */
  value: string[];
  /** Fires with the full next tag list on add or remove. */
  onValueChange: (value: string[]) => void;
  placeholder?: string;
  label?: string;
  description?: string;
  class?: string;
  disabled?: boolean;
}

/**
 * Multi-value (tags) freesolo combobox. Add a tag by pressing Enter on typed
 * text or selecting a suggestion; remove via the chip's ✕. Duplicates are
 * ignored. Already-selected values are hidden from the suggestion list.
 */
export function ComboboxFreeMulti(props: ComboboxFreeMultiProps) {
  const [input, setInput] = createSignal("");
  const [open, setOpen] = createSignal(false);

  const filtered = createMemo(() => {
    const q = input().trim().toLowerCase();
    const selected = new Set(props.value);
    return props.options
      .filter((o) => !selected.has(o))
      .filter((o) => (q ? o.toLowerCase().includes(q) : true));
  });

  function addTag(raw: string) {
    const tag = raw.trim();
    if (!tag || props.value.includes(tag)) {
      setInput("");
      return;
    }
    props.onValueChange([...props.value, tag]);
    setInput("");
  }

  function removeTag(tag: string) {
    props.onValueChange(props.value.filter((t) => t !== tag));
  }

  return (
    <div class={cn("relative flex flex-col gap-2", props.class)}>
      <Show when={props.label}>
        <label class={textfieldLabel()}>{props.label}</label>
      </Show>
      <div class="relative">
        <input
          type="text"
          disabled={props.disabled}
          value={input()}
          placeholder={props.placeholder}
          class={inputClass}
          onInput={(e) => {
            setInput(e.currentTarget.value);
            setOpen(true);
          }}
          onFocus={() => setOpen(true)}
          onBlur={() => setTimeout(() => setOpen(false), 150)}
          onKeyDown={(e) => {
            if (e.key === "Enter") {
              e.preventDefault();
              addTag(input());
            } else if (e.key === "Backspace" && input() === "") {
              const last = props.value[props.value.length - 1];
              if (last !== undefined) removeTag(last);
            } else if (e.key === "Escape") {
              setOpen(false);
            }
          }}
        />
        <Show when={open() && filtered().length > 0}>
          <div class={suggestionListClass}>
            <For each={filtered()}>
              {(option) => (
                <button
                  type="button"
                  class={suggestionItemClass}
                  onMouseDown={(e) => {
                    e.preventDefault();
                    addTag(option);
                  }}
                >
                  {option}
                </button>
              )}
            </For>
          </div>
        </Show>
      </div>
      <Show when={props.value.length > 0}>
        <div class="flex flex-wrap gap-2">
          <For each={props.value}>
            {(tag) => (
              <Badge variant="secondary" class="flex items-center gap-2">
                {tag}
                <button
                  type="button"
                  disabled={props.disabled}
                  onClick={() => removeTag(tag)}
                  class="ml-1 hover:text-destructive"
                  aria-label={`Remove ${tag}`}
                >
                  ✕
                </button>
              </Badge>
            )}
          </For>
        </div>
      </Show>
      <Show when={props.description}>
        <p class={textfieldLabel({ description: true, label: false })}>
          {props.description}
        </p>
      </Show>
    </div>
  );
}

export type { ComboboxFreeProps, ComboboxFreeMultiProps };
