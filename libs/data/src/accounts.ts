import { StandardTable } from "./helpers";

export interface Account extends StandardTable {
  user_id: number;
  type: string;
  provider: string;
  provider_account_id: string;
}
