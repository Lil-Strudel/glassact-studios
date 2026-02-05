import { StandardTable } from "./helpers";

export type DealershipAccount = StandardTable<{
  dealership_user_id: number;
  type: string;
  provider: string;
  provider_account_id: string;
}>;
