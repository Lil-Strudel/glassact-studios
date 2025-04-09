export interface Account {
  id: number;
  uuid: string;
  user_id: number;
  type: string;
  provider: string;
  provider_account_id: string;
  created_at: string;
  version: number;
}
