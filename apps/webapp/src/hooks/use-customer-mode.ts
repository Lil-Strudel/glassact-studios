import { createSignal, type Accessor } from "solid-js";
import { useUserContext } from "../providers/user";

const STORAGE_KEY = "glassact:customer-mode";

function readStoredCustomerMode() {
  try {
    return localStorage.getItem(STORAGE_KEY) === "on";
  } catch {
    // Private-browsing / blocked storage: start with prices shown.
    return false;
  }
}

function writeStoredCustomerMode(isOn: boolean) {
  try {
    if (isOn) localStorage.setItem(STORAGE_KEY, "on");
    else localStorage.removeItem(STORAGE_KEY);
  } catch {
    // The mode simply won't survive a reload.
  }
}

// Module-level so turning it on in a project keeps prices hidden on every page
// the dealer walks a family through (catalog, inlay pages, the stone designer).
const [isTurnedOn, setIsTurnedOn] = createSignal(readStoredCustomerMode());

export interface CustomerMode {
  isCustomerMode: Accessor<boolean>;
  setCustomerMode: (isOn: boolean) => void;
}

export function useCustomerMode(): CustomerMode {
  const userContext = useUserContext();

  return {
    // Staff never see the toggle, so a leftover flag from a dealership login on
    // the same browser must not hide prices from them.
    isCustomerMode: () => isTurnedOn() && userContext.isDealership(),
    setCustomerMode: (isOn) => {
      setIsTurnedOn(isOn);
      writeStoredCustomerMode(isOn);
    },
  };
}
