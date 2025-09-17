import { GET, User } from "@glassact/data";
import {
  createContext,
  createSignal,
  ParentComponent,
  Setter,
  useContext,
} from "solid-js";
import { DeferredPromise } from "../utils/deferred-promise";
import { getUserSelfOpts } from "../queries/user";
import { useQuery } from "@tanstack/solid-query";
import { useAuthContext } from "./auth";

type UserStatus = "pending" | "success" | "failure";
type SettledUserStatus = "success" | "failure";

export interface UserState {
  status: () => UserStatus;
  setStatus: Setter<UserStatus>;
  deferredStatus: () => DeferredPromise<SettledUserStatus>;
  user: () => GET<User>;
}

export const UserContext = createContext<UserState>();

export function useUserContext() {
  const context = useContext(UserContext);
  if (!context) {
    throw new Error("Can't find UserContext");
  }
  return context;
}

const placeholderUser: GET<User> = {
  id: -1,
  uuid: "",
  name: "",
  email: "",
  avatar: "",
  dealership_id: -1,
  role: "user",
  created_at: "",
  updated_at: "",
  version: -1,
};

export const UserProvider: ParentComponent = (props) => {
  const [status, setStatus] = createSignal<UserStatus>("pending");
  const [deferredStatus] = createSignal(
    new DeferredPromise<SettledUserStatus>(),
  );

  const auth = useAuthContext();

  const queryOptions = () => {
    const queryOptions = getUserSelfOpts();
    queryOptions.staleTime = Infinity;
    queryOptions.retry = false;
    queryOptions.refetchInterval = 1000 * 60 * 15;
    queryOptions.enabled = auth.status() === "authenticated";

    return queryOptions;
  };

  const query = useQuery(queryOptions);

  const user = () => (query.isSuccess ? query.data : placeholderUser);

  return (
    <UserContext.Provider value={{ status, setStatus, deferredStatus, user }}>
      {props.children}
    </UserContext.Provider>
  );
};
