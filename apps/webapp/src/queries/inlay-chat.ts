import { queryOptions } from "@tanstack/solid-query";
import api from "./api";

import type { InlayChat, GET, POST } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getInlayChats(): Promise<GET<InlayChat>[]> {
  const res = await api.get("/inlay-chat");
  return res.data;
}

export function getInlayChatsOpts() {
  return queryOptions({
    queryKey: ["inlay-chat"],
    queryFn: getInlayChats,
  });
}

export async function getInlayChatsByInlayUUID(
  uuid: string,
): Promise<GET<InlayChat>[]> {
  const res = await api.get(`/inlay-chat/inlay/${uuid}`);
  return res.data;
}

export function getInlayChatsByInlayUUIDOpts(uuid: string) {
  return queryOptions({
    queryKey: ["inlay-chat", "inlay", uuid],
    queryFn: () => getInlayChatsByInlayUUID(uuid),
  });
}

export async function getInlayChat(uuid: string): Promise<GET<InlayChat>> {
  const res = await api.get(`/inlay-chat/${uuid}`);
  return res.data;
}

export function getInlayChatOpts(uuid: string) {
  return () =>
    queryOptions({
      queryKey: ["inlay-chat", uuid],
      queryFn: () => getInlayChat(uuid),
    });
}

export async function postInlayChat(
  body: Omit<POST<InlayChat>, "user_id">,
): Promise<GET<InlayChat>> {
  const res = await api.post("/inlay-chat", body);
  return res.data;
}

export function postInlayChatOpts() {
  return mutationOptions({
    mutationFn: postInlayChat,
  });
}
