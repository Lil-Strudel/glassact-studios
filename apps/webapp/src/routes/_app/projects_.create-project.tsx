import { createFileRoute, Outlet, useNavigate } from "@tanstack/solid-router";
import { showToast } from "@glassact/ui";
import { createForm, formOptions } from "@tanstack/solid-form";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import {
  postProjectWithInlaysOpts,
  PostProjectWithInlaysRequest,
} from "../../queries/project";
import { isApiError } from "../../utils/is-api-error";
import { createContext, useContext } from "solid-js";

export const Route = createFileRoute("/_app/projects_/create-project")({
  component: RouteComponent,
});

const formOpts = formOptions({
  defaultValues: {
    name: "",
    inlays: [] as PostProjectWithInlaysRequest["inlays"],
  },
  validators: {},
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
  const queryClient = useQueryClient();

  const postProjectWithInlays = useMutation(postProjectWithInlaysOpts);

  const form = createForm(() => ({
    ...formOpts,
    onSubmit: async ({ value }) => {
      postProjectWithInlays.mutate(
        { name: value.name, inlays: value.inlays },
        {
          onSuccess(data) {
            showToast({
              title: "Created new project!",
              description: `${value.name}'s project was created.`,
              variant: "success",
            });
            queryClient.invalidateQueries({ queryKey: ["project"] });
            navigate({ to: `/projects/${data.uuid}` });
          },
          onError(error) {
            if (isApiError(error)) {
              showToast({
                title: "Problem creating project",
                description: error?.data?.error ?? "Unknown error",
                variant: "error",
              });
            }
          },
        },
      );
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
