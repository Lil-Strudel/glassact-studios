import { createFileRoute, redirect } from "@tanstack/solid-router";
import { getUserSelfOpts } from "../../queries/user";
import { getDealershipSelfOpts } from "../../queries/dealership";

export const Route = createFileRoute("/_app/dealership/")({
  component: () => null,
  beforeLoad: async ({ context }) => {
    const user = await context.queryClient.ensureQueryData(getUserSelfOpts());
    if (!("dealership_id" in user)) {
      throw redirect({ to: "/", replace: true });
    }

    const dealership = await context.queryClient.ensureQueryData(
      getDealershipSelfOpts(),
    );

    throw redirect({
      to: "/dealership/$id/users",
      params: { id: dealership.uuid },
      replace: true,
    });
  },
});
