import { createFileRoute, useNavigate } from "@tanstack/solid-router";
import { Button, TextField, TextFieldRoot } from "@glassact/ui";
import { useAuthContext } from "../providers/auth";
import { createEffect } from "solid-js";

export const Route = createFileRoute("/login")({
  component: RouteComponent,
});

function RouteComponent() {
  const auth = useAuthContext();
  const navigate = useNavigate();

  createEffect(() => {
    if (auth.status() === "authenticated") {
      navigate({ to: "/dashboard", replace: true });
    }
  });

  return (
    <div>
      <div class="flex flex-col max-w-[400px] mx-auto gap-2">
        <Button as="a" href="/api/auth/google" rel="external">
          Login with Google
        </Button>
        <Button>Log in with microsoft</Button>
        or
        <TextFieldRoot>
          <TextField />
        </TextFieldRoot>
        <Button>Send Magic Link</Button>
      </div>
    </div>
  );
}
