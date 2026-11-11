import { For, createMemo } from "solid-js";
import { STONE_SHAPES, type StoneShape } from "@glassact/data";
import { cn } from "@glassact/ui";
import { GranitePicker } from "../granite/granite-picker";
import { CUSTOM_GRANITE_KEY, graniteChoice } from "../granite/granite";
import type { GranitePreference } from "../../hooks/use-granite-preference";
import type { StoneEditor } from "./create-stone-editor";
import { STONE_SHAPE_INFO, stoneOutlinePath, type StoneSize } from "./stone-shapes";

interface StonePanelProps {
  editor: StoneEditor;
}

const MIN_INCHES = 6;
const MAX_INCHES = 120;

export function StonePanel(props: StonePanelProps) {
  const draft = () => props.editor.draft;

  const granite = createMemo(() => graniteChoice(draft().graniteKey, draft().customHex));
  const granitePreference: GranitePreference = {
    graniteKey: () => draft().graniteKey,
    granite,
    customHex: () => draft().customHex,
    setGraniteKey: (key) => props.editor.update((d) => (d.graniteKey = key)),
    setCustomHex: (hex) =>
      props.editor.update((d) => {
        d.customHex = hex;
        d.graniteKey = CUSTOM_GRANITE_KEY;
      }),
  };

  // Changing the stone keeps everything where it sits relative to the
  // stone's center, so a resize does not strand the layout in one corner.
  function resizeStone(size: StoneSize, shape: StoneShape = draft().shape) {
    const dx = (size.width - draft().width) / 2;
    const dy = (size.height - draft().height) / 2;
    props.editor.update((d) => {
      d.shape = shape;
      d.width = size.width;
      d.height = size.height;
      d.layout.elements.forEach((el) => {
        el.x += dx;
        el.y += dy;
      });
    });
  }

  function chooseShape(shape: StoneShape) {
    const sameFamily = STONE_SHAPE_INFO[shape].presets.some(
      (p) => p.width === draft().width && p.height === draft().height,
    );
    resizeStone(
      sameFamily ? { width: draft().width, height: draft().height } : STONE_SHAPE_INFO[shape].defaultSize,
      shape,
    );
  }

  function setDimension(key: keyof StoneSize, raw: string) {
    const value = Number(raw);
    if (!Number.isFinite(value)) return;
    const clamped = Math.min(MAX_INCHES, Math.max(MIN_INCHES, value));
    resizeStone({ width: draft().width, height: draft().height, [key]: clamped });
  }

  return (
    <div class="flex flex-col gap-6">
      <section>
        <h3 class="text-sm font-medium text-gray-900">Shape</h3>
        <div class="mt-2 grid grid-cols-4 gap-2">
          <For each={STONE_SHAPES}>
            {(shape) => {
              const info = STONE_SHAPE_INFO[shape];
              const isSelected = () => draft().shape === shape;
              return (
                <button
                  type="button"
                  title={info.label}
                  class={cn(
                    "flex flex-col items-center gap-1 rounded-md border p-2 text-[11px] text-gray-700 transition",
                    isSelected() ? "border-blue-600 ring-2 ring-blue-500/30" : "border-gray-200 hover:border-gray-400",
                  )}
                  onClick={() => chooseShape(shape)}
                >
                  <svg
                    viewBox={`-1 -1 ${info.defaultSize.width + 2} ${info.defaultSize.height + 2}`}
                    class="h-8 w-full"
                  >
                    <path
                      d={stoneOutlinePath(shape, info.defaultSize.width, info.defaultSize.height)}
                      fill="#9ca3af"
                    />
                  </svg>
                  {info.label}
                </button>
              );
            }}
          </For>
        </div>
      </section>

      <section>
        <h3 class="text-sm font-medium text-gray-900">Size (inches)</h3>
        <div class="mt-2 flex items-center gap-2">
          <label class="flex flex-1 flex-col text-xs text-gray-600">
            Width
            <input
              type="number"
              min={MIN_INCHES}
              max={MAX_INCHES}
              step="0.5"
              value={draft().width}
              onChange={(e) => setDimension("width", e.currentTarget.value)}
              class="mt-1 rounded-md border border-gray-300 px-2 py-1 text-sm"
            />
          </label>
          <span class="mt-5 text-gray-400">×</span>
          <label class="flex flex-1 flex-col text-xs text-gray-600">
            Height
            <input
              type="number"
              min={MIN_INCHES}
              max={MAX_INCHES}
              step="0.5"
              value={draft().height}
              onChange={(e) => setDimension("height", e.currentTarget.value)}
              class="mt-1 rounded-md border border-gray-300 px-2 py-1 text-sm"
            />
          </label>
        </div>
        <div class="mt-2 flex flex-wrap gap-1.5">
          <For each={STONE_SHAPE_INFO[draft().shape].presets}>
            {(preset) => (
              <button
                type="button"
                class={cn(
                  "rounded-full border px-2.5 py-0.5 text-xs",
                  preset.width === draft().width && preset.height === draft().height
                    ? "border-blue-600 bg-blue-50 text-blue-700"
                    : "border-gray-300 text-gray-700 hover:border-gray-500",
                )}
                onClick={() => resizeStone(preset)}
              >
                {preset.width}″ × {preset.height}″
              </button>
            )}
          </For>
        </div>
      </section>

      <section>
        <h3 class="text-sm font-medium text-gray-900">Granite</h3>
        <GranitePicker preference={granitePreference} class="mt-2" />
      </section>
    </div>
  );
}
