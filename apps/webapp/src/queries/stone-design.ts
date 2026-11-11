import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { GET, PutStoneDesignRequest, StoneDesign } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

// Null when the project has no saved design yet; the designer starts from
// defaults rather than treating that as an error.
export async function getStoneDesign(
  projectUuid: string,
): Promise<GET<StoneDesign> | null> {
  const res = await api.get(`/project/${projectUuid}/stone-design`, {
    validateStatus: (status) => status === 200 || status === 404,
  });
  return res.status === 404 ? null : res.data;
}

export function getStoneDesignOpts(projectUuid: string) {
  return queryOptions({
    queryKey: ["project", projectUuid, "stone-design"],
    queryFn: () => getStoneDesign(projectUuid),
  });
}

export async function putStoneDesign(params: {
  projectUuid: string;
  body: PutStoneDesignRequest;
}): Promise<GET<StoneDesign>> {
  const res = await api.put(
    `/project/${params.projectUuid}/stone-design`,
    params.body,
  );
  return res.data;
}

export function putStoneDesignOpts() {
  return mutationOptions({
    mutationFn: putStoneDesign,
  });
}
