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
import { PageHeader } from "@glassact/ui";

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
    <div class="mx-auto max-w-3xl space-y-8">
      <PageHeader eyebrow="Your account" title="Settings" />

      <section class="space-y-5">
        <div class="space-y-2">
          <h2 class="text-2xl">Notification Preferences</h2>
          <p class="max-w-prose text-sm leading-relaxed text-muted-foreground">
            You receive notifications for projects you watch, plus anything that
            needs your role to act on it. Creating a project, placing an order,
            sending a message, or acting on a proof starts you watching
            automatically &mdash; use the Watch button on a project to change
            that.
          </p>
          <p class="max-w-prose text-sm leading-relaxed text-muted-foreground">
            These toggles control which of those notifications also reach you by
            email. Everything still appears in your notification bell.
          </p>
        </div>

        <Show
          when={!userQuery.isLoading && !prefsQuery.isLoading}
          fallback={<p class="text-sm text-faint">Loading preferences...</p>}
        >
          <div class="overflow-hidden rounded-lg border border-border bg-card">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-border bg-muted">
                  <th class="eyebrow px-4 py-3 text-left">Event</th>
                  <th class="eyebrow w-32 px-4 py-3 text-center">Email</th>
                </tr>
              </thead>
              <tbody>
                <For each={eventTypes()}>
                  {(eventType) => (
                    <tr class="border-b border-border transition-colors last:border-b-0 hover:bg-accent/60">
                      <td class="px-4 py-3 text-foreground">
                        {NOTIFICATION_EVENT_LABELS[eventType]}
                      </td>
                      <td class="px-4 py-3 text-center">
                        <input
                          type="checkbox"
                          class="h-4 w-4 cursor-pointer accent-primary"
                          aria-label={`Email me: ${NOTIFICATION_EVENT_LABELS[eventType]}`}
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
