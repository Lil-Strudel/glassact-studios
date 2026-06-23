import { StandardTable } from "./helpers";

export type GlassColor = StandardTable<{
  name: string;
  hex: string;
  family: string | null;
  sort_order: number;
  is_active: boolean;
}>;
