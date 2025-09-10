import { createFileRoute, Outlet, useNavigate } from "@tanstack/solid-router";
import { showToast } from "@glassact/ui";
import { createForm, formOptions } from "@tanstack/solid-form";
import { z } from "zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import {
  postProjectWithInlaysOpts,
  PostProjectWithInlaysRequest,
} from "../../queries/project";
import { isApiError } from "../../utils/is-api-error";
import { getUserSelfOpts } from "../../queries/user";
import { createContext, useContext } from "solid-js";

export const Route = createFileRoute("/_app/projects_/create-project")({
  component: RouteComponent,
});

const formOpts = formOptions({
  defaultValues: {
    name: "",
    inlays: [] as PostProjectWithInlaysRequest["inlays"],
  },
  validators: {
    onSubmit: z.object({
      name: z.string().min(1),
      inlays: z
        .array(
          z
            .object({
              name: z.string(),
              preview_url: z.string(),
              price_group: z.int(),
            })
            .and(
              z.discriminatedUnion("type", [
                z.object({
                  type: z.literal("catalog"),
                  catalog_info: z.object({
                    catalog_item_id: z.int(),
                  }),
                }),
                z.object({
                  type: z.literal("custom"),
                  custom_info: z.object({
                    description: z.string(),
                    width: z.number(),
                    height: z.number(),
                  }),
                }),
              ]),
            ),
        )
        .min(1),
    }),
  },
});

const dummyFormJustForType = createForm(() => formOpts);
export const ProjectFormContext = createContext<typeof dummyFormJustForType>();

export function useProjectFormContext() {
  const context = useContext(ProjectFormContext);
  if (!context) {
    throw new Error("Can't find ProjectFormContext");
  }
  return context;
}

function RouteComponent() {
  const navigate = useNavigate();
  const query = useQuery(getUserSelfOpts);
  const queryClient = useQueryClient();

  const postProjectWithInlays = useMutation(postProjectWithInlaysOpts);

  const form = createForm(() => ({
    ...formOpts,
    onSubmit: async ({ value }) => {
      if (!query.isSuccess) return;

      const body: PostProjectWithInlaysRequest = {
        ...value,
        status: "awaiting-proof",
        approved: false,
        dealership_id: query.data.dealership_id,
      };

      postProjectWithInlays.mutate(body, {
        onSuccess() {
          showToast({
            title: "Created new project!",
            description: `${value.name}'s project was created.`,
            variant: "success",
          });

          navigate({ to: "/projects" });
        },
        onError(error) {
          if (isApiError(error)) {
            showToast({
              title: "Problem creating new user...",
              description: error?.data?.error ?? "Unknown error",
              variant: "error",
            });
          }
        },
        onSettled() {
          queryClient.invalidateQueries({ queryKey: ["project"] });
        },
      });
    },
  }));

  return (
    <ProjectFormContext.Provider value={form}>
      <form
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();
          form.handleSubmit();
        }}
      >
        <Outlet />
      </form>
    </ProjectFormContext.Provider>
  );
}
