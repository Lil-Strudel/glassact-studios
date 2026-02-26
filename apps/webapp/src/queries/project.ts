import { queryOptions } from "@tanstack/solid-query";
import api from "./api";

import type { Project, GET, InlayType, InlayWithInfo } from "@glassact/data";
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
  return queryOptions({
    queryKey: ["project", uuid],
    queryFn: () => getProject(uuid),
  });
}

export async function postProject(body: {
  name: string;
}): Promise<GET<Project>> {
  const res = await api.post("/project", body);
  return res.data;
}

export function postProjectOpts() {
  return mutationOptions({
    mutationFn: postProject,
  });
}

export interface PostProjectWithInlaysInlay {
  name: string;
  type: InlayType;
  preview_url: string;
  catalog_info?: {
    catalog_item_id: number;
    customization_notes: string;
  };
  custom_info?: {
    description: string;
    requested_width: number;
    requested_height: number;
  };
}

export type PostProjectWithInlaysRequest = {
  name: string;
  inlays: PostProjectWithInlaysInlay[];
};

export async function postProjectWithInlays(
  body: PostProjectWithInlaysRequest,
): Promise<GET<Project> & { inlays: InlayWithInfo[] }> {
  const res = await api.post("/project/with-inlays", body);
  return res.data;
}

export function postProjectWithInlaysOpts() {
  return mutationOptions({
    mutationFn: postProjectWithInlays,
  });
}

export async function patchProject(params: {
  uuid: string;
  body: { name?: string };
}): Promise<GET<Project>> {
  const res = await api.patch(`/project/${params.uuid}`, params.body);
  return res.data;
}

export function patchProjectOpts() {
  return mutationOptions({
    mutationFn: patchProject,
  });
}

export async function deleteProject(uuid: string): Promise<GET<Project>> {
  const res = await api.delete(`/project/${uuid}`);
  return res.data;
}

export function deleteProjectOpts() {
  return mutationOptions({
    mutationFn: deleteProject,
  });
}
