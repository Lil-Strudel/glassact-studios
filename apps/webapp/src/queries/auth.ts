import api from "./api";
import { mutationOptions } from "../utils/mutation-options";

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
