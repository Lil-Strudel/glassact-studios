import { createFileRoute, Outlet, redirect } from "@tanstack/solid-router";
import {
  IoPersonOutline,
  IoBusinessOutline,
  IoShapesOutline,
  IoAppsOutline,
  IoPricetagOutline,
  IoColorPaletteOutline,
  IoBrushOutline,
  IoHelpCircleOutline,
  IoImageOutline,
} from "solid-icons/io";
import { getUserSelfOpts } from "../../queries/user";
import {
  SectionNav,
  type SectionNavItem,
} from "../../components/section-nav";

export const Route = createFileRoute("/_app/admin")({
  component: RouteComponent,
  beforeLoad: async ({ context, location }) => {
    if (location.pathname.split("/").length < 3) {
      throw redirect({ to: "/admin/dealerships", replace: true });
    }

    const user = await context.queryClient.ensureQueryData(getUserSelfOpts());

    if ("dealership_id" in user) {
      throw redirect({ to: "/", replace: true });
    }
  },
});

const navigationItems: SectionNavItem[] = [
  {
    label: "Dealerships",
    icon: IoBusinessOutline,
    path: "/admin/dealerships",
  },
  {
    label: "Users",
    icon: IoPersonOutline,
    path: "/admin/users",
  },
  {
    label: "Catalog",
    icon: IoShapesOutline,
    path: "/admin/catalog",
  },
  {
    label: "Compounds",
    icon: IoAppsOutline,
    path: "/admin/compounds",
  },
  {
    label: "Backgrounds",
    icon: IoImageOutline,
    path: "/admin/backgrounds",
  },
  {
    label: "Price Groups",
    icon: IoPricetagOutline,
    path: "/admin/price-groups",
  },
  {
    label: "Glass",
    icon: IoColorPaletteOutline,
    path: "/admin/glass",
  },
  {
    label: "Grout",
    icon: IoBrushOutline,
    path: "/admin/grout",
  },
  {
    label: "Support",
    icon: IoHelpCircleOutline,
    path: "/admin/support",
  },
];

function RouteComponent() {
  return (
    <div class="lg:flex lg:gap-x-12">
      <SectionNav label="Administration" items={navigationItems} />
      <div class="mt-8 min-w-0 flex-1 space-y-8 lg:mt-0">
        <Outlet />
      </div>
    </div>
  );
}
