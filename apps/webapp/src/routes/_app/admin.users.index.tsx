import { createFileRoute, Link } from "@tanstack/solid-router";
import { IoArrowForward } from "solid-icons/io";
import { PageHeader } from "@glassact/ui";

export const Route = createFileRoute("/_app/admin/users/")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div class="space-y-8">
      <PageHeader
        eyebrow="Administration"
        title="User Management"
        description="Manage dealership and internal users"
      />

      <div class="grid grid-cols-1 gap-6 md:grid-cols-2">
        <Link
          to="/admin/users/dealership"
          class="group flex flex-col gap-2 rounded-lg border border-border bg-card p-6 transition-colors hover:border-gilt/45 hover:bg-accent/40"
        >
          <h2 class="text-2xl">Dealership Users</h2>
          <p class="text-muted-foreground">Manage users from dealerships</p>
          <span class="mt-3 inline-flex items-center gap-2 text-sm font-medium text-crimson-light">
            Manage Dealership Users
            <IoArrowForward class="transition-transform group-hover:translate-x-0.5" />
          </span>
        </Link>

        <Link
          to="/admin/users/internal"
          class="group flex flex-col gap-2 rounded-lg border border-border bg-card p-6 transition-colors hover:border-gilt/45 hover:bg-accent/40"
        >
          <h2 class="text-2xl">Internal Users</h2>
          <p class="text-muted-foreground">Manage GlassAct Studios staff</p>
          <span class="mt-3 inline-flex items-center gap-2 text-sm font-medium text-crimson-light">
            Manage Internal Users
            <IoArrowForward class="transition-transform group-hover:translate-x-0.5" />
          </span>
        </Link>
      </div>
    </div>
  );
}
