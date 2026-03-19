import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type {
  GET,
  Notification,
  NotificationPreference,
  NotificationEventType,
} from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getNotifications(): Promise<GET<Notification>[]> {
  const res = await api.get("/notifications");
  return res.data;
}

export function getNotificationsOpts() {
  return queryOptions({
    queryKey: ["notifications"],
    queryFn: getNotifications,
    refetchInterval: 30000,
  });
}

export async function getUnreadCount(): Promise<{ count: number }> {
  const res = await api.get("/notifications/unread-count");
  return res.data;
}

export function getUnreadCountOpts() {
  return queryOptions({
    queryKey: ["notifications", "unread-count"],
    queryFn: getUnreadCount,
    refetchInterval: 15000,
  });
}

export async function getNotificationPreferences(): Promise<
  NotificationPreference[]
> {
  const res = await api.get("/notification-preferences");
  return res.data;
}

export function getNotificationPreferencesOpts() {
  return queryOptions({
    queryKey: ["notification-preferences"],
    queryFn: getNotificationPreferences,
  });
}

export async function markNotificationRead(
  uuid: string,
): Promise<GET<Notification>> {
  const res = await api.patch(`/notification/${uuid}/read`);
  return res.data;
}

export function markNotificationReadOpts() {
  return mutationOptions({
    mutationFn: markNotificationRead,
  });
}

export async function markAllNotificationsRead(): Promise<{
  success: boolean;
}> {
  const res = await api.post("/notifications/read-all");
  return res.data;
}

export function markAllNotificationsReadOpts() {
  return mutationOptions({
    mutationFn: markAllNotificationsRead,
  });
}

export async function patchNotificationPreference(params: {
  eventType: NotificationEventType;
  body: { email_enabled: boolean };
}): Promise<NotificationPreference> {
  const res = await api.patch(
    `/notification-preferences/${params.eventType}`,
    params.body,
  );
  return res.data;
}

export function patchNotificationPreferenceOpts() {
  return mutationOptions({
    mutationFn: patchNotificationPreference,
  });
}
