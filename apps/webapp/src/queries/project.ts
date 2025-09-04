import { queryOptions, SolidMutationOptions } from "@tanstack/solid-query";
import api from "./api";

import type { Project, GET, POST } from "@glassact/data";

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

export function postProjectsOpts(): SolidMutationOptions<
  GET<Project>,
  Error,
  POST<Project>,
  unknown
> {
  return {
    mutationFn: postProject,
  };
}
