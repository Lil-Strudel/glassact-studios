import { StandardTable } from "./helpers";

export interface Dealership extends StandardTable {
  name: string;
  address: string;
  location: [number, number];
}
