import { queryOptions } from "@tanstack/solid-query";
import api from "./api";

import type {
  SimplifyDeep,
  OmitDeep,
  Project,
  GET,
  POST,
  Inlay,
} from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getProjects(): Promise<GET<Project>[]> {
  const res = await api.get("/project");
  return res.data;
}

export function getProjectsOpts() {
  return queryOptions({
    queryKey: ["project"],
    queryFn: getProjects,
  });
}

export async function getProject(uuid: string): Promise<GET<Project>> {
  const res = await api.get(`/project/${uuid}`);
  return res.data;
}

export function getProjectOpts(uuid: string) {
  return () =>
    queryOptions({
      queryKey: ["project", uuid],
      queryFn: () => getProject(uuid),
    });
}

export async function postProject(body: POST<Project>): Promise<GET<Project>> {
  const res = await api.post("/project", body);
  return res.data;
}

export function postProjectOpts() {
  return mutationOptions({
    mutationFn: postProject,
  });
}

type UnneededRefIds =
  | "project_id"
  | "catalog_info.inlay_id"
  | "custom_info.inlay_id";
export type PostProjectWithInlaysRequest = SimplifyDeep<
  POST<Project> & {
    inlays: OmitDeep<POST<Inlay>, UnneededRefIds>[];
  }
>;
export async function postProjectWithInlays(
  body: PostProjectWithInlaysRequest,
): Promise<GET<Project> & { inlays: GET<Inlay>[] }> {
  const res = await api.post("/project/with-inlays", body);
  return res.data;
}

export function postProjectWithInlaysOpts() {
  return mutationOptions({
    mutationFn: postProjectWithInlays,
  });
}
