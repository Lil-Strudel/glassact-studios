import { StandardTable } from "./helpers";

export type Grout = StandardTable<{
  name: string;
  hex: string;
  sort_order: number;
  is_active: boolean;
}>;
