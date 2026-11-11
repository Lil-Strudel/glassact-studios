import { Show, createMemo, createSignal, untrack } from "solid-js";
import { useQueryClient } from "@tanstack/solid-query";
import type { GET, InlayWithInfo, StoneDesign } from "@glassact/data";
import { Tabs, TabsContent, TabsIndicator, TabsList, TabsTrigger, showToast } from "@glassact/ui";
import { graniteChoice } from "../granite/granite";
import { getStoneDesign } from "../../queries/stone-design";
import { createStoneEditor } from "./create-stone-editor";
import { createStoneAutosave } from "./create-stone-autosave";
import { StoneCanvas } from "./stone-canvas";
import { draftFromDesign, inlayArt } from "./stone-draft";
import { useBackgroundArt } from "./use-background-art";
import { useBackgroundItems } from "./use-background-items";
import { useDesignerShortcuts } from "./use-designer-shortcuts";
import { DesignerToolbar } from "./designer-toolbar";
import { StonePanel } from "./stone-panel";
import { InlayPanel } from "./inlay-panel";
import { BackgroundPanel } from "./background-panel";
import { TextPanel } from "./text-panel";
import { LayersPanel } from "./layers-panel";

interface StoneDesignerProps {
  projectUuid: string;
  design: GET<StoneDesign> | null;
  inlays: InlayWithInfo[];
  isEditable: boolean;
  onAddCompound: (compoundUuid: string) => Promise<void>;
}

type PanelTab = "stone" | "inlays" | "background" | "text" | "layers";

export function StoneDesigner(props: StoneDesignerProps) {
  const queryClient = useQueryClient();
  // The editor owns the draft from here on; later prop changes to `design`
  // are applied explicitly (after a conflict or a compound add).
  // eslint-disable-next-line solid/reactivity
  const editor = createStoneEditor(draftFromDesign(props.design));
  const [tab, setTab] = createSignal<PanelTab>(
    untrack(() => props.isEditable) ? "inlays" : "background",
  );

  const designQueryKey = () => ["project", props.projectUuid, "stone-design"];

  async function reloadFromServer() {
    const latest = await getStoneDesign(props.projectUuid);
    queryClient.setQueryData(designQueryKey(), latest);
    editor.load(draftFromDesign(latest));
    autosave.setVersion(latest?.version ?? 0);
  }

  const autosave = createStoneAutosave({
    editor,
    projectUuid: () => props.projectUuid,
    // eslint-disable-next-line solid/reactivity
    initialVersion: props.design?.version ?? 0,
    isEnabled: () => props.isEditable,
    onSaved: (saved) => queryClient.setQueryData(designQueryKey(), saved),
    onConflict: () => {
      showToast({
        title: "The design changed elsewhere",
        description: "Loaded the latest version; your last edit was not saved.",
        variant: "error",
      });
      void reloadFromServer();
    },
  });

  useDesignerShortcuts(editor, () => props.isEditable);

  const inlaysByUuid = createMemo(
    () => new Map(props.inlays.map((inlay) => [inlay.uuid, inlay])),
  );
  const backgroundItems = useBackgroundItems();
  const resolveBackground = useBackgroundArt(() => editor.draft.layout.elements);

  const stone = createMemo(() => ({
    shape: editor.draft.shape,
    width: editor.draft.width,
    height: editor.draft.height,
    granite: graniteChoice(editor.draft.graniteKey, editor.draft.customHex),
  }));

  async function addCompound(compoundUuid: string) {
    await autosave.flush();
    await props.onAddCompound(compoundUuid);
    await reloadFromServer();
  }

  return (
    <div class="flex h-full min-h-0 flex-col gap-3 lg:flex-row">
      <div class="flex min-h-[420px] min-w-0 flex-1 flex-col gap-3">
        <Show when={props.isEditable}>
          <DesignerToolbar editor={editor} saveStatus={autosave.status()} />
        </Show>
        <StoneCanvas
          stone={stone()}
          elements={editor.draft.layout.elements}
          resolveInlay={(uuid) => {
            const inlay = inlaysByUuid().get(uuid);
            return inlay ? inlayArt(inlay) : undefined;
          }}
          resolveBackground={resolveBackground}
          editor={props.isEditable ? editor : undefined}
          isZoomable
          class="min-h-[420px] flex-1 rounded-lg border border-gray-200 bg-gradient-to-b from-sky-50 to-emerald-50"
        />
        <p class="text-xs text-gray-500">
          Drawn to scale: {editor.draft.width}″ × {editor.draft.height}″.
          <Show when={props.isEditable}>
            {" "}Drag to move, use the round handle to rotate, Shift to snap or
            multi-select, arrow keys to nudge.
          </Show>
        </p>
      </div>

      <aside class="w-full flex-shrink-0 overflow-y-auto rounded-lg border border-gray-200 bg-white lg:w-96">
        <Tabs value={tab()} onChange={(value) => setTab(value as PanelTab)}>
          <TabsList class="sticky top-0 z-10 w-full bg-white">
            <Show when={props.isEditable}>
              <TabsTrigger value="stone">Stone</TabsTrigger>
              <TabsTrigger value="inlays">Inlays</TabsTrigger>
            </Show>
            <TabsTrigger value="background">Background</TabsTrigger>
            <Show when={props.isEditable}>
              <TabsTrigger value="text">Text</TabsTrigger>
            </Show>
            <TabsTrigger value="layers">Layers</TabsTrigger>
            <TabsIndicator />
          </TabsList>
          <div class="p-4">
            <TabsContent value="stone">
              <StonePanel editor={editor} />
            </TabsContent>
            <TabsContent value="inlays">
              <InlayPanel
                editor={editor}
                inlays={props.inlays}
                canAddCompounds={props.isEditable}
                onAddCompound={addCompound}
              />
            </TabsContent>
            <TabsContent value="background">
              <BackgroundPanel editor={editor} isEditable={props.isEditable} />
            </TabsContent>
            <TabsContent value="text">
              <TextPanel editor={editor} />
            </TabsContent>
            <TabsContent value="layers">
              <LayersPanel
                editor={editor}
                inlayName={(uuid) => inlaysByUuid().get(uuid)?.name}
                backgroundName={(uuid) => backgroundItems.byUuid().get(uuid)?.name}
              />
            </TabsContent>
          </div>
        </Tabs>
      </aside>
    </div>
  );
}
