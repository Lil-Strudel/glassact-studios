import { createFileRoute } from "@tanstack/solid-router";

export const Route = createFileRoute("/_appLayout/admin/users")({
  component: RouteComponent,
});

function RouteComponent() {
  return <div>Hello "/admin/users"!</div>;
}
