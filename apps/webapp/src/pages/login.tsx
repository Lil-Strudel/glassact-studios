import { Button, TextField, TextFieldRoot } from "@glassact/ui";
import { createEffect, type Component } from "solid-js";
import { useAuthContext } from "../providers/auth";
import { useNavigate } from "@solidjs/router";

const Login: Component = () => {
  const { state } = useAuthContext();
  const navigate = useNavigate();

  createEffect(() => {
    if (state.status === "authenticated") {
      navigate("/dashboard", { replace: true });
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
};

export default Login;
