import { Button, TextField, TextFieldRoot } from "@glassact/ui";
import { type Component } from "solid-js";

const Login: Component = () => {
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
