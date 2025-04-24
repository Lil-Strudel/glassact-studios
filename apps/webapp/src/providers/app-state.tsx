import { Inlay, POST } from "@glassact/data";
import { createContext, useContext, ParentComponent } from "solid-js";
import { SetStoreFunction, createStore } from "solid-js/store";
import { deepCopy } from "../utils/deep-copy";

interface CreateProject {
  name: string;
  inlays: POST<Inlay>[];
}
const initialCreateProject: CreateProject = {
  name: "",
  inlays: [],
};

export interface AppState {
  createProject: CreateProject;
}
export const initialAppState: AppState = {
  createProject: initialCreateProject,
};

export const AppStateContext =
  createContext<[AppState, SetStoreFunction<AppState>]>();

export const AppStateProvider: ParentComponent = (props) => {
  const appState = deepCopy(initialAppState);
  const [state, setState] = createStore<AppState>(appState);
  return (
    <AppStateContext.Provider value={[state, setState]}>
      {props.children}
    </AppStateContext.Provider>
  );
};

export function useAppState() {
  const context = useContext(AppStateContext);
  if (!context) {
    throw new Error("Can't find AppStateContext");
  }
  return context;
}
