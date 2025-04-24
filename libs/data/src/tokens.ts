export interface Token {
  hash: Uint8Array;
  user_id: number;
  expiry: string;
  scope: string;
}
