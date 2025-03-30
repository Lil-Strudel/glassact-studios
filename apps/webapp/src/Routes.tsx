import { createEffect, Match, Switch, type Component } from "solid-js";
import { Router, Route, useNavigate, RouteSectionProps } from "@solidjs/router";

import Home from "./pages/index";
import NotFound from "./pages/not-found";
import AppLayout from "./layouts/app-layout";
import Orders from "./pages/orders";
import PlaceOrder from "./pages/place-order";
import AddItem from "./pages/add-item";
import Login from "./pages/login";
import Dashboard from "./pages/dashboard";
import Organization from "./pages/organization";
import Help from "./pages/help";
import Settings from "./pages/settings";
import { useAuthContext } from "./providers/auth";

const Unauthenticated = (
  Component: Component<RouteSectionProps<unknown>>,
): Component<RouteSectionProps<unknown>> => {
  return (props) => {
    const { state } = useAuthContext();
    const navigate = useNavigate();

    createEffect(() => {
      if (state.status === "authenticated") {
        navigate("/dashboard", { replace: true });
      }
    });

    return <Component {...props} />;
  };
};

const Authenticated = (
  Component: Component<RouteSectionProps<unknown>>,
): Component<RouteSectionProps<unknown>> => {
  return (props) => {
    const { state } = useAuthContext();
    const navigate = useNavigate();

    createEffect(() => {
      if (state.status === "unauthenticated") {
        navigate("/login", { replace: true });
      }
    });

    return (
      <Switch>
        <Match when={state.status === "pending"}>Loading...</Match>
        <Match when={state.status === "authenticated"}>
          <Component {...props} />
        </Match>
      </Switch>
    );
  };
};

const Routes: Component = () => {
  return (
    <Router>
      <Route path="/" component={Unauthenticated(Home)} />
      <Route path="/login" component={Unauthenticated(Login)} />
      <Route path="/" component={Authenticated(AppLayout)}>
        <Route path="/dashboard" component={Dashboard} />
        <Route path="/orders" component={Orders} />
        <Route path="/orders/place-order" component={PlaceOrder} />
        <Route path="/orders/place-order/add-item" component={AddItem} />
        <Route path="/organization" component={Organization} />
        <Route path="/help" component={Help} />
        <Route path="/settings" component={Settings} />
      </Route>
      <Route path="*" component={NotFound} />
    </Router>
  );
};

export default Routes;
