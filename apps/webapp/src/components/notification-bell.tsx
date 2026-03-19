import { For, Show } from "solid-js";
import {
  Button,
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@glassact/ui";
import { IoNotificationsOutline } from "solid-icons/io";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import {
  getUnreadCountOpts,
  getNotificationsOpts,
  markNotificationReadOpts,
  markAllNotificationsReadOpts,
} from "../queries/notifications";

function formatRelativeTime(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const seconds = Math.floor(diff / 1000);
  if (seconds < 60) return "just now";
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

export function NotificationBell() {
  const queryClient = useQueryClient();
  const unreadCount = useQuery(() => getUnreadCountOpts());
  const notifications = useQuery(() => getNotificationsOpts());

  const markRead = useMutation(() => markNotificationReadOpts());
  const markAllRead = useMutation(() => markAllNotificationsReadOpts());

  const recentNotifications = () => (notifications.data ?? []).slice(0, 10);

  const hasUnread = () => (unreadCount.data?.count ?? 0) > 0;

  function handleMarkRead(uuid: string) {
    markRead.mutate(uuid, {
      onSuccess() {
        queryClient.invalidateQueries({ queryKey: ["notifications"] });
        queryClient.invalidateQueries({
          queryKey: ["notifications", "unread-count"],
        });
      },
    });
  }

  function handleMarkAllRead() {
    markAllRead.mutate(undefined, {
      onSuccess() {
        queryClient.invalidateQueries({ queryKey: ["notifications"] });
        queryClient.invalidateQueries({
          queryKey: ["notifications", "unread-count"],
        });
      },
    });
  }

  return (
    <DropdownMenu placement="bottom-end">
      <DropdownMenuTrigger>
        <Button size="icon" variant="ghost" class="relative">
          <IoNotificationsOutline size={24} />
          <Show when={hasUnread()}>
            <span class="absolute top-1 right-1 flex h-4 w-4 items-center justify-center rounded-full bg-red-600 text-[10px] font-bold text-white">
              {unreadCount.data?.count}
            </span>
          </Show>
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent class="w-80 max-h-96 overflow-y-auto p-0">
        <div class="flex items-center justify-between px-3 py-2 border-b">
          <span class="text-sm font-semibold">Notifications</span>
          <Show when={hasUnread()}>
            <button
              class="text-xs text-blue-600 hover:underline"
              onClick={handleMarkAllRead}
              disabled={markAllRead.isPending}
            >
              Mark all as read
            </button>
          </Show>
        </div>
        <Show
          when={recentNotifications().length > 0}
          fallback={
            <div class="px-3 py-6 text-center text-sm text-gray-500">
              No notifications yet
            </div>
          }
        >
          <For each={recentNotifications()}>
            {(notification) => (
              <DropdownMenuItem
                class="flex flex-col items-start gap-0.5 px-3 py-2 cursor-pointer"
                onSelect={() => handleMarkRead(notification.uuid)}
              >
                <div class="flex w-full items-start gap-2">
                  <Show when={!notification.read_at}>
                    <span class="mt-1.5 h-2 w-2 shrink-0 rounded-full bg-blue-500" />
                  </Show>
                  <Show when={notification.read_at}>
                    <span class="mt-1.5 h-2 w-2 shrink-0" />
                  </Show>
                  <div class="min-w-0 flex-1">
                    <p class="text-sm font-semibold leading-tight truncate">
                      {notification.title}
                    </p>
                    <p class="text-xs text-gray-500 line-clamp-2 mt-0.5">
                      {notification.body}
                    </p>
                  </div>
                  <span class="shrink-0 text-[10px] text-gray-400 mt-0.5">
                    {formatRelativeTime(notification.created_at)}
                  </span>
                </div>
              </DropdownMenuItem>
            )}
          </For>
        </Show>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
