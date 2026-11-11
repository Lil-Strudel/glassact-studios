import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type {
  CatalogCompound,
  CatalogCompoundWrite,
  GET,
  InlayWithInfo,
} from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function browseCompounds(): Promise<GET<CatalogCompound>[]> {
  const res = await api.get("/compound/browse");
  return res.data;
}

export function browseCompoundsOpts() {
  return queryOptions({
    queryKey: ["compound", "browse"],
    queryFn: browseCompounds,
  });
}

export async function getCompounds(): Promise<GET<CatalogCompound>[]> {
  const res = await api.get("/compound");
  return res.data;
}

export function getCompoundsOpts() {
  return queryOptions({
    queryKey: ["compound"],
    queryFn: getCompounds,
  });
}

export async function getCompound(uuid: string): Promise<GET<CatalogCompound>> {
  const res = await api.get(`/compound/${uuid}`);
  return res.data;
}

export function getCompoundOpts(uuid: string) {
  return queryOptions({
    queryKey: ["compound", uuid],
    queryFn: () => getCompound(uuid),
  });
}

export async function postCompound(
  body: CatalogCompoundWrite,
): Promise<GET<CatalogCompound>> {
  const res = await api.post("/compound", body);
  return res.data;
}

export function postCompoundOpts() {
  return mutationOptions({ mutationFn: postCompound });
}

export async function putCompound(params: {
  uuid: string;
  body: CatalogCompoundWrite & { version: number };
}): Promise<GET<CatalogCompound>> {
  const res = await api.put(`/compound/${params.uuid}`, params.body);
  return res.data;
}

export function putCompoundOpts() {
  return mutationOptions({ mutationFn: putCompound });
}

export async function patchCompound(params: {
  uuid: string;
  body: { is_active: boolean; version: number };
}): Promise<GET<CatalogCompound>> {
  const res = await api.patch(`/compound/${params.uuid}`, params.body);
  return res.data;
}

export function patchCompoundOpts() {
  return mutationOptions({ mutationFn: patchCompound });
}

export async function deleteCompound(uuid: string): Promise<void> {
  await api.delete(`/compound/${uuid}`);
}

export function deleteCompoundOpts() {
  return mutationOptions({ mutationFn: deleteCompound });
}

export async function postProjectCompound(params: {
  projectUuid: string;
  compoundUuid: string;
}): Promise<InlayWithInfo[]> {
  const res = await api.post(`/project/${params.projectUuid}/inlays/compound`, {
    compound_uuid: params.compoundUuid,
  });
  return res.data;
}

export function postProjectCompoundOpts() {
  return mutationOptions({ mutationFn: postProjectCompound });
}
