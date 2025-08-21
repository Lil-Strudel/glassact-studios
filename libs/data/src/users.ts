import { StandardTable } from "./helpers";

export interface User extends StandardTable {
  name: string;
  email: string;
  avatar: string;
  dealership_id: number;
}
