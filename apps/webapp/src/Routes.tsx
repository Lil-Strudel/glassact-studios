import { createEffect, Match, Switch, type Component } from "solid-js";
import { Router, Route, useNavigate, RouteSectionProps } from "@solidjs/router";

import Home from "./pages/index";
import NotFound from "./pages/not-found";
import AppLayout from "./layouts/app-layout";
import Projects from "./pages/projects";
import CreateProject from "./pages/create-project";
import AddInlay from "./pages/add-inlay";
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
    const { status } = useAuthContext();
    const navigate = useNavigate();

    createEffect(() => {
      if (status() === "authenticated") {
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
    const { status } = useAuthContext();
    const navigate = useNavigate();

    createEffect(() => {
      if (status() === "unauthenticated") {
        navigate("/login", { replace: true });
      }
    });

    return (
      <Switch>
        <Match when={status() === "pending"}>Loading...</Match>
        <Match when={status() === "authenticated"}>
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
        <Route path="/projects" component={Projects} />
        <Route path="/projects/create-project" component={CreateProject} />
        <Route path="/projects/create-project/add-inlay" component={AddInlay} />
        <Route path="/organization" component={Organization} />
        <Route path="/help" component={Help} />
        <Route path="/settings" component={Settings} />
      </Route>
      <Route path="*" component={NotFound} />
    </Router>
  );
};

export default Routes;
