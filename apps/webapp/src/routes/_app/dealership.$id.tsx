import {
  createFileRoute,
  Link,
  Outlet,
  redirect,
  useParams,
} from "@tanstack/solid-router";
import { For } from "solid-js";
import { IoPersonOutline, IoBusinessOutline } from "solid-icons/io";
import { useQuery } from "@tanstack/solid-query";
import { getDealershipOpts } from "../../queries/dealership";

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

const navigationItems = [
  {
    label: "Users",
    icon: IoBusinessOutline,
    path: "/dealership/$id/users",
  },
  {
    label: "Settings",
    icon: IoPersonOutline,
    path: "/dealership/$id/settings",
  },
];
function RouteComponent() {
  const params = Route.useParams();
  const query = useQuery(getDealershipOpts(params().id));

  return (
    <div>
      <div class="pt-16 lg:flex lg:gap-x-16">
        <aside class="flex overflow-x-auto border-b border-gray-900/5 py-4 lg:block lg:w-64 lg:flex-none lg:border-0 lg:py-20">
          <nav class="flex-none px-4 sm:px-6 lg:px-0">
            <ul
              role="list"
              class="flex gap-x-3 gap-y-1 whitespace-nowrap lg:flex-col"
            >
              <For each={navigationItems}>
                {(item) => {
                  return (
                    <li>
                      <Link
                        to={item.path}
                        class="group flex gap-x-3 rounded-md py-2 pl-2 pr-3 text-sm/6 font-semibold"
                        activeProps={{ class: "bg-gray-50 text-primary" }}
                        inactiveProps={{
                          class:
                            "text-gray-700 hover:bg-gray-50 hover:text-primary",
                        }}
                      >
                        <item.icon size={24} />
                        {item.label}
                      </Link>
                    </li>
                  );
                }}
              </For>
            </ul>
          </nav>
        </aside>

        <div class="mx-auto max-w-2xl w-full space-y-16 sm:space-y-20 lg:mx-0 lg:max-w-none">
          <div class="border-b border-gray-200">
            <h1 class="text-3xl font-bold text-gray-900">
              {query.data?.name || "Loading..."}
            </h1>
          </div>
          <Outlet />
        </div>
      </div>
    </div>
  );
}
