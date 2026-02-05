import { StandardTable } from "./helpers";

export type InternalAccount = StandardTable<{
  internal_user_id: number;
  type: string;
  provider: string;
  provider_account_id: string;
}>;
