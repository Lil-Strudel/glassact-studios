import { createMemo, Show, type Component } from "solid-js";
import { Badge } from "@glassact/ui";
import { useQuery } from "@tanstack/solid-query";
import { getSupportPriceGroupsOpts } from "../queries/support";

interface PriceGroupTagProps {
  priceGroupId: number;
}

// Catalog responses carry only default_price_group_id, so the name is resolved
// client-side. This reads the support endpoint rather than getPriceGroupsOpts:
// GET /api/price-groups is gated behind manage_price_groups and would 403 for
// every dealership user browsing the catalog.
const PriceGroupTag: Component<PriceGroupTagProps> = (props) => {
  const query = useQuery(() => getSupportPriceGroupsOpts());

  const name = createMemo(
    () =>
      query.data?.find((priceGroup) => priceGroup.id === props.priceGroupId)
        ?.name ?? null,
  );

  return (
    <Show when={name()}>
      {(priceGroupName) => (
        <Badge variant="secondary">{priceGroupName()}</Badge>
      )}
    </Show>
  );
};

export default PriceGroupTag;
