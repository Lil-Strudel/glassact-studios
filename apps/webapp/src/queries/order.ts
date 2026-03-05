import api from "./api";
import type { GET, Project } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export interface PlaceOrderRequest {
  projectUuid: string;
  inlayUuids: string[];
}

export async function postPlaceOrder(
  request: PlaceOrderRequest,
): Promise<GET<Project>> {
  const res = await api.post(`/project/${request.projectUuid}/place-order`, {
    inlay_uuids: request.inlayUuids,
  });
  return res.data;
}

export function postPlaceOrderOpts() {
  return mutationOptions({
    mutationFn: postPlaceOrder,
  });
}
