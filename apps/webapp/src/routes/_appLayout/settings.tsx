import { createFileRoute } from "@tanstack/solid-router";

export const Route = createFileRoute("/_appLayout/settings")({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/settings"!</div>;
}
