import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type {
  GET,
  SandblastBackground,
  SandblastBackgroundUpdate,
  SandblastBackgroundWrite,
} from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function browseBackgrounds(): Promise<GET<SandblastBackground>[]> {
  const res = await api.get("/background/browse");
  return res.data;
}

export function browseBackgroundsOpts() {
  return queryOptions({
    queryKey: ["background", "browse"],
    queryFn: browseBackgrounds,
  });
}

export async function getBackgrounds(): Promise<GET<SandblastBackground>[]> {
  const res = await api.get("/background");
  return res.data;
}

export function getBackgroundsOpts() {
  return queryOptions({
    queryKey: ["background"],
    queryFn: getBackgrounds,
  });
}

export async function getBackground(
  uuid: string,
): Promise<GET<SandblastBackground>> {
  const res = await api.get(`/background/${uuid}`);
  return res.data;
}

export function getBackgroundOpts(uuid: string) {
  return queryOptions({
    queryKey: ["background", uuid],
    queryFn: () => getBackground(uuid),
  });
}

// Fetched same-origin so the designer can inline the mask without S3 CORS.
export async function getBackgroundSvgText(uuid: string): Promise<string> {
  const res = await api.get(`/background/${uuid}/svg`, {
    responseType: "text",
  });
  return res.data as string;
}

export function getBackgroundSvgTextOpts(uuid: string) {
  return queryOptions({
    queryKey: ["background", uuid, "svg"],
    queryFn: () => getBackgroundSvgText(uuid),
    staleTime: 1000 * 60 * 30,
  });
}

export async function getBackgroundSandblastUrl(
  uuid: string,
): Promise<{ url: string }> {
  const res = await api.get(`/background/${uuid}/sandblast`);
  return res.data;
}

export async function postPrepareBackground(body: {
  svg_url: string;
}): Promise<{ mask_svg: string; warnings: string[] }> {
  const res = await api.post("/background/prepare", body);
  return res.data;
}

export function postPrepareBackgroundOpts() {
  return mutationOptions({ mutationFn: postPrepareBackground });
}

export async function postBackground(
  body: SandblastBackgroundWrite,
): Promise<GET<SandblastBackground>> {
  const res = await api.post("/background", body);
  return res.data;
}

export function postBackgroundOpts() {
  return mutationOptions({ mutationFn: postBackground });
}

export async function putBackground(params: {
  uuid: string;
  body: SandblastBackgroundUpdate;
}): Promise<GET<SandblastBackground>> {
  const res = await api.put(`/background/${params.uuid}`, params.body);
  return res.data;
}

export function putBackgroundOpts() {
  return mutationOptions({ mutationFn: putBackground });
}

export async function patchBackground(params: {
  uuid: string;
  body: { is_active: boolean; version: number };
}): Promise<GET<SandblastBackground>> {
  const res = await api.patch(`/background/${params.uuid}`, params.body);
  return res.data;
}

export function patchBackgroundOpts() {
  return mutationOptions({ mutationFn: patchBackground });
}
