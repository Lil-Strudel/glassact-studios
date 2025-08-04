import { createFileRoute } from "@tanstack/solid-router";

export const Route = createFileRoute("/_appLayout/dealership/$id/users")({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/dealership/$id/users"!</div>;
}
