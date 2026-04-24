import {
  createFileRoute,
  Outlet,
  Link,
  redirect,
} from "@tanstack/solid-router";
import { createMemo, createSignal, For, Show } from "solid-js";
import logoEmblem from "../assets/images/logo-emblem.png";
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@glassact/ui";
import { IoClose, IoMenu } from "solid-icons/io";
import { useUserContext } from "../providers/user";
import { NotificationBell } from "../components/notification-bell";
import { PERMISSION_ACTIONS } from "@glassact/data";

export const Route = createFileRoute("/_app")({
  component: RouteComponent,
  beforeLoad: async ({ context, location }) => {
    const status = await context.auth.deferredStatus().promise;

    if (status === "unauthenticated") {
      throw redirect({
        to: "/login",
        replace: true,
        search: {
          redirect: location.href,
        },
      });
    }
  },
});

const navigation = [
  { name: "Dashboard", to: "/dashboard" },
  { name: "Projects", to: "/projects" },
  { name: "Catalog", to: "/catalog" },
  { name: "Inlays", to: "/inlays" },
  { name: "Admin", to: "/admin", permission: PERMISSION_ACTIONS.ACCESS_ADMIN },
];
const userNavigation = [
  { component: Link, name: "Settings", props: { to: "/settings" } },
  {
    component: "a",
    name: "Logout",
    props: { href: "/api/auth/logout", rel: "external" },
  },
];

function RouteComponent() {
  const { user, can } = useUserContext();
  const [open, setOpen] = createSignal(false);

  const filteredNavigation = createMemo(() =>
    navigation.filter((item) => !item.permission || can(item.permission)),
  );

  function toggleOpen() {
    setOpen((open) => !open);
  }

  return (
    <div>
      <div class="min-h-full">
        <nav class="border-b border-gray-200">
          <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div class="flex h-16 justify-between">
              <div class="flex">
                <div class="flex shrink-0 items-center">
                  <img
                    class="block h-12 w-auto"
                    src={logoEmblem}
                    alt="GlassAct Studios"
                  />
                </div>
                <div class="hidden sm:-my-px sm:ml-6 sm:flex sm:space-x-8">
                  <For each={filteredNavigation()}>
                    {(item) => (
                      <Link
                        to={item.to}
                        class="inline-flex items-center border-b-2 px-1 pt-1 text-sm font-medium"
                        activeProps={{ class: "border-primary" }}
                        inactiveProps={{
                          class:
                            "border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700",
                        }}
                      >
                        {item.name}
                      </Link>
                    )}
                  </For>
                </div>
              </div>
              <div class="hidden sm:ml-6 sm:flex sm:items-center">
                <NotificationBell />

                <Show when={user()}>
                  {(currentUser) => (
                    <DropdownMenu placement="bottom-end">
                      <DropdownMenuTrigger class="ml-3">
                        <img
                          class="size-8 rounded-full"
                          src={currentUser().avatar}
                          alt="Avatar"
                        />
                      </DropdownMenuTrigger>
                      <DropdownMenuContent>
                        <For each={userNavigation}>
                          {(item) => (
                            <DropdownMenuItem
                              as={item.component}
                              {...item.props}
                            >
                              {item.name}
                            </DropdownMenuItem>
                          )}
                        </For>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  )}
                </Show>
              </div>
              <div class="-mr-2 flex items-center sm:hidden">
                <Button size="icon" variant="ghost" onClick={toggleOpen}>
                  {open() ? <IoClose size={24} /> : <IoMenu size={24} />}
                </Button>
              </div>
            </div>
          </div>

          {open() && (
            <div class="bg-white w-full drop-shadow absolute sm:hidden">
              <div class="space-y-1 pb-3 pt-2">
                <For each={filteredNavigation()}>
                  {(item) => (
                    <Link
                      to={item.to}
                      class="block border-l-4 py-2 pl-3 pr-4 text-base font-medium text-gray-600"
                      activeProps={{ class: "border-primary bg-red-50" }}
                      inactiveProps={{
                        class:
                          "border-transparent hover:border-gray-300 hover:bg-gray-50 hover:text-gray-800",
                      }}
                    >
                      {item.name}
                    </Link>
                  )}
                </For>
              </div>
              <div class="border-t border-gray-200 pb-3 pt-4">
                <Show when={user()}>
                  {(currentUser) => (
                    <>
                      <div class="flex items-center px-4">
                        <div class="shrink-0">
                          <img
                            class="size-10 rounded-full"
                            src={currentUser().avatar}
                            alt="Avatar"
                          />
                        </div>
                        <div class="ml-3">
                          <div class="text-base font-medium text-gray-800">
                            {currentUser().name}
                          </div>
                          <div class="text-sm font-medium text-gray-500">
                            {currentUser().email}
                          </div>
                        </div>
                        <div class="ml-auto">
                          <NotificationBell />
                        </div>
                      </div>
                      <div class="mt-3 space-y-1">
                        <For each={userNavigation}>
                          {(item) =>
                            item.component === "a" ? (
                              <a
                                {...item.props}
                                class="block border-l-4 py-2 pl-3 pr-4 text-base font-medium text-gray-600"
                              >
                                {item.name}
                              </a>
                            ) : (
                              <item.component
                                {...item.props}
                                class="block border-l-4 py-2 pl-3 pr-4 text-base font-medium text-gray-600"
                                activeProps={{
                                  class: "border-primary bg-red-50",
                                }}
                                inactiveProps={{
                                  class:
                                    "border-transparent hover:border-gray-300 hover:bg-gray-50 hover:text-gray-800",
                                }}
                              >
                                {item.name}
                              </item.component>
                            )
                          }
                        </For>
                      </div>
                    </>
                  )}
                </Show>
              </div>
            </div>
          )}
        </nav>
        <main>
          <div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
}
