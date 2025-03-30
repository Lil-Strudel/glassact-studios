import { queryOptions } from "@tanstack/solid-query";
import api from "./api";

interface Cat {
  id: number;
  name: string;
}

export async function getCats(): Promise<Cat[]> {
  const res = await api.get("/cat");
  return await res.data;
}

export function getCatsOpts() {
  return queryOptions({
    queryKey: ["cat"],
    queryFn: getCats,
  });
}

interface PostCatData {
  name: string;
}
export async function postCat(data: PostCatData): Promise<Cat[]> {
  const res = await api.post("/cat", data);
  return await res.data;
}

export function postCatOpts() {
  return {
    mutationKey: ["cat"],
    mutationFn: postCat,
  };
}
