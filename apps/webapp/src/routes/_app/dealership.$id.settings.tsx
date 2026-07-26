import { createFileRoute } from "@tanstack/solid-router";
import { Show } from "solid-js";
import { Form, Button, showToast } from "@glassact/ui";
import { GET, Dealership, PERMISSION_ACTIONS } from "@glassact/data";
import { createForm } from "@tanstack/solid-form";
import { z } from "zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import {
  getDealershipOpts,
  patchDealershipOpts,
} from "../../queries/dealership";
import { useUserContext } from "../../providers/user";
import { isApiError } from "../../utils/is-api-error";

export const Route = createFileRoute("/_app/dealership/$id/settings")({
  component: RouteComponent,
});

const formSchema = z.object({
  name: z.string().min(1),
  requires_payment_before_shipping: z.boolean(),
  address: z.object({
    street: z.string().min(1),
    street_ext: z.string(),
    city: z.string().min(1),
    state: z.string().min(1),
    postal_code: z.string().min(1),
    country: z.string().min(1),
    latitude: z.preprocess(Number, z.number().gt(-90).lt(90)),
    longitude: z.preprocess(Number, z.number().gt(-180).lt(180)),
  }),
});

function RouteComponent() {
  const params = Route.useParams();
  const userContext = useUserContext();
  const query = useQuery(() => getDealershipOpts(params().id));

  const canEdit = () => userContext.can(PERMISSION_ACTIONS.MANAGE_DEALERSHIP);

  return (
    <div>
      <div class="mb-6">
        <h2 class="text-xl font-semibold">Settings</h2>
        <p class="text-gray-600">Dealership name, billing, and address.</p>
      </div>

      <Show
        when={query.data}
        fallback={<p class="text-sm text-gray-400">Loading...</p>}
      >
        {(dealership) => (
          <Show
            when={canEdit()}
            fallback={<ReadOnlyView dealership={dealership()} />}
          >
            <SettingsForm dealership={dealership()} uuid={params().id} />
          </Show>
        )}
      </Show>
    </div>
  );
}

function ReadOnlyView(props: { dealership: GET<Dealership> }) {
  const address = () => props.dealership.address;
  return (
    <dl class="max-w-xl space-y-4">
      <div>
        <dt class="text-sm font-medium text-gray-500">Name</dt>
        <dd class="text-gray-900">{props.dealership.name}</dd>
      </div>
      <div>
        <dt class="text-sm font-medium text-gray-500">
          Requires payment before shipping
        </dt>
        <dd class="text-gray-900">
          {props.dealership.requires_payment_before_shipping ? "Yes" : "No"}
        </dd>
      </div>
      <div>
        <dt class="text-sm font-medium text-gray-500">Address</dt>
        <dd class="text-gray-900">
          {address().street}
          {address().street_ext ? `, ${address().street_ext}` : ""}, {address().city},{" "}
          {address().state} {address().postal_code}, {address().country}
        </dd>
      </div>
    </dl>
  );
}

function SettingsForm(props: { dealership: GET<Dealership>; uuid: string }) {
  const queryClient = useQueryClient();
  const patchDealership = useMutation(() => patchDealershipOpts());

  const form = createForm(() => ({
    defaultValues: {
      name: props.dealership.name,
      requires_payment_before_shipping:
        props.dealership.requires_payment_before_shipping,
      address: {
        street: props.dealership.address.street,
        street_ext: props.dealership.address.street_ext,
        city: props.dealership.address.city,
        state: props.dealership.address.state,
        postal_code: props.dealership.address.postal_code,
        country: props.dealership.address.country,
        latitude: props.dealership.address.latitude as unknown as number,
        longitude: props.dealership.address.longitude as unknown as number,
      },
    },
    validators: {
      onSubmit: formSchema,
    },
    onSubmit: async ({ value }) => {
      const output = formSchema.parse(value);
      patchDealership.mutate(
        { uuid: props.uuid, body: output },
        {
          onSuccess() {
            showToast({
              title: "Saved",
              description: "Dealership settings were updated.",
              variant: "success",
            });
          },
          onError(error) {
            if (isApiError(error)) {
              showToast({
                title: "Problem saving settings...",
                description: error?.data?.error ?? "Unknown error",
                variant: "error",
              });
            }
          },
          onSettled() {
            queryClient.invalidateQueries({ queryKey: ["dealership"] });
          },
        },
      );
    },
  }));

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        e.stopPropagation();
        form.handleSubmit();
      }}
      class="flex max-w-xl flex-col gap-4"
    >
      <form.Field
        name="name"
        children={(field) => <Form.TextField field={field} label="Name" />}
      />

      <form.Field
        name="requires_payment_before_shipping"
        children={(field) => (
          <Form.Checkbox
            field={field}
            label="Require payment before shipping"
            description="Projects for this dealership show a notice that they will not ship until the invoice is paid. This does not block shipping."
          />
        )}
      />

      <div class="border-t pt-4">
        <h3 class="mb-3 text-sm font-medium text-gray-900">Address</h3>

        <Form.AddressField
          form={form}
          name="address"
          apiKey={import.meta.env.VITE_GOOGLE_MAPS_API_KEY}
          label="Search to replace address"
          class="mb-4"
        />

        <div class="grid grid-cols-2 gap-4">
          <form.Field
            name="address.street"
            children={(field) => (
              <Form.TextField field={field} label="Street" />
            )}
          />
          <form.Field
            name="address.street_ext"
            children={(field) => (
              <Form.TextField field={field} label="Street (line 2)" />
            )}
          />
          <form.Field
            name="address.city"
            children={(field) => (
              <Form.TextField field={field} label="City" />
            )}
          />
          <form.Field
            name="address.state"
            children={(field) => (
              <Form.TextField field={field} label="State" />
            )}
          />
          <form.Field
            name="address.postal_code"
            children={(field) => (
              <Form.TextField field={field} label="Postal code" />
            )}
          />
          <form.Field
            name="address.country"
            children={(field) => (
              <Form.TextField field={field} label="Country (ISO-2)" />
            )}
          />
        </div>
      </div>

      <Button type="submit" disabled={patchDealership.isPending}>
        Save changes
      </Button>
    </form>
  );
}
