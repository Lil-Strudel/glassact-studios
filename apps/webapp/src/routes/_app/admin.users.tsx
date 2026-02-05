import { createFileRoute } from "@tanstack/solid-router";
import { Link } from "@tanstack/solid-router";
import { Button } from "@glassact/ui";

export const Route = createFileRoute("/_app/admin/users")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div class="space-y-6">
      <div>
        <h1 class="text-3xl font-bold">User Management</h1>
        <p class="text-gray-600 mt-2">Manage dealership and internal users</p>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Link
          to="/admin/users/dealership"
          class="block p-6 border rounded-lg hover:shadow-lg transition-shadow"
        >
          <div>
            <h2 class="text-2xl font-bold mb-2">Dealership Users</h2>
            <p class="text-gray-600 mb-4">Manage users from dealerships</p>
            <Button variant="outline">Manage Dealership Users</Button>
          </div>
        </Link>

        <Link
          to="/admin/users/internal"
          class="block p-6 border rounded-lg hover:shadow-lg transition-shadow"
        >
          <div>
            <h2 class="text-2xl font-bold mb-2">Internal Users</h2>
            <p class="text-gray-600 mb-4">Manage GlassAct Studios staff</p>
            <Button variant="outline">Manage Internal Users</Button>
          </div>
        </Link>
      </div>
    </div>
  );
}
