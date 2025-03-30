import { createEffect, type Component } from "solid-js";
import { useAuthContext } from "../providers/auth";
import { useNavigate } from "@solidjs/router";
import { Button } from "@glassact/ui";

const Home: Component = () => {
  const { state } = useAuthContext();
  const navigate = useNavigate();

  createEffect(() => {
    if (state.status === "authenticated") {
      navigate("/dashboard", { replace: true });
    }
  });

  return (
    <div>
      <div class="min-h-full">
        <nav>
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
              </div>
              <div class="hidden sm:ml-6 sm:flex sm:items-center">
                <Button as="a" href="/login" variant="ghost" class="gap-2">
                  Login <span aria-hidden="true">&rarr;</span>
                </Button>
              </div>
            </div>
          </div>
        </nav>
        <main>
          <div class="px-6 pt-14 lg:px-8">
            <div class="mx-auto max-w-2xl py-32 sm:py-48 lg:py-56">
              <div class="text-center">
                <h1 class="text-balance text-5xl font-semibold tracking-tight text-gray-900 sm:text-7xl">
                  GlassAct Studios
                </h1>
                <p class="mt-8 text-pretty text-lg font-medium text-gray-500 sm:text-xl/8">
                  Track and place new orders using our new platform!
                </p>
                <div class="mt-10 flex items-center justify-center gap-x-6">
                  <Button as="a" href="/login">
                    Login
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
};

export default Home;
