import { StandardTable } from "./helpers";

// Subscribes a single user to a project's notifications. A missing row means the
// user has never been subscribed; `is_watching: false` means they explicitly
// unwatched and later activity must not resubscribe them.
export type ProjectWatcher = StandardTable<{
  project_id: number;
  dealership_user_id: number | null;
  internal_user_id: number | null;
  is_watching: boolean;
}>;

export type ProjectWatcherUserType = "dealership" | "internal";

// How a watcher is shown to everyone else on the project. Email is deliberately
// absent so dealership users never see internal staff contact details.
export type ProjectWatcherSummary = {
  uuid: string;
  name: string;
  avatar: string;
  role: string;
  user_type: ProjectWatcherUserType;
};

export type SetProjectWatchResponse = {
  is_watching: boolean;
  watcher_count: number;
};
