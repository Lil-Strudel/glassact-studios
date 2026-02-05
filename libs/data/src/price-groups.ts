import { StandardTable } from "./helpers";

export type PriceGroup = StandardTable<{
  name: string;
  base_price_cents: number;
  description: string | null;
  is_active: boolean;
}>;
