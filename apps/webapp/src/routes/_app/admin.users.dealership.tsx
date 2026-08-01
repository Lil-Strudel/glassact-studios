import { createFileRoute, Link } from "@tanstack/solid-router";
import { createMemo, createSignal, Show } from "solid-js";
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
import { ColumnDef } from "@tanstack/solid-table";
import {
  GET,
  DealershipUser,
  DealershipUserRole,
  PERMISSION_ACTIONS,
} from "@glassact/data";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import {
  getDealershipUsersOpts,
  postDealershipUserOpts,
  patchDealershipUserOpts,
  deleteDealershipUserOpts,
} from "../../queries/user";
import { getDealershipsOpts } from "../../queries/dealership";
import DealershipCombobox from "../../components/dealership-combobox";
import { UserTable } from "../../components/user-table";
import {
  DEALERSHIP_ROLE_OPTIONS,
  buildAvatarUrl,
} from "../../utils/user-roles";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/admin/users/dealership")({
  component: RouteComponent,
});

function RouteComponent() {
  const usersQuery = useQuery(() => getDealershipUsersOpts());
  const dealershipsQuery = useQuery(() => getDealershipsOpts());
  const queryClient = useQueryClient();

  const postUser = useMutation(() => postDealershipUserOpts());
  const patchUser = useMutation(() => patchDealershipUserOpts());
  const deleteUser = useMutation(() => deleteDealershipUserOpts());

  const [dialogOpen, setDialogOpen] = createSignal(false);

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

  async function setActive(user: GET<DealershipUser>, isActive: boolean) {
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

  const dealershipsById = createMemo(
    () => new Map((dealershipsQuery.data ?? []).map((d) => [d.id, d])),
  );

  const dealershipColumn = createMemo<ColumnDef<GET<DealershipUser>>[]>(() => [
    {
      accessorKey: "dealership_id",
      header: "Dealership",
      cell: (info) => (
        <Show
          when={dealershipsById().get(info.row.original.dealership_id)}
          fallback={info.row.original.dealership_id}
        >
          {(dealership) => (
            <Link
              to="/dealership/$id/users"
              params={{ id: dealership().uuid }}
              class="text-primary hover:underline"
            >
              {dealership().name}
            </Link>
          )}
        </Show>
      ),
    },
  ]);

  const form = createForm(() => ({
    defaultValues: {
      name: "",
      email: "",
      dealership_id: undefined as unknown as number,
      role: "" as DealershipUserRole,
    },
    validators: {
      onSubmit: z.object({
        name: z.string().min(1),
        email: z.email(),
        dealership_id: z.number().int(),
        role: z.enum(["viewer", "submitter", "approver", "admin"]),
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
        <h1 class="text-2xl font-bold">Dealership Users</h1>
        <p class="text-gray-600">Manage users from dealerships</p>
      </div>

      <UserTable
        users={usersQuery.data ?? []}
        permission={PERMISSION_ACTIONS.MANAGE_DEALERSHIP_USERS}
        roleOptions={DEALERSHIP_ROLE_OPTIONS}
        extraColumns={dealershipColumn()}
        onUpdateRole={updateRole}
        onSetActive={setActive}
        addAction={
          <Dialog open={dialogOpen()} onOpenChange={setDialogOpen}>
            <DialogTrigger>
              <Button>Add a new user</Button>
            </DialogTrigger>
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Add a new dealership user</DialogTitle>
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
                  name="dealership_id"
                  children={(field) => <DealershipCombobox field={field} />}
                />
                <form.Field
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
