import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { GET, InlayChat } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getInlayChats(
  inlayUuid: string,
): Promise<GET<InlayChat>[]> {
  const res = await api.get(`/inlay/${inlayUuid}/chats`);
  return res.data;
}

export function getInlayChatsOpts(inlayUuid: string) {
  return queryOptions({
    queryKey: ["inlay", inlayUuid, "chats"],
    queryFn: () => getInlayChats(inlayUuid),
    refetchInterval: 15000,
  });
}

export interface PostInlayChatRequest {
  message: string;
  message_type: "text" | "image";
  attachment_url?: string;
}

export async function postInlayChat(params: {
  inlayUuid: string;
  body: PostInlayChatRequest;
}): Promise<GET<InlayChat>> {
  const res = await api.post(
    `/inlay/${params.inlayUuid}/chats`,
    params.body,
  );
  return res.data;
}

export function postInlayChatOpts() {
  return mutationOptions({
    mutationFn: postInlayChat,
  });
}
