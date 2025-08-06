import { createFileRoute } from "@tanstack/solid-router";

export const Route = createFileRoute("/_appLayout/inlays")({
  component: RouteComponent,
});

function RouteComponent() {
  return <div></div>;
}
