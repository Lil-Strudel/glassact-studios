import { createMemo, createSignal, type Accessor } from "solid-js";
import {
  DEFAULT_GRANITE_KEY,
  GRANITE_PRESETS,
  graniteByKey,
  type GranitePreset,
} from "../components/granite/granite";

const STORAGE_KEY = "glassact:catalog-granite";

// null = the user has not picked a stone yet, so each surface falls back to the
// backdrop that suits it (the catalog opens plain, the customizer on gray).
function readStoredKey(): string | null {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored && GRANITE_PRESETS.some((preset) => preset.key === stored)) {
      return stored;
    }
  } catch {
    // Private-browsing / blocked storage: fall through to the fallback.
  }
  return null;
}

// Module-level so every surface (catalog grid, detail modal, add-inlay browse,
// filter sidebar, customizer) reads and writes one shared choice without prop
// drilling it through unrelated components.
const [storedKey, setStoredKey] = createSignal(readStoredKey());

export function useGranitePreference(
  fallbackKey: string = DEFAULT_GRANITE_KEY,
): {
  graniteKey: Accessor<string>;
  granite: Accessor<GranitePreset>;
  setGraniteKey: (key: string) => void;
} {
  const graniteKey = createMemo(() => storedKey() ?? fallbackKey);
  const granite = createMemo(() => graniteByKey(graniteKey()));

  return {
    graniteKey,
    granite,
    setGraniteKey: (key: string) => {
      setStoredKey(key);
      try {
        localStorage.setItem(STORAGE_KEY, key);
      } catch {
        // Preference simply won't survive a reload.
      }
    },
  };
}
