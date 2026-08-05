import { useQuery } from "@tanstack/solid-query";
import {
  createSignal,
  createContext,
  createEffect,
  onCleanup,
  onMount,
  untrack,
  useContext,
  ParentComponent,
  Setter,
} from "solid-js";
import {
  getValidAccessToken,
  refreshAccessToken,
  setOnSessionExpired,
} from "../queries/access-token";
import { DeferredPromise } from "../utils/deferred-promise";

type AuthStatus = "pending" | "unauthenticated" | "authenticated";
type SettledAuthStatus = "unauthenticated" | "authenticated";

export interface AuthState {
  status: () => AuthStatus;
  setStatus: Setter<AuthStatus>;
  deferredStatus: () => DeferredPromise<SettledAuthStatus>;
}
export const AuthContext = createContext<AuthState>();

export function useAuthContext() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("Can't find AuthContext");
  }
  return context;
}

export const AuthProvider: ParentComponent = (props) => {
  const [status, setStatus] = createSignal<AuthStatus>("pending");
  const [deferredStatus] = createSignal(
    new DeferredPromise<SettledAuthStatus>(),
  );

  // This query only answers "is there a session?" for the router gate. Keeping
  // the bearer fresh afterwards belongs to the api client's interceptor.
  const queryOptions = () => ({
    queryKey: ["token", "authentication"],
    queryFn: () => refreshAccessToken(),
    staleTime: Infinity,
    retry: false,
  });

  const query = useQuery(queryOptions);

  onMount(() => {
    // Fired imperatively from the api layer, so every status read here is a
    // point-in-time check rather than a subscription.
    setOnSessionExpired(() => {
      // A first-load 401 is the ordinary signed-out path, already handled by the
      // router's beforeLoad — redirecting on it would loop the login page.
      if (untrack(status) !== "authenticated") return;

      setStatus("unauthenticated");

      // deferredStatus resolves once ever, so the router's gate cannot re-run in
      // place. A full navigation re-arms it and drops the stale query cache.
      const redirect = window.location.pathname + window.location.search;
      window.location.assign(`/login?redirect=${encodeURIComponent(redirect)}`);
    });

    // The interceptor already guarantees correctness; topping up on wake just
    // spares the user a refresh round trip on their first click back.
    const topUpToken = () => {
      if (untrack(status) !== "authenticated") return;
      if (document.visibilityState !== "visible") return;

      void getValidAccessToken().catch(() => {});
    };

    document.addEventListener("visibilitychange", topUpToken);
    window.addEventListener("focus", topUpToken);
    window.addEventListener("online", topUpToken);

    onCleanup(() => {
      document.removeEventListener("visibilitychange", topUpToken);
      window.removeEventListener("focus", topUpToken);
      window.removeEventListener("online", topUpToken);
    });
  });

  createEffect(() => {
    switch (query.status) {
      case "success": {
        setStatus("authenticated");
        deferredStatus().resolve("authenticated");
        break;
      }

      case "error": {
        setStatus("unauthenticated");
        deferredStatus().resolve("unauthenticated");
        break;
      }
      default: {
        break;
      }
    }
  });

  return (
    <AuthContext.Provider value={{ status, setStatus, deferredStatus }}>
      {props.children}
    </AuthContext.Provider>
  );
};
