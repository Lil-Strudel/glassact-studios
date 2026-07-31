import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { GET, ProjectChat } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getProjectChats(
  projectUuid: string,
): Promise<GET<ProjectChat>[]> {
  const res = await api.get(`/project/${projectUuid}/chats`);
  return res.data;
}

export function getProjectChatsOpts(projectUuid: string) {
  return queryOptions({
    queryKey: ["project", projectUuid, "chats"],
    queryFn: () => getProjectChats(projectUuid),
    refetchInterval: 15000,
  });
}

export interface PostProjectChatRequest {
  message: string;
  message_type: "text" | "image";
  attachment_url?: string;
  // Tags the message with the inlay it is about; the thread labels it and links
  // there. Omit for a message about the project as a whole.
  inlay_uuid?: string;
}

export async function postProjectChat(params: {
  projectUuid: string;
  body: PostProjectChatRequest;
}): Promise<GET<ProjectChat>> {
  const res = await api.post(
    `/project/${params.projectUuid}/chats`,
    params.body,
  );
  return res.data;
}

export function postProjectChatOpts() {
  return mutationOptions({
    mutationFn: postProjectChat,
  });
}
