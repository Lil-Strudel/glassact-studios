import { createFileRoute } from "@tanstack/solid-router";

export const Route = createFileRoute("/_appLayout/dashboard")({
  component: RouteComponent,
});

function RouteComponent() {
  return <div></div>;
}
