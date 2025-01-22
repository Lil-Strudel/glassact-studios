import type { Component } from "solid-js";
import { Router, Route } from "@solidjs/router";

import Home from "./pages/index";
import NotFound from "./pages/not-found";

const Routes: Component = () => {
  return (
    <Router>
      <Route path="/" component={Home} />
      <Route path="/*" component={NotFound} />
    </Router>
  );
};

export default Routes;
