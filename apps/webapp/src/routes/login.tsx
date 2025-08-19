import { createFileRoute, useNavigate } from "@tanstack/solid-router";
import { useAuthContext } from "../providers/auth";
import { createEffect } from "solid-js";
import { Button, Form } from "@glassact/ui";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { useMutation } from "@tanstack/solid-query";
import { postAuthMagicLinkOpts } from "../queries/auth";

export const Route = createFileRoute("/login")({
  component: RouteComponent,
});

function RouteComponent() {
  const auth = useAuthContext();
  const navigate = useNavigate();

  const postAuthMagicLink = useMutation(postAuthMagicLinkOpts);

  const form = createForm(() => ({
    defaultValues: {
      email: "",
    },
    validators: {
      onSubmit: z.object({
        email: z.string(),
      }),
    },
    onSubmit: async ({ value }) => {
      postAuthMagicLink.mutate(value);
    },
  }));

  createEffect(() => {
    if (auth.status() === "authenticated") {
      navigate({ to: "/dashboard", replace: true });
    }
  });

  return (
    <div class="flex min-h-full flex-col justify-center py-12 sm:px-6 lg:px-8">
      <div class="sm:mx-auto sm:w-full sm:max-w-md">
        <img
          src="/src/assets/images/logo-emblem.png"
          alt="GlassAct Studios"
          class="mx-auto h-14 w-auto"
        />
        <h2 class="mt-6 text-center text-2xl/9 font-bold tracking-tight text-gray-900">
          Sign in to your account
        </h2>
      </div>

      <div class="mt-10 sm:mx-auto sm:w-full sm:max-w-[480px]">
        <div class="bg-white px-6 py-12 shadow sm:rounded-lg sm:px-12">
          <form
            onSubmit={(e) => {
              e.preventDefault();
              e.stopPropagation();
              form.handleSubmit();
            }}
            class="flex flex-col gap-4"
          >
            <form.Field
              name="email"
              children={(field) => (
                <Form.TextField field={field} label="Email" />
              )}
            />

            <Button type="submit" class="w-full">
              Send Magic Link
            </Button>
          </form>

          <div>
            <div class="mt-10 flex items-center gap-x-6">
              <div class="w-full flex-1 border-t border-gray-200"></div>
              <p class="text-nowrap text-sm/6 font-medium text-gray-900">
                Or continue with
              </p>
              <div class="w-full flex-1 border-t border-gray-200"></div>
            </div>

            <div class="mt-6 grid grid-cols-2 gap-4">
              <a
                href="/api/auth/google"
                rel="external"
                class="flex w-full items-center justify-center gap-3 rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus-visible:ring-transparent"
              >
                <svg viewBox="0 0 24 24" aria-hidden="true" class="h-5 w-5">
                  <path
                    d="M12.0003 4.75C13.7703 4.75 15.3553 5.36002 16.6053 6.54998L20.0303 3.125C17.9502 1.19 15.2353 0 12.0003 0C7.31028 0 3.25527 2.69 1.28027 6.60998L5.27028 9.70498C6.21525 6.86002 8.87028 4.75 12.0003 4.75Z"
                    fill="#EA4335"
                  />
                  <path
                    d="M23.49 12.275C23.49 11.49 23.415 10.73 23.3 10H12V14.51H18.47C18.18 15.99 17.34 17.25 16.08 18.1L19.945 21.1C22.2 19.01 23.49 15.92 23.49 12.275Z"
                    fill="#4285F4"
                  />
                  <path
                    d="M5.26498 14.2949C5.02498 13.5699 4.88501 12.7999 4.88501 11.9999C4.88501 11.1999 5.01998 10.4299 5.26498 9.7049L1.275 6.60986C0.46 8.22986 0 10.0599 0 11.9999C0 13.9399 0.46 15.7699 1.28 17.3899L5.26498 14.2949Z"
                    fill="#FBBC05"
                  />
                  <path
                    d="M12.0004 24.0001C15.2404 24.0001 17.9654 22.935 19.9454 21.095L16.0804 18.095C15.0054 18.82 13.6204 19.245 12.0004 19.245C8.8704 19.245 6.21537 17.135 5.2654 14.29L1.27539 17.385C3.25539 21.31 7.3104 24.0001 12.0004 24.0001Z"
                    fill="#34A853"
                  />
                </svg>
                <span class="text-sm/6 font-semibold">Google</span>
              </a>

              <a
                href="/api/auth/microsoft"
                rel="external"
                class="flex w-full items-center justify-center gap-3 rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50 focus-visible:ring-transparent"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 23 23"
                  class="size-5"
                >
                  <path fill="#f3f3f3" d="M0 0h23v23H0z" />
                  <path fill="#f35325" d="M1 1h10v10H1z" />
                  <path fill="#81bc06" d="M12 1h10v10H12z" />
                  <path fill="#05a6f0" d="M1 12h10v10H1z" />
                  <path fill="#ffba08" d="M12 12h10v10H12z" />
                </svg>
                <span class="text-sm/6 font-semibold">Microsoft</span>
              </a>
            </div>
          </div>
        </div>

        <p class="mt-10 text-center text-sm/6 text-gray-500">
          This app is currently a closed beta.
        </p>
      </div>
    </div>
  );
}
