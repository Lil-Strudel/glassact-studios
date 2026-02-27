import api from "./api";
import type { GET, Project } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function postPlaceOrder(
  projectUuid: string,
): Promise<GET<Project>> {
  const res = await api.post(`/project/${projectUuid}/place-order`);
  return res.data;
}

export function postPlaceOrderOpts() {
  return mutationOptions({
    mutationFn: postPlaceOrder,
  });
}
