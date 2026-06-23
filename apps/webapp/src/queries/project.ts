import { queryOptions } from "@tanstack/solid-query";
import api from "./api";

import type { Project, GET } from "@glassact/data";
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

export interface PostProjectRequest {
  name: string;
  internal_reference?: string | null;
}

export async function postProject(
  body: PostProjectRequest,
): Promise<GET<Project>> {
  const res = await api.post("/project", body);
  return res.data;
}

export function postProjectOpts() {
  return mutationOptions({
    mutationFn: postProject,
  });
}

export interface PatchProjectRequest {
  name?: string;
  internal_reference?: string | null;
}

export async function patchProject(params: {
  uuid: string;
  body: PatchProjectRequest;
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
