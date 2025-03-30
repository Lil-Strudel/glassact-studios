import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import type { Component } from "solid-js";
import Routes from "./Routes";
import { AuthProvider } from "./providers/auth";

const queryClient = new QueryClient();

const App: Component = () => {
  return (
    <div>
      <QueryClientProvider client={queryClient}>
        <AuthProvider>
          <Routes />
        </AuthProvider>
      </QueryClientProvider>
    </div>
  );
};

export default App;
