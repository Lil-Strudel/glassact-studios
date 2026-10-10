import { createMemo, createSignal, type Accessor } from "solid-js";
import {
  CUSTOM_GRANITE_KEY,
  DEFAULT_CUSTOM_HEX,
  DEFAULT_GRANITE_KEY,
  graniteChoice,
  isGraniteKey,
  type GraniteChoice,
} from "../components/granite/granite";

const KEY_STORAGE_KEY = "glassact:catalog-granite";
const CUSTOM_HEX_STORAGE_KEY = "glassact:catalog-granite-custom";
const HEX_PATTERN = /^#[0-9a-f]{6}$/i;

function readStored(storageKey: string, isValid: (value: string) => boolean) {
  try {
    const stored = localStorage.getItem(storageKey);
    if (stored && isValid(stored)) return stored;
  } catch {
    // Private-browsing / blocked storage: fall through to the fallback.
  }
  return null;
}

function writeStored(storageKey: string, value: string) {
  try {
    localStorage.setItem(storageKey, value);
  } catch {
    // Preference simply won't survive a reload.
  }
}

// null = the user has not picked a stone yet, so each surface falls back to the
// backdrop that suits it (the catalog opens plain, the customizer on gray).
// Module-level so every surface (catalog grid, detail modal, add-inlay browse,
// filter sidebar, customizer) reads and writes one shared choice without prop
// drilling it through unrelated components.
const [storedKey, setStoredKey] = createSignal(
  readStored(KEY_STORAGE_KEY, isGraniteKey),
);
const [customHex, setStoredCustomHex] = createSignal(
  readStored(CUSTOM_HEX_STORAGE_KEY, (value) => HEX_PATTERN.test(value)) ??
    DEFAULT_CUSTOM_HEX,
);

export interface GranitePreference {
  graniteKey: Accessor<string>;
  granite: Accessor<GraniteChoice>;
  customHex: Accessor<string>;
  setGraniteKey: (key: string) => void;
  // Picking a color also switches the backdrop to it.
  setCustomHex: (hex: string) => void;
}

export function useGranitePreference(
  fallbackKey: string = DEFAULT_GRANITE_KEY,
): GranitePreference {
  const graniteKey = createMemo(() => storedKey() ?? fallbackKey);
  const granite = createMemo(() => graniteChoice(graniteKey(), customHex()));

  function setGraniteKey(key: string) {
    setStoredKey(key);
    writeStored(KEY_STORAGE_KEY, key);
  }

  return {
    graniteKey,
    granite,
    customHex,
    setGraniteKey,
    setCustomHex: (hex: string) => {
      if (!HEX_PATTERN.test(hex)) return;
      setStoredCustomHex(hex);
      writeStored(CUSTOM_HEX_STORAGE_KEY, hex);
      setGraniteKey(CUSTOM_GRANITE_KEY);
    },
  };
}
