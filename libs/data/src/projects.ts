import { StandardTable } from "./helpers";

export interface Project extends StandardTable {
  name: string;
  status: string;
  approved: boolean;
  dealership_id: number;
  shipment_id?: number;
}
