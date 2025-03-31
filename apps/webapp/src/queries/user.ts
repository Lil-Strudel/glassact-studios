import { queryOptions } from "@tanstack/solid-query";
import api from "./api";

interface User {
  id: string;
  uuid: string;
  name: string;
  email: string;
  avatar: string;
  created_at: string;
  version: number;
}

export async function getUserSelf(): Promise<User> {
  const res = await api.get("/user/self");
  return await res.data;
}

export function getUserSelfOpts() {
  return queryOptions({
    queryKey: ["user", "self"],
    queryFn: getUserSelf,
  });
}
