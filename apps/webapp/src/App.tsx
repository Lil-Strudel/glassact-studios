import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import type { Component } from "solid-js";
import Routes from "./Routes";

const queryClient = new QueryClient();

const App: Component = () => {
  return (
    <div>
      <QueryClientProvider client={queryClient}>
        <Routes />
      </QueryClientProvider>
    </div>
  );
};

export default App;
