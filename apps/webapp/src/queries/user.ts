import { queryOptions } from "@tanstack/solid-query";
import api from "./api";

import type { User } from "@glassact/data";

export async function getUserSelf(): Promise<User> {
  const res = await api.get("/user/self");
  return res.data;
}

export function getUserSelfOpts() {
  return queryOptions({
    queryKey: ["user", "self"],
    queryFn: getUserSelf,
  });
}
