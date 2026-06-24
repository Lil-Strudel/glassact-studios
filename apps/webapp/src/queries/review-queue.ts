import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { ReviewQueue } from "@glassact/data";

export async function getReviewQueue(): Promise<ReviewQueue> {
  const res = await api.get("/review-queue");
  return res.data;
}

export function getReviewQueueOpts() {
  return queryOptions({
    queryKey: ["review-queue"],
    queryFn: getReviewQueue,
  });
}
