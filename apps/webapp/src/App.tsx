import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import type { Component } from "solid-js";
import Routes from "./Routes";
import { AuthProvider } from "./providers/auth";
import { AppStateProvider } from "./providers/app-state";

const queryClient = new QueryClient();

const App: Component = () => {
  return (
    <div>
      <QueryClientProvider client={queryClient}>
        <AppStateProvider>
          <AuthProvider>
            <Routes />
          </AuthProvider>
        </AppStateProvider>
      </QueryClientProvider>
    </div>
  );
};

export default App;
