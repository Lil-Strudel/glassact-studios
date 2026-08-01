import { createFileRoute } from "@tanstack/solid-router";
import { createSignal } from "solid-js";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  Form,
  Button,
  showToast,
} from "@glassact/ui";
import {
  GET,
  InternalUser,
  InternalUserRole,
  PERMISSION_ACTIONS,
} from "@glassact/data";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import {
  getInternalUsersOpts,
  postInternalUserOpts,
  patchInternalUserOpts,
  deleteInternalUserOpts,
} from "../../queries/user";
import { UserTable } from "../../components/user-table";
import { INTERNAL_ROLE_OPTIONS, buildAvatarUrl } from "../../utils/user-roles";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/admin/users/internal")({
  component: RouteComponent,
});

function RouteComponent() {
  const usersQuery = useQuery(() => getInternalUsersOpts());
  const queryClient = useQueryClient();

  const postUser = useMutation(() => postInternalUserOpts());
  const patchUser = useMutation(() => patchInternalUserOpts());
  const deleteUser = useMutation(() => deleteInternalUserOpts());

  const [dialogOpen, setDialogOpen] = createSignal(false);

  function invalidateUsers() {
    return queryClient.invalidateQueries({ queryKey: ["internal-user"] });
  }

  async function updateRole(user: GET<InternalUser>, role: string) {
    await patchUser.mutateAsync({
      uuid: user.uuid,
      body: { role: role as InternalUserRole },
    });
    await invalidateUsers();
  }

  async function setActive(user: GET<InternalUser>, isActive: boolean) {
    if (isActive) {
      await patchUser.mutateAsync({
        uuid: user.uuid,
        body: { is_active: true },
      });
    } else {
      await deleteUser.mutateAsync(user.uuid);
    }
    await invalidateUsers();
  }

  const form = createForm(() => ({
    defaultValues: {
      name: "",
      email: "",
      role: "" as InternalUserRole,
    },
    validators: {
      onSubmit: z.object({
        name: z.string().min(1),
        email: z.email(),
        role: z.enum(["designer", "production", "billing", "admin"]),
      }),
    },
    onSubmit: async ({ value }) => {
      postUser.mutate(
        {
          ...value,
          is_active: true,
          avatar: buildAvatarUrl(value.name),
        },
        {
          onSuccess() {
            setDialogOpen(false);
            showToast({
              title: "Created new user!",
              description: `${value.name}'s account was created.`,
              variant: "success",
            });
            setTimeout(() => form.reset(), 300);
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
          onSettled: invalidateUsers,
        },
      );
    },
  }));

  return (
    <div>
      <div class="mb-6">
        <h1 class="text-2xl font-bold">Internal Users</h1>
        <p class="text-gray-600">Manage GlassAct Studios staff</p>
      </div>

      <UserTable
        users={usersQuery.data ?? []}
        permission={PERMISSION_ACTIONS.MANAGE_INTERNAL_USERS}
        roleOptions={INTERNAL_ROLE_OPTIONS}
        onUpdateRole={updateRole}
        onSetActive={setActive}
        addAction={
          <Dialog open={dialogOpen()} onOpenChange={setDialogOpen}>
            <DialogTrigger>
              <Button>Add a new user</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Add a new internal user</DialogTitle>
              </DialogHeader>
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  form.handleSubmit();
                }}
                class="flex flex-col gap-4"
              >
                <form.Field
                  name="name"
                  children={(field) => (
                    <Form.TextField field={field} label="Name" />
                  )}
                />
                <form.Field
                  name="email"
                  children={(field) => (
                    <Form.TextField field={field} label="Email" />
                  )}
                />
                <form.Field
                  name="role"
                  children={(field) => (
                    <Form.Combobox
                      field={field}
                      label="Role"
                      options={INTERNAL_ROLE_OPTIONS}
                    />
                  )}
                />
                <Button type="submit" disabled={postUser.isPending}>
                  Add
                </Button>
              </form>
            </DialogContent>
          </Dialog>
        }
      />
    </div>
  );
}
