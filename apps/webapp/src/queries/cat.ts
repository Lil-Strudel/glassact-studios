import { queryOptions, SolidMutationOptions } from "@tanstack/solid-query";

interface Cat {
  id: number;
  name: string;
}

export async function getCats(): Promise<Cat[]> {
  const res = await fetch("/api/cat");
  return await res.json();
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
  const res = await fetch("/api/cat", {
    method: "POST",
    body: JSON.stringify(data),
    headers: {
      "content-type": "application/json",
    },
  });
  return await res.json();
}

export function postCatOpts() {
  return {
    mutationKey: ["cat"],
    mutationFn: postCat,
  };
}
