import { createQuery } from "@tanstack/solid-query";
import { createContext, createEffect, useContext } from "solid-js";
import { createStore, type SetStoreFunction } from "solid-js/store";
import { ParentComponent } from "solid-js/types/server/rendering.js";
import { postAuthTokenAccessOpts } from "../queries/auth";

type AuthState =
  | {
      status: "unauthenticated" | "pending";
    }
  | {
      status: "authenticated";
      accessToken: string;
      accessTokenExp: Date;
    };

export const AuthContext = createContext<{
  state: AuthState;
  setState: SetStoreFunction<AuthState>;
}>();

export function useAuthContext() {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("Can't find AuthContext");
  }
  return context;
}

export const AuthProvider: ParentComponent = (props) => {
  const [state, setState] = createStore<AuthState>({
    status: "pending",
  });

  const queryOptions = postAuthTokenAccessOpts();
  queryOptions.staleTime = Infinity;
  queryOptions.retry = false;

  const query = createQuery(() => queryOptions);

  createEffect(() => {
    switch (query.status) {
      case "success": {
        setState({
          status: "authenticated",
          accessToken: query.data.access_token,
          accessTokenExp: new Date(query.data.access_token_exp),
        });
        break;
      }

      case "error": {
        setState("status", "unauthenticated");
        break;
      }
      default: {
        break;
      }
    }
  });

  return (
    <AuthContext.Provider value={{ state, setState }}>
      {props.children}
    </AuthContext.Provider>
  );
};
