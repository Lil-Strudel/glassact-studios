import { queryOptions } from "@tanstack/solid-query";
import api from "./api";

export async function postAuthTokenAccess(): Promise<{
  access_token: string;
  access_token_exp: string;
}> {
  const res = await api.post("/auth/token/access");
  return await res.data;
}

export function postAuthTokenAccessOpts() {
  return queryOptions({
    queryKey: ["token", "authentication"],
    queryFn: postAuthTokenAccess,
  });
}
