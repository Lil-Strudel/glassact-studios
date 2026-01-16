import { queryOptions } from "@tanstack/solid-query";
import api from "./api";

import type {
  SimplifyDeep,
  OmitDeep,
  Project,
  GET,
  POST,
  Inlay,
  Simplify,
} from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

type getProjectProps = {
  expand?: {
    inlays?: boolean;
  };
};

// prettier-ignore
type ExpandWithDefaults<T extends getProjectProps> = {
  inlays: T["expand"] extends { inlays: infer I } ? I extends true ? true : false : false;
};

// prettier-ignore
type ProjectWithExpands<T extends getProjectProps> = Simplify<
  GET<Project> &
    (ExpandWithDefaults<T>["inlays"] extends true ? { inlays: GET<Inlay>[] } : {})
>;

export async function getProjects<T extends getProjectProps = {}>(
  props: T = {} as T,
): Promise<ProjectWithExpands<T>[]> {
  const expand = Object.entries(props.expand || {})
    .filter(([, value]) => value)
    .map(([key]) => key);

  const params = {
    ...(expand.length ? { expand: expand.join(",") } : {}),
  };

  const res = await api.get("/project", { params });

  return res.data;
}

export function getProjectsOpts<T extends getProjectProps = {}>(
  props: T = {} as T,
) {
  return () =>
    queryOptions({
      queryKey: ["project", props],
      queryFn: () => getProjects(props),
    });
}

export async function getProject<T extends getProjectProps = {}>(
  uuid: string,
  props: T = {} as T,
): Promise<ProjectWithExpands<T>> {
  const expand = Object.entries(props.expand || {})
    .filter(([, value]) => value)
    .map(([key]) => key);

  const params = {
    ...(expand.length ? { expand: expand.join(",") } : {}),
  };

  const res = await api.get(`/project/${uuid}`, { params });
  return res.data;
}

export function getProjectOpts<T extends getProjectProps = {}>(
  uuid: string,
  props: T = {} as T,
) {
  return () =>
    queryOptions({
      queryKey: ["project", uuid, props],
      queryFn: () => getProject(uuid, props),
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
