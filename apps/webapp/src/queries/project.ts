import { queryOptions } from "@tanstack/solid-query";
import api from "./api";

import type {
  Project,
  ProjectDetail,
  ProjectListItem,
  ProjectWatcherSummary,
  SetProjectWatchResponse,
  GET,
} from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getProjects(): Promise<ProjectListItem[]> {
  const res = await api.get("/project");
  return res.data;
}

export function getProjectsOpts() {
  return queryOptions({
    queryKey: ["project"],
    queryFn: getProjects,
  });
}

export async function getProject(uuid: string): Promise<ProjectDetail> {
  const res = await api.get(`/project/${uuid}`);
  return res.data;
}

export function getProjectOpts(uuid: string) {
  return queryOptions({
    queryKey: ["project", uuid],
    queryFn: () => getProject(uuid),
    refetchInterval: 60_000,
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

export async function postMarkProjectShipped(params: {
  uuid: string;
  trackingNumber: string;
}): Promise<GET<Project>> {
  const res = await api.post(`/project/${params.uuid}/ship`, {
    tracking_number: params.trackingNumber,
  });
  return res.data;
}

export function postMarkProjectShippedOpts() {
  return mutationOptions({
    mutationFn: postMarkProjectShipped,
  });
}

export async function postMarkProjectDelivered(
  uuid: string,
): Promise<GET<Project>> {
  const res = await api.post(`/project/${uuid}/deliver`);
  return res.data;
}

export function postMarkProjectDeliveredOpts() {
  return mutationOptions({
    mutationFn: postMarkProjectDelivered,
  });
}

export async function setProjectWatch(params: {
  uuid: string;
  isWatching: boolean;
}): Promise<SetProjectWatchResponse> {
  const res = await api.put(`/project/${params.uuid}/watch`, {
    is_watching: params.isWatching,
  });
  return res.data;
}

export function setProjectWatchOpts() {
  return mutationOptions({
    mutationFn: setProjectWatch,
  });
}

export async function getProjectWatchers(
  uuid: string,
): Promise<ProjectWatcherSummary[]> {
  const res = await api.get(`/project/${uuid}/watchers`);
  return res.data;
}

export function getProjectWatchersOpts(uuid: string) {
  return queryOptions({
    queryKey: ["project", uuid, "watchers"],
    queryFn: () => getProjectWatchers(uuid),
  });
}
