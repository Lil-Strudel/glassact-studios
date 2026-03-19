import { createFileRoute } from "@tanstack/solid-router";
import { For, Show } from "solid-js";
import { useMutation, useQuery, useQueryClient } from "@tanstack/solid-query";
import { getUserSelfOpts } from "../../queries/user";
import {
  getNotificationPreferencesOpts,
  patchNotificationPreferenceOpts,
} from "../../queries/notifications";
import {
  DEALERSHIP_NOTIFICATION_EVENT_TYPES,
  INTERNAL_NOTIFICATION_EVENT_TYPES,
  NOTIFICATION_EVENT_LABELS,
} from "@glassact/data";
import type {
  NotificationEventType,
  NotificationPreference,
} from "@glassact/data";

export const Route = createFileRoute("/_app/settings")({
  component: RouteComponent,
});

function RouteComponent() {
  const queryClient = useQueryClient();
  const userQuery = useQuery(() => getUserSelfOpts());
  const prefsQuery = useQuery(() => getNotificationPreferencesOpts());
  const patchPref = useMutation(() => patchNotificationPreferenceOpts());

  const eventTypes = () => {
    const u = userQuery.data;
    if (!u) return [];
    return "dealership_id" in u
      ? DEALERSHIP_NOTIFICATION_EVENT_TYPES
      : INTERNAL_NOTIFICATION_EVENT_TYPES;
  };

  function getPref(
    eventType: NotificationEventType,
  ): NotificationPreference | undefined {
    return prefsQuery.data?.find((p) => p.event_type === eventType);
  }

  function isEmailEnabled(eventType: NotificationEventType): boolean {
    const pref = getPref(eventType);
    return pref ? pref.email_enabled : true;
  }

  function handleToggle(eventType: NotificationEventType, checked: boolean) {
    patchPref.mutate(
      { eventType, body: { email_enabled: checked } },
      {
        onSuccess() {
          queryClient.invalidateQueries({
            queryKey: ["notification-preferences"],
          });
        },
      },
    );
  }

  return (
    <div class="space-y-8">
      <h1 class="text-2xl font-semibold">Settings</h1>

      <section class="space-y-4">
        <div>
          <h2 class="text-lg font-semibold">Notification Preferences</h2>
          <p class="text-sm text-gray-500 mt-1">
            Manage how you receive notifications for each event type.
          </p>
        </div>

        <Show
          when={!userQuery.isLoading && !prefsQuery.isLoading}
          fallback={<p class="text-sm text-gray-400">Loading preferences...</p>}
        >
          <div class="border rounded-lg overflow-hidden">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b bg-gray-50">
                  <th class="px-4 py-3 text-left font-medium text-gray-700">
                    Event
                  </th>
                  <th class="px-4 py-3 text-left font-medium text-gray-700 w-32">
                    Email
                  </th>
                </tr>
              </thead>
              <tbody>
                <For each={eventTypes()}>
                  {(eventType) => (
                    <tr class="border-b last:border-b-0 hover:bg-gray-50 transition-colors">
                      <td class="px-4 py-3 text-gray-800">
                        {NOTIFICATION_EVENT_LABELS[eventType]}
                      </td>
                      <td class="px-4 py-3">
                        <input
                          type="checkbox"
                          class="h-4 w-4 cursor-pointer accent-primary"
                          checked={isEmailEnabled(eventType)}
                          disabled={patchPref.isPending}
                          onChange={(e) =>
                            handleToggle(eventType, e.currentTarget.checked)
                          }
                        />
                      </td>
                    </tr>
                  )}
                </For>
              </tbody>
            </table>
          </div>
        </Show>
      </section>
    </div>
  );
}
