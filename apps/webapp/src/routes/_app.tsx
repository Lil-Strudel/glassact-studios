import {
  createFileRoute,
  Outlet,
  Link,
  redirect,
} from "@tanstack/solid-router";
import { createMemo, createSignal, For, Show } from "solid-js";
import logoFull from "../assets/images/logo-full.svg";
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
import { CustomerModeToggle } from "../components/customer-mode-toggle";
import { useCustomerMode } from "../hooks/use-customer-mode";
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
  {
    name: "Review",
    to: "/review-queue",
    permission: PERMISSION_ACTIONS.INTERNAL_APPROVE_PROOF,
  },
  {
    name: "Inlays",
    to: "/inlays",
    permission: PERMISSION_ACTIONS.MANAGE_KANBAN,
  },
  {
    name: "My Dealership",
    to: "/dealership",
    permission: PERMISSION_ACTIONS.MANAGE_DEALERSHIP_USERS,
    dealershipOnly: true,
  },
  { name: "Support", to: "/support" },
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
  const { user, can, isDealership } = useUserContext();
  const [open, setOpen] = createSignal(false);
  const { isCustomerMode } = useCustomerMode();

  const filteredNavigation = createMemo(() =>
    navigation.filter(
      (item) =>
        (!item.permission || can(item.permission)) &&
        (!item.dealershipOnly || isDealership()),
    ),
  );

  function toggleOpen() {
    setOpen((open) => !open);
  }

  const mobileLinkClass =
    "block border-l-2 py-2 pl-4 pr-4 text-base font-medium transition-colors";
  const mobileActiveClass = "border-crimson-light bg-accent text-foreground";
  const mobileInactiveClass =
    "border-transparent text-muted-foreground hover:bg-accent/60 hover:text-foreground";

  return (
    <div
      class="flex h-dvh flex-col bg-background"
      style={{ "--app-pane-height": "calc(100dvh - 4rem - 1px - 4rem)" }}
    >
      <nav class="relative z-40 shrink-0 border-b border-border bg-card">
        <div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div class="flex h-16 justify-between">
            <div class="flex">
              <Link to="/dashboard" class="flex shrink-0 items-center">
                <img
                  class="block h-9 w-auto"
                  src={logoFull}
                  alt="GlassAct Studios"
                />
              </Link>
              <div class="hidden sm:-my-px sm:ml-10 sm:flex sm:gap-7">
                <For each={filteredNavigation()}>
                  {(item) => (
                    <Link
                      to={item.to}
                      class="inline-flex items-center border-b-2 px-0.5 pt-0.5 text-sm font-medium transition-colors"
                      activeProps={{
                        class: "border-crimson-light text-foreground",
                      }}
                      inactiveProps={{
                        class:
                          "border-transparent text-muted-foreground hover:border-input hover:text-foreground",
                      }}
                    >
                      {item.name}
                    </Link>
                  )}
                </For>
              </div>
            </div>
            <div class="hidden sm:ml-6 sm:flex sm:items-center sm:gap-2">
              <Show when={isCustomerMode()}>
                <CustomerModeToggle size="sm" />
              </Show>
              <NotificationBell />

              <Show when={user()}>
                {(currentUser) => (
                  <DropdownMenu placement="bottom-end">
                    <DropdownMenuTrigger class="rounded-full">
                      <img
                        class="size-8 rounded-full border border-input object-cover transition-colors hover:border-gilt/70"
                        src={currentUser().avatar}
                        alt="Avatar"
                      />
                    </DropdownMenuTrigger>
                    <DropdownMenuContent class="min-w-48">
                      <div class="border-b border-border px-2 pb-2 pt-1">
                        <p class="truncate text-sm font-medium">
                          {currentUser().name}
                        </p>
                        <p class="truncate text-xs text-faint">
                          {currentUser().email}
                        </p>
                      </div>
                      <div class="pt-1">
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
                      </div>
                    </DropdownMenuContent>
                  </DropdownMenu>
                )}
              </Show>
            </div>
            <div class="-mr-2 flex items-center gap-2 sm:hidden">
              <Show when={isCustomerMode()}>
                <CustomerModeToggle size="sm" />
              </Show>
              <Button
                size="icon"
                variant="ghost"
                onClick={toggleOpen}
                aria-label={open() ? "Close menu" : "Open menu"}
              >
                {open() ? <IoClose size={24} /> : <IoMenu size={24} />}
              </Button>
            </div>
          </div>
        </div>

        {open() && (
          <div class="absolute w-full border-b border-border bg-popover shadow-2xl shadow-black/50 sm:hidden">
            <div class="space-y-1 pb-3 pt-2">
              <For each={filteredNavigation()}>
                {(item) => (
                  <Link
                    to={item.to}
                    class={mobileLinkClass}
                    activeProps={{ class: mobileActiveClass }}
                    inactiveProps={{ class: mobileInactiveClass }}
                  >
                    {item.name}
                  </Link>
                )}
              </For>
            </div>
            <div class="border-t border-border pb-3 pt-4">
              <Show when={user()}>
                {(currentUser) => (
                  <>
                    <div class="flex items-center px-4">
                      <div class="shrink-0">
                        <img
                          class="size-10 rounded-full border border-input object-cover"
                          src={currentUser().avatar}
                          alt="Avatar"
                        />
                      </div>
                      <div class="ml-3 min-w-0">
                        <div class="truncate text-base font-medium text-foreground">
                          {currentUser().name}
                        </div>
                        <div class="truncate text-sm text-faint">
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
                              class={`${mobileLinkClass} ${mobileInactiveClass}`}
                            >
                              {item.name}
                            </a>
                          ) : (
                            <item.component
                              {...item.props}
                              class={mobileLinkClass}
                              activeProps={{ class: mobileActiveClass }}
                              inactiveProps={{ class: mobileInactiveClass }}
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
      <main class="min-h-0 flex-1 overflow-y-auto">
        <div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
          <Outlet />
        </div>
      </main>
    </div>
  );
}
