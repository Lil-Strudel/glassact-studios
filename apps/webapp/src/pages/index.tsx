import { createMutation, createQuery } from "@tanstack/solid-query";
import { Button, TextField, TextFieldRoot } from "@glassact/ui";
import type { Component } from "solid-js";
import { Switch, Match, createSignal } from "solid-js";
import { getCatsOpts, postCatOpts } from "../queries/cat";

interface HelloRes {
  message: string;
}
async function fetchHello(): Promise<HelloRes> {
  const res = await fetch("/api");
  return await res.json();
}

const Home: Component = () => {
  const [cat, setCat] = createSignal("");
  const query = createQuery(() => ({
    queryKey: ["todos"],
    queryFn: fetchHello,
  }));

  const catQuery = createQuery(getCatsOpts);

  const postCat = createMutation(postCatOpts);

  async function handleClick() {
    await postCat.mutateAsync({ name: cat() });
    catQuery.refetch();
  }

  return (
    <div>
      <div class="flex flex-col max-w-[400px] mx-auto gap-2">
        <Switch>
          <Match when={query.isPending}>
            <p>Loading...</p>
          </Match>
          <Match when={query.isError}>
            <p>Error: {query.error?.message}</p>
          </Match>
          <Match when={query.isSuccess}>
            <p>{query.data?.message}</p>
          </Match>
        </Switch>
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

      <div class="w-[200px]">
        <TextFieldRoot value={cat()} onChange={(v) => setCat(v)}>
          <TextField />
        </TextFieldRoot>
      </div>
      <Button onClick={handleClick}>Create Cat</Button>
      <Switch>
        <Match when={catQuery.isPending}>
          <p>Loading...</p>
        </Match>
        <Match when={catQuery.isError}>
          <p>Error: {catQuery.error?.message}</p>
        </Match>
        <Match when={catQuery.isSuccess}>
          <p>{JSON.stringify(catQuery.data)}</p>
        </Match>
      </Switch>
    </div>
  );
};

export default Home;
