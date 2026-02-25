import { Show, type Component } from "solid-js";
import { Form } from "@glassact/ui";
import { AnyFieldApi } from "@tanstack/solid-form";
import { useQuery } from "@tanstack/solid-query";
import { getPriceGroupsOpts } from "../queries/price-group";

interface PriceGroupComboboxProps {
  field: () => AnyFieldApi;
}
const PriceGroupCombobox: Component<PriceGroupComboboxProps> = (props) => {
  const query = useQuery(() => getPriceGroupsOpts());

  const options = () =>
    query.isSuccess
      ? query.data.items.map((priceGroup) => ({
          label: priceGroup.name,
          value: priceGroup.id,
        }))
      : [];

  return (
    <Show
      when={query.isSuccess}
      fallback={
        <Form.Combobox
          field={props.field}
          label="Price Group"
          options={[{ label: "Loading...", value: undefined }]}
        />
      }
    >
      <Form.Combobox
        field={props.field}
        label="Price Group"
        options={options()}
      />
    </Show>
  );
};

export default PriceGroupCombobox;
