import { StandardTable } from "./helpers";

// When the dealership is expected to have paid. Informational only — nothing in
// the app is gated on it, it only decides which notice a project shows.
export type PaymentTiming =
  | "pre-manufacturing"
  | "pre-shipping"
  | "post-shipping";

export const PAYMENT_TIMINGS: PaymentTiming[] = [
  "pre-manufacturing",
  "pre-shipping",
  "post-shipping",
];

export type SandblastFileFormat = "pdf" | "svg" | "png" | "dxf";

export const SANDBLAST_FILE_FORMATS: SandblastFileFormat[] = [
  "pdf",
  "svg",
  "png",
  "dxf",
];

export type Dealership = StandardTable<{
  name: string;
  // Digits only, e.g. "5551234567". Empty string means none on file.
  phone: string;
  payment_timing: PaymentTiming;
  sandblast_file_format: SandblastFileFormat;
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
