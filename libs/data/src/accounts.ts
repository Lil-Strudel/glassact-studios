import { StandardTable } from "./helpers";

export type Account = StandardTable<{
  user_id: number;
  type: string;
  provider: string;
  provider_account_id: string;
}>;
