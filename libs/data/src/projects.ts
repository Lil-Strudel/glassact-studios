export interface Project {
  id: number;
  uuid: string;
  status: string;
  approved: boolean;
  dealership_id: number;
  shipment_id?: number;
  created_at: string;
  updated_at: string;
  version: number;
}
