import { queryOptions } from "@tanstack/solid-query";
import api from "./api";

import type { Dealership, GET, POST, PATCH } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getDealerships(): Promise<GET<Dealership>[]> {
  const res = await api.get("/dealership");
  return res.data;
}

export function getDealershipsOpts() {
  return queryOptions({
    queryKey: ["dealership"],
    queryFn: getDealerships,
  });
}

export async function getDealership(uuid: string): Promise<GET<Dealership>> {
  const res = await api.get(`/dealership/${uuid}`);
  return res.data;
}

export function getDealershipOpts(uuid: string) {
  return queryOptions({
    queryKey: ["dealership", uuid],
    queryFn: () => getDealership(uuid),
  });
}

export async function getDealershipSelf(): Promise<GET<Dealership>> {
  const res = await api.get("/dealership/self");
  return res.data;
}

export function getDealershipSelfOpts() {
  return queryOptions({
    queryKey: ["dealership", "self"],
    queryFn: getDealershipSelf,
  });
}

export async function postDealership(
  body: POST<Dealership>,
): Promise<GET<Dealership>> {
  const res = await api.post("/dealership", body);
  return res.data;
}

export function postDealershipOpts() {
  return mutationOptions({
    mutationFn: postDealership,
  });
}

export async function patchDealership(args: {
  uuid: string;
  body: PATCH<Dealership>;
}): Promise<GET<Dealership>> {
  const res = await api.patch(`/dealership/${args.uuid}`, args.body);
  return res.data;
}

export function patchDealershipOpts() {
  return mutationOptions({
    mutationFn: patchDealership,
  });
}
