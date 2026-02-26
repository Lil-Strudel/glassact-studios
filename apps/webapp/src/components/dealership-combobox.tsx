import { Show, type Component } from "solid-js";
import { Form } from "@glassact/ui";
import { AnyFieldApi } from "@tanstack/solid-form";
import { useQuery } from "@tanstack/solid-query";
import { getDealershipsOpts } from "../queries/dealership";

interface DealershipComboboxProps {
  field: () => AnyFieldApi;
}
const DealershipCombobox: Component<DealershipComboboxProps> = (props) => {
  const query = useQuery(() => getDealershipsOpts());

  const options = () =>
    query.isSuccess
      ? query.data.map((dealership) => ({
          label: dealership.name,
          value: dealership.id,
        }))
      : [];

  return (
    <Show
      when={query.isSuccess}
      fallback={
        <Form.Combobox
          field={props.field}
          label="Dealership"
          options={[{ label: "Loading...", value: undefined }]}
        />
      }
    >
      <Form.Combobox
        field={props.field}
        label="Dealership"
        options={options()}
      />
    </Show>
  );
};

export default DealershipCombobox;
