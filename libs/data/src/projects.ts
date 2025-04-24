import { StandardTable } from "./helpers";

export interface Project extends StandardTable {
  status: string;
  approved: boolean;
  dealership_id: number;
  shipment_id?: number;
}
