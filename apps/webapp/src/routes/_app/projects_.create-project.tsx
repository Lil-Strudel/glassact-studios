import { createFileRoute, Link, useNavigate } from "@tanstack/solid-router";
import { Breadcrumb, Button, Form, showToast } from "@glassact/ui";
import { createForm } from "@tanstack/solid-form";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import { z } from "zod";
import { postProjectOpts } from "../../queries/project";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/projects_/create-project")({
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const postProject = useMutation(postProjectOpts);

  const form = createForm(() => ({
    defaultValues: {
      name: "",
      internal_reference: "",
    },
    validators: {
      onChange: z.object({
        name: z.string().min(1, "Name is required"),
        internal_reference: z.string(),
      }),
    },
    onSubmit: async ({ value }) => {
      const reference = value.internal_reference.trim();
      postProject.mutate(
        {
          name: value.name,
          internal_reference: reference === "" ? null : reference,
        },
        {
          onSuccess(data) {
            showToast({
              title: "Created new project!",
              description: `${value.name} was created.`,
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
    <div>
      <Breadcrumb
        crumbs={[
          { title: "Projects", to: "/projects" },
          { title: "Create Project", to: "/projects/create-project" },
        ]}
      />
      <div class="mx-auto max-w-2xl px-4 sm:px-6 lg:px-0">
        <h1 class="text-center text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
          Create a New Project
        </h1>

        <form
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
          class="mt-10 flex flex-col gap-6"
        >
          <form.Field
            name="name"
            children={(field) => (
              <Form.TextField
                field={field}
                label="Project Name"
                placeholder="John Smith Memorial"
              />
            )}
          />

          <form.Field
            name="internal_reference"
            children={(field) => (
              <Form.TextField
                field={field}
                label="PO / Internal Reference # (optional)"
                placeholder="e.g. PO-2025-0142"
              />
            )}
          />

          <div class="mt-4 flex items-center justify-center gap-3">
            <Button
              variant="outline"
              as={Link}
              to="/projects"
              disabled={postProject.isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={postProject.isPending}>
              {postProject.isPending ? "Creating..." : "Create Project"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
