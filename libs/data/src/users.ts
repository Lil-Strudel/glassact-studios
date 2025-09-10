import { StandardTable } from "./helpers";

export type User = StandardTable<{
  name: string;
  email: string;
  avatar: string;
  dealership_id: number;
}>;
