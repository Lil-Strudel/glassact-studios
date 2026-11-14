import { createFileRoute, Outlet, redirect } from "@tanstack/solid-router";
import { IoPeopleOutline, IoSettingsOutline } from "solid-icons/io";
import { useQuery } from "@tanstack/solid-query";
import { getDealershipOpts } from "../../queries/dealership";
import {
  SectionNav,
  type SectionNavItem,
} from "../../components/section-nav";
import { PageHeader } from "@glassact/ui";

export const Route = createFileRoute("/_app/dealership/$id")({
  component: RouteComponent,
  beforeLoad({ params, location }) {
    if (location.pathname.split("/").length < 4) {
      throw redirect({
        to: "/dealership/$id/users",
        replace: true,
        params,
      });
    }
  },
});

const navigationItems: SectionNavItem[] = [
  {
    label: "Users",
    icon: IoPeopleOutline,
    path: "/dealership/$id/users",
  },
  {
    label: "Settings",
    icon: IoSettingsOutline,
    path: "/dealership/$id/settings",
  },
];

function RouteComponent() {
  const params = Route.useParams();
  const query = useQuery(() => getDealershipOpts(params().id));

  return (
    <div class="lg:flex lg:gap-x-12">
      <SectionNav label="Dealership" items={navigationItems} />
      <div class="mt-8 min-w-0 flex-1 space-y-8 lg:mt-0">
        <PageHeader
          eyebrow="Dealership"
          title={query.data?.name || "Loading..."}
        />
        <Outlet />
      </div>
    </div>
  );
}
