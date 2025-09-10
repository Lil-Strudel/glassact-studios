import { queryOptions, SolidMutationOptions } from "@tanstack/solid-query";
import api from "./api";
import { mutationOptions } from "../utils/mutation-options";

export async function postAuthTokenAccess(): Promise<{
  access_token: string;
  access_token_exp: string;
}> {
  const res = await api.post("/auth/token/access");
  return res.data;
}

export function postAuthTokenAccessOpts() {
  return queryOptions({
    queryKey: ["token", "authentication"],
    queryFn: postAuthTokenAccess,
  });
}

interface postAuthMagicLinkBody {
  email: string;
}
interface postAuthMagicLinkResponse {
  message: string;
}
export async function postAuthMagicLink(
  body: postAuthMagicLinkBody,
): Promise<postAuthMagicLinkResponse> {
  const res = await api.post("/auth/magic-link", body);
  return res.data;
}

export function postAuthMagicLinkOpts() {
  return mutationOptions({
    mutationFn: postAuthMagicLink,
  });
}
