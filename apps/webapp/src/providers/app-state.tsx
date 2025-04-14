import { Inlay, POST } from "@glassact/data";
import { createContext, useContext, ParentComponent } from "solid-js";
import { SetStoreFunction, createStore } from "solid-js/store";

export interface AppState {
  createProject: {
    inlays: POST<Inlay>[];
  };
}

export const initialAppState: AppState = {
  createProject: {
    inlays: [
      {
        project_id: -1,
        preview_url:
          "https://tailwindui.com/plus-assets/img/ecommerce-images/checkout-page-03-product-04.jpg",
        name: "Artwork Tee",
        price_group: 1,
        type: "catalog",
        catalog_info: {
          inlay_id: -1,
          catalog_item_id: 321,
        },
      },

      {
        project_id: -1,
        preview_url:
          "https://tailwindui.com/plus-assets/img/ecommerce-images/shopping-cart-page-01-product-02.jpg",
        name: "Black Tee",
        price_group: 2,
        type: "custom",
        custom_info: {
          inlay_id: -1,
          description: "",
          width: 3,
          height: 4,
          images: [],
        },
      },
    ],
  },
};

export const AppStateContext =
  createContext<[AppState, SetStoreFunction<AppState>]>();

export const AppStateProvider: ParentComponent = (props) => {
  const [state, setState] = createStore<AppState>(initialAppState);
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
