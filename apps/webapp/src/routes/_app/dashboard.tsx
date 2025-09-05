import { createFileRoute } from "@tanstack/solid-router";

export const Route = createFileRoute("/_app/dashboard")({
  component: RouteComponent,
});

function RouteComponent() {
  return <div></div>;
}
