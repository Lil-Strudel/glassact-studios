import type { Component } from "solid-js";
import { createSignal } from "solid-js";
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@glassact/ui";
import { A, RouteSectionProps } from "@solidjs/router";
import { IoClose, IoMenu, IoNotificationsOutline } from "solid-icons/io";
import { createQuery } from "@tanstack/solid-query";
import { getUserSelfOpts } from "../queries/user";

const navigation = [
  { name: "Dashboard", href: "/dashboard" },
  { name: "Orders", href: "/orders" },
  { name: "My Org", href: "/organization" },
  { name: "Help", href: "/help" },
];
const userNavigation = [
  { name: "Settings", href: "/settings", props: {} },
  { name: "Logout", href: "/api/auth/logout", props: { rel: "external" } },
];

const AppLayout: Component<RouteSectionProps<unknown>> = (props) => {
  const query = createQuery(getUserSelfOpts);

  const [open, setOpen] = createSignal(false);

  function toggleOpen() {
    setOpen((open) => !open);
  }

  const user = () => ({
    name: query.data?.name || "Unnamed User",
    email: query.data?.email || "placeholder@email.com",
    imageUrl: query.data?.avatar || "https://placehold.co/400",
  });

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
                    src="/src/assets/images/logo-emblem.png"
                    alt="GlassAct Studios"
                  />
                </div>
                <div class="hidden sm:-my-px sm:ml-6 sm:flex sm:space-x-8">
                  {navigation.map((item) => (
                    <A
                      href={item.href}
                      activeClass="border-primary"
                      inactiveClass="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
                      class="inline-flex items-center border-b-2 px-1 pt-1 text-sm font-medium"
                    >
                      {item.name}
                    </A>
                  ))}
                </div>
              </div>
              <div class="hidden sm:ml-6 sm:flex sm:items-center">
                <Button size="icon" variant="ghost">
                  <IoNotificationsOutline size={24} />
                </Button>

                <DropdownMenu placement="bottom-end">
                  <DropdownMenuTrigger class="ml-3">
                    <img
                      class="size-8 rounded-full"
                      src={user().imageUrl}
                      alt="Avatar"
                    />
                  </DropdownMenuTrigger>
                  <DropdownMenuContent>
                    {userNavigation.map((item) => (
                      <DropdownMenuItem as="a" href={item.href} {...item.props}>
                        {item.name}
                      </DropdownMenuItem>
                    ))}
                  </DropdownMenuContent>
                </DropdownMenu>
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
                {navigation.map((item) => (
                  <A
                    href={item.href}
                    activeClass="border-primary bg-red-50"
                    inactiveClass="border-transparent hover:border-gray-300 hover:bg-gray-50 hover:text-gray-800"
                    class="block border-l-4 py-2 pl-3 pr-4 text-base font-medium text-gray-600"
                  >
                    {item.name}
                  </A>
                ))}
              </div>
              <div class="border-t border-gray-200 pb-3 pt-4">
                <div class="flex items-center px-4">
                  <div class="shrink-0">
                    <img
                      class="size-10 rounded-full"
                      src={user().imageUrl}
                      alt="Avatar"
                    />
                  </div>
                  <div class="ml-3">
                    <div class="text-base font-medium text-gray-800">
                      {user().name}
                    </div>
                    <div class="text-sm font-medium text-gray-500">
                      {user().email}
                    </div>
                  </div>
                  <Button size="icon" variant="ghost" class="ml-auto">
                    <IoNotificationsOutline size={24} />
                  </Button>
                </div>
                <div class="mt-3 space-y-1">
                  {userNavigation.map((item) => (
                    <A
                      href={item.href}
                      activeClass="border-primary bg-red-50"
                      inactiveClass="border-transparent hover:border-gray-300 hover:bg-gray-50 hover:text-gray-800"
                      class="block border-l-4 py-2 pl-3 pr-4 text-base font-medium text-gray-600"
                    >
                      {item.name}
                    </A>
                  ))}
                </div>
              </div>
            </div>
          )}
        </nav>
        <main>
          <div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
            {props.children}
          </div>
        </main>
      </div>
    </div>
  );
};

export default AppLayout;
