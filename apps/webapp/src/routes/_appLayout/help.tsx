import { createFileRoute } from "@tanstack/solid-router";

export const Route = createFileRoute("/_appLayout/help")({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/help"!</div>;
}
