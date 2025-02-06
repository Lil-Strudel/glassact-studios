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

const user = {
  name: "Tom Cook",
  email: "tom@example.com",
  imageUrl:
    "https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?ixlib=rb-1.2.1&ixid=eyJhcHBfaWQiOjEyMDd9&auto=format&fit=facearea&facepad=2&w=256&h=256&q=80",
};
const navigation = [
  { name: "Dashboard", href: "/dashboard" },
  { name: "Orders", href: "/orders" },
  { name: "My Org", href: "/organization" },
  { name: "Help", href: "/help" },
];
const userNavigation = [
  { name: "Settings", href: "/settings", props: {} },
  { name: "Sign out", href: "/api/auth/sign-out", props: { rel: "external" } },
];

const AppLayout: Component<RouteSectionProps<unknown>> = (props) => {
  const [open, setOpen] = createSignal(false);

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
                  <svg
                    class="size-6"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke-width="1.5"
                    stroke="currentColor"
                    aria-hidden="true"
                    data-slot="icon"
                  >
                    <path
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      d="M14.857 17.082a23.848 23.848 0 0 0 5.454-1.31A8.967 8.967 0 0 1 18 9.75V9A6 6 0 0 0 6 9v.75a8.967 8.967 0 0 1-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 0 1-5.714 0m5.714 0a3 3 0 1 1-5.714 0"
                    />
                  </svg>
                </Button>

                <DropdownMenu placement="bottom-end">
                  <DropdownMenuTrigger class="ml-3">
                    <img
                      class="size-8 rounded-full"
                      src={user.imageUrl}
                      alt="Profile"
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
                  {open() ? (
                    <svg
                      class="size-6"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke-width="1.5"
                      stroke="currentColor"
                      data-slot="icon"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M6 18 18 6M6 6l12 12"
                      />
                    </svg>
                  ) : (
                    <svg
                      class="size-6"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke-width="1.5"
                      stroke="currentColor"
                      data-slot="icon"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5"
                      />
                    </svg>
                  )}
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
                      src={user.imageUrl}
                      alt=""
                    />
                  </div>
                  <div class="ml-3">
                    <div class="text-base font-medium text-gray-800">
                      {user.name}
                    </div>
                    <div class="text-sm font-medium text-gray-500">
                      {user.email}
                    </div>
                  </div>
                  <Button size="icon" variant="ghost" class="ml-auto">
                    <svg
                      class="size-6"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke-width="1.5"
                      stroke="currentColor"
                      aria-hidden="true"
                      data-slot="icon"
                    >
                      <path
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        d="M14.857 17.082a23.848 23.848 0 0 0 5.454-1.31A8.967 8.967 0 0 1 18 9.75V9A6 6 0 0 0 6 9v.75a8.967 8.967 0 0 1-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 0 1-5.714 0m5.714 0a3 3 0 1 1-5.714 0"
                      />
                    </svg>
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
