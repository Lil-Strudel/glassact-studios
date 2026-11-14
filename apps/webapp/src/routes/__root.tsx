import {
  ErrorComponentProps,
  Outlet,
  createRootRouteWithContext,
} from "@tanstack/solid-router";
import { TanStackDevtools } from "@tanstack/solid-devtools";
import { SolidQueryDevtoolsPanel } from "@tanstack/solid-query-devtools";
import { TanStackRouterDevtoolsPanel } from "@tanstack/solid-router-devtools";
import { formDevtoolsPlugin } from "@tanstack/solid-form-devtools";
import { RouterContext } from "../App";
import { Toaster } from "@glassact/ui";
import type { JSX } from "solid-js";

export const Route = createRootRouteWithContext<RouterContext>()({
  component: RootComponent,
  notFoundComponent: NotFoundComponent,
  errorComponent: ErrorComponent,
});

function RootComponent() {
  return (
    <>
      <Outlet />
      <Toaster />
      <TanStackDevtools
        plugins={[
          {
            name: "TanStack Query",
            render: () => <SolidQueryDevtoolsPanel />,
          },
          {
            name: "TanStack Router",
            render: () => <TanStackRouterDevtoolsPanel />,
          },
          formDevtoolsPlugin(),
        ]}
      />
    </>
  );
}

function NotFoundComponent() {
  return (
    <StatusPage code="404" title="Page not found">
      Sorry, we couldn’t find the page you’re looking for.
    </StatusPage>
  );
}

function ErrorComponent(props: ErrorComponentProps) {
  console.log(props);
  return (
    <StatusPage code="500" title="Something went wrong">
      {props.error.message}
    </StatusPage>
  );
}

function StatusPage(props: { code: string; title: string; children: JSX.Element }) {
  return (
    <main class="grid min-h-full place-items-center bg-background px-6 py-24 sm:py-32 lg:px-8">
      <div class="flex max-w-xl flex-col items-center gap-6 text-center">
        <p class="eyebrow text-gilt">{props.code}</p>
        <h1 class="text-5xl sm:text-6xl">{props.title}</h1>
        <div class="ornament-rule w-48">
          <span class="lozenge" />
        </div>
        <p class="text-pretty text-lg leading-relaxed text-muted-foreground">
          {props.children}
        </p>
      </div>
    </main>
  );
}
