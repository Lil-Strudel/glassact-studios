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
          { title: "New project", to: "/projects/create-project" },
        ]}
      />
      <div class="mx-auto mt-6 flex max-w-xl flex-col gap-8">
        <div class="flex flex-col items-center gap-3 text-center">
          <p class="eyebrow">New project</p>
          <h1 class="text-3xl sm:text-4xl">Create a new project</h1>
          <div class="ornament-rule w-40" aria-hidden="true">
            <span class="lozenge" />
          </div>
          <p class="max-w-prose text-sm text-muted-foreground">
            You'll add inlays and design the stone next.
          </p>
        </div>

        <form
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
          class="flex flex-col gap-6 rounded-lg border border-border bg-card p-6 sm:p-8"
        >
          <form.Field
            name="name"
            children={(field) => (
              <Form.TextField
                field={field}
                label="Project name"
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

          <div class="flex items-center justify-end gap-3 border-t border-border pt-6">
            <Button
              variant="ghost"
              as={Link}
              to="/projects"
              disabled={postProject.isPending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={postProject.isPending}>
              {postProject.isPending ? "Creating..." : "Create project"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
