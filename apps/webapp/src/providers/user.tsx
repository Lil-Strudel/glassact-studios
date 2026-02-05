import { GET, User, isDealershipUser, isInternalUser, PERMISSION_ACTIONS } from "@glassact/data";
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
  user: () => GET<User> | null;
  isDealership: () => boolean;
  isInternal: () => boolean;
  can: (action: string) => boolean;
}

export const UserContext = createContext<UserState>();

export function useUserContext() {
  const context = useContext(UserContext);
  if (!context) {
    throw new Error("Can't find UserContext");
  }
  return context;
}

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

  const user = () => (query.isSuccess ? query.data : null);

  const isDealership = () => {
    const u = user();
    return u ? isDealershipUser(u) : false;
  };

  const isInternal = () => {
    const u = user();
    return u ? isInternalUser(u) : false;
  };

  const can = (action: string) => {
    const u = user();
    if (!u) return false;

    const role = u.role;

    if (isDealershipUser(u)) {
      switch (action) {
        case PERMISSION_ACTIONS.CREATE_PROJECT:
          return role === "submitter" || role === "approver" || role === "admin";
        case PERMISSION_ACTIONS.APPROVE_PROOF:
          return role === "approver" || role === "admin";
        case PERMISSION_ACTIONS.PLACE_ORDER:
          return role === "approver" || role === "admin";
        case PERMISSION_ACTIONS.PAY_INVOICE:
          return role === "admin";
        case PERMISSION_ACTIONS.MANAGE_DEALERSHIP_USERS:
          return role === "admin";
        case PERMISSION_ACTIONS.VIEW_PROJECTS:
          return true;
        case PERMISSION_ACTIONS.VIEW_INVOICES:
          return true;
        default:
          return false;
      }
    } else if (isInternalUser(u)) {
      switch (action) {
        case PERMISSION_ACTIONS.CREATE_PROOF:
          return role === "designer" || role === "admin";
        case PERMISSION_ACTIONS.MANAGE_KANBAN:
          return role === "production" || role === "admin";
        case PERMISSION_ACTIONS.CREATE_BLOCKER:
          return role === "production" || role === "admin";
        case PERMISSION_ACTIONS.CREATE_INVOICE:
          return role === "billing" || role === "admin";
        case PERMISSION_ACTIONS.MANAGE_INTERNAL_USERS:
          return role === "admin";
        case PERMISSION_ACTIONS.VIEW_ALL:
          return role === "admin";
        default:
          return false;
      }
    }

    return false;
  };

  return (
    <UserContext.Provider value={{ status, setStatus, deferredStatus, user, isDealership, isInternal, can }}>
      {props.children}
    </UserContext.Provider>
  );
};
