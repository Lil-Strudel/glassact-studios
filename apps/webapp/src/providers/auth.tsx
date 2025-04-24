import { createQuery } from "@tanstack/solid-query";
import {
  createSignal,
  createContext,
  createEffect,
  useContext,
  ParentComponent,
} from "solid-js";
import { postAuthTokenAccessOpts } from "../queries/auth";
import api from "../queries/api";

type AuthStatus = "unauthenticated" | "pending" | "authenticated";
export const AuthContext = createContext<{
  status: () => AuthStatus;
  setStatus: (v: AuthStatus | ((prev: AuthStatus) => AuthStatus)) => AuthStatus;
}>();

export function useAuthContext() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("Can't find AuthContext");
  }
  return context;
}

export const AuthProvider: ParentComponent = (props) => {
  const [status, setStatus] = createSignal<AuthStatus>("pending");

  const queryOptions = postAuthTokenAccessOpts();
  queryOptions.staleTime = Infinity;
  queryOptions.retry = false;
  queryOptions.refetchInterval = 1000 * 60 * 60 * 1.5;

  const query = createQuery(() => queryOptions);

  createEffect(() => {
    switch (query.status) {
      case "success": {
        setStatus("authenticated");
        api.defaults.headers = {
          Authorization: `Bearer ${query.data.access_token}`,
        };
        break;
      }

      case "error": {
        setStatus("unauthenticated");
        break;
      }
      default: {
        break;
      }
    }
  });

  return (
    <AuthContext.Provider value={{ status, setStatus }}>
      {props.children}
    </AuthContext.Provider>
  );
};
