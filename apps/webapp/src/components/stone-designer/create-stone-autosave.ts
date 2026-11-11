import { createSignal, on, onCleanup, createEffect, type Accessor } from "solid-js";
import { isAxiosError } from "axios";
import type { StoneEditor } from "./create-stone-editor";
import { putStoneDesign } from "../../queries/stone-design";
import { requestFromDraft } from "./stone-draft";
import type { GET, StoneDesign } from "@glassact/data";

export type SaveStatus = "saved" | "pending" | "saving" | "error" | "conflict";

const SAVE_DELAY_MS = 800;

interface StoneAutosaveOptions {
  editor: StoneEditor;
  projectUuid: Accessor<string>;
  initialVersion: number;
  isEnabled: Accessor<boolean>;
  onSaved: (design: GET<StoneDesign>) => void;
  onConflict: () => void;
}

export function createStoneAutosave(options: StoneAutosaveOptions) {
  const [version, setVersion] = createSignal(options.initialVersion);
  const [status, setStatus] = createSignal<SaveStatus>("saved");
  let timer: ReturnType<typeof setTimeout> | undefined;
  let inflight: Promise<void> | null = null;
  let isDirty = false;

  async function flush() {
    clearTimeout(timer);
    if (inflight) await inflight;
    if (!isDirty || !options.isEnabled()) return;
    isDirty = false;
    setStatus("saving");

    inflight = putStoneDesign({
      projectUuid: options.projectUuid(),
      body: requestFromDraft(options.editor.draft, version()),
    })
      .then((saved) => {
        setVersion(saved.version);
        options.onSaved(saved);
        setStatus(isDirty ? "pending" : "saved");
      })
      .catch((error: unknown) => {
        if (isAxiosError(error) && error.response?.status === 409) {
          setStatus("conflict");
          options.onConflict();
          return;
        }
        isDirty = true;
        setStatus("error");
      })
      .finally(() => {
        inflight = null;
      });
    await inflight;
  }

  createEffect(
    on(
      options.editor.revision,
      () => {
        if (!options.isEnabled()) return;
        isDirty = true;
        setStatus("pending");
        clearTimeout(timer);
        timer = setTimeout(flush, SAVE_DELAY_MS);
      },
      { defer: true },
    ),
  );

  function warnBeforeUnload(e: BeforeUnloadEvent) {
    if (isDirty || inflight) e.preventDefault();
  }
  window.addEventListener("beforeunload", warnBeforeUnload);

  onCleanup(() => {
    window.removeEventListener("beforeunload", warnBeforeUnload);
    void flush();
  });

  return { status, version, setVersion, flush };
}
