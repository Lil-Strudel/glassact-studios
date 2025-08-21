import { StandardTable } from "./helpers";

export interface Dealership extends StandardTable {
  name: string;
  address: {
    street: string;
    street_ext: string | null;
    city: string;
    state: string;
    postal_code: string;
    country: string;
    latitude: number;
    longitude: number;
  };
}
