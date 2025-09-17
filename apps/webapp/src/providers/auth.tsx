import { useQuery } from "@tanstack/solid-query";
import {
  createSignal,
  createContext,
  createEffect,
  useContext,
  ParentComponent,
  Setter,
} from "solid-js";
import { postAuthTokenAccessOpts } from "../queries/auth";
import api from "../queries/api";
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

  const queryOptions = postAuthTokenAccessOpts();
  queryOptions.staleTime = Infinity;
  queryOptions.retry = false;
  queryOptions.refetchInterval = 1000 * 60 * 60 * 1.5;

  const query = useQuery(() => queryOptions);

  createEffect(() => {
    switch (query.status) {
      case "success": {
        setStatus("authenticated");
        deferredStatus().resolve("authenticated");
        api.defaults.headers = {
          Authorization: `Bearer ${query.data.access_token}`,
        };
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
