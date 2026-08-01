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
  DealershipUser,
  DealershipUserRole,
  PERMISSION_ACTIONS,
} from "@glassact/data";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import { getDealershipOpts } from "../../queries/dealership";
import {
  getDealershipUsersOpts,
  postDealershipUserOpts,
  patchDealershipUserOpts,
  deleteDealershipUserOpts,
} from "../../queries/user";
import { UserTable } from "../../components/user-table";
import {
  DEALERSHIP_ROLE_OPTIONS,
  buildAvatarUrl,
} from "../../utils/user-roles";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/dealership/$id/users")({
  component: RouteComponent,
});

function RouteComponent() {
  const params = Route.useParams();
  const queryClient = useQueryClient();

  const dealershipQuery = useQuery(() => getDealershipOpts(params().id));
  const dealershipId = () => dealershipQuery.data?.id;

  const usersQuery = useQuery(() => ({
    ...getDealershipUsersOpts(dealershipId()),
    enabled: dealershipId() !== undefined,
  }));

  const postUser = useMutation(() => postDealershipUserOpts());
  const patchUser = useMutation(() => patchDealershipUserOpts());
  const deleteUser = useMutation(() => deleteDealershipUserOpts());

  const [addOpen, setAddOpen] = createSignal(false);

  function invalidateUsers() {
    return queryClient.invalidateQueries({ queryKey: ["dealership-user"] });
  }

  async function updateRole(user: GET<DealershipUser>, role: string) {
    await patchUser.mutateAsync({
      uuid: user.uuid,
      body: { role: role as DealershipUserRole },
    });
    await invalidateUsers();
  }

  async function deactivate(user: GET<DealershipUser>) {
    await deleteUser.mutateAsync(user.uuid);
    await invalidateUsers();
  }

  const addForm = createForm(() => ({
    defaultValues: {
      name: "",
      email: "",
      role: "" as DealershipUserRole,
    },
    validators: {
      onSubmit: z.object({
        name: z.string().min(1),
        email: z.email(),
        role: z.enum(["viewer", "submitter", "approver", "admin"]),
      }),
    },
    onSubmit: async ({ value }) => {
      const id = dealershipId();
      if (id === undefined) return;

      postUser.mutate(
        {
          ...value,
          dealership_id: id,
          is_active: true,
          avatar: buildAvatarUrl(value.name),
        },
        {
          onSuccess() {
            setAddOpen(false);
            showToast({
              title: "Created new user!",
              description: `${value.name}'s account was created.`,
              variant: "success",
            });
            setTimeout(() => addForm.reset(), 300);
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
        <h2 class="text-xl font-semibold">Users</h2>
        <p class="text-gray-600">Manage this dealership's team members.</p>
      </div>

      <UserTable
        users={usersQuery.data ?? []}
        permission={PERMISSION_ACTIONS.MANAGE_DEALERSHIP_USERS}
        roleOptions={DEALERSHIP_ROLE_OPTIONS}
        onUpdateRole={updateRole}
        onDeactivate={deactivate}
        addAction={
          <Dialog open={addOpen()} onOpenChange={setAddOpen}>
            <DialogTrigger>
              <Button>Add a new user</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Add a new user</DialogTitle>
              </DialogHeader>
              <form
                onSubmit={(e) => {
                  e.preventDefault();
                  e.stopPropagation();
                  addForm.handleSubmit();
                }}
                class="flex flex-col gap-4"
              >
                <addForm.Field
                  name="name"
                  children={(field) => (
                    <Form.TextField field={field} label="Name" />
                  )}
                />
                <addForm.Field
                  name="email"
                  children={(field) => (
                    <Form.TextField field={field} label="Email" />
                  )}
                />
                <addForm.Field
                  name="role"
                  children={(field) => (
                    <Form.Combobox
                      field={field}
                      label="Role"
                      options={DEALERSHIP_ROLE_OPTIONS}
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
