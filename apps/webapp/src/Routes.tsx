import type { Component } from "solid-js";
import { Router, Route } from "@solidjs/router";

import Home from "./pages/index";
import NotFound from "./pages/not-found";
import AppLayout from "./layouts/app-layout";
import Orders from "./pages/orders";
import PlaceOrder from "./pages/place-order";
import AddItem from "./pages/add-item";

const Routes: Component = () => {
  return (
    <Router>
      <Route path="/" component={Home} />
      <Route path="/" component={AppLayout}>
        <Route path="/dashboard" component={Home} />
        <Route path="/orders" component={Orders} />
        <Route path="/orders/place-order" component={PlaceOrder} />
        <Route path="/orders/place-order/add-item" component={AddItem} />
        <Route path="/organization" component={Home} />
        <Route path="/help" component={Home} />
        <Route path="/settings" component={Home} />
      </Route>
      <Route path="/*" component={NotFound} />
    </Router>
  );
};

export default Routes;
