import { createFileRoute, Link, redirect } from "@tanstack/solid-router";
import { Button } from "@glassact/ui";
import logoMark from "../assets/images/logo-mark.svg";
import logoReversed from "../assets/images/logo-reversed.png";

export const Route = createFileRoute("/")({
  component: RouteComponent,
  beforeLoad: async ({ context }) => {
    const status = await context.auth.deferredStatus().promise;

    if (status === "authenticated") {
      throw redirect({
        to: "/dashboard",
        replace: true,
      });
    }
  },
});

function RouteComponent() {
  return (
    <div class="relative flex min-h-dvh flex-col overflow-hidden">
      <div
        aria-hidden="true"
        class="pointer-events-none absolute inset-x-0 top-0 h-[70vh] bg-gradient-to-b from-crimson-deep/35 via-crimson-deep/10 to-transparent"
      />
      <nav class="relative mx-auto flex h-16 w-full max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <img
          class="block h-10 w-auto"
          src={logoMark}
          alt="GlassAct Studios"
        />
        <Button as={Link} to="/login" variant="ghost" class="gap-2">
          Sign in <span aria-hidden="true">&rarr;</span>
        </Button>
      </nav>
      <main class="relative flex flex-1 items-center justify-center px-6 py-24">
        <div class="flex max-w-2xl flex-col items-center gap-8 text-center">
          <img
            src={logoReversed}
            alt=""
            aria-hidden="true"
            class="h-20 w-auto sm:h-24"
          />
          <div class="ornament-rule w-56">
            <span class="lozenge" />
          </div>
          <h1 class="text-5xl leading-[1.05] sm:text-6xl">GlassAct Studios</h1>
          <p class="max-w-prose text-pretty text-lg leading-relaxed text-muted-foreground sm:text-xl">
            Place and track your stained glass inlay orders.
          </p>
          <Button as={Link} to="/login" size="lg">
            Sign in
          </Button>
        </div>
      </main>
    </div>
  );
}
