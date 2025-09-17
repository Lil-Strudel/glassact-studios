import { RouterProvider, createRouter } from "@tanstack/solid-router";
import { QueryClient, QueryClientProvider } from "@tanstack/solid-query";
import { AuthProvider, AuthState, useAuthContext } from "./providers/auth";
import { AppStateProvider } from "./providers/app-state";
import { routeTree } from "./routeTree.gen";
import { UserProvider } from "./providers/user";

export interface RouterContext {
  queryClient: QueryClient;
  auth: AuthState;
}

const queryClient = new QueryClient();
const router = createRouter({
  routeTree,
  context: {
    queryClient,
    auth: null!,
  },
  scrollRestoration: true,
  defaultPreload: "intent",
  defaultPreloadStaleTime: 0,
  notFoundMode: "root",
});

declare module "@tanstack/solid-router" {
  interface Register {
    router: typeof router;
  }
}

function RouterWrapper() {
  const auth = useAuthContext();
  return <RouterProvider router={router} context={{ auth }} />;
}

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <AppStateProvider>
        <AuthProvider>
          <UserProvider>
            <RouterWrapper />
          </UserProvider>
        </AuthProvider>
      </AppStateProvider>
    </QueryClientProvider>
  );
}

export default App;
