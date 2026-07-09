import { StandardTable } from "./helpers";

export type Dealership = StandardTable<{
  name: string;
  requires_payment_before_shipping: boolean;
  address: {
    street: string;
    street_ext: string;
    city: string;
    state: string;
    postal_code: string;
    country: string;
    latitude: number;
    longitude: number;
  };
}>;
