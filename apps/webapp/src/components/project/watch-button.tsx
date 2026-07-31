import { createMemo, Show } from "solid-js";
import { useMutation, useQueryClient } from "@tanstack/solid-query";
import {
  Button,
  Tooltip,
  TooltipContent,
  TooltipTrigger,
  showToast,
} from "@glassact/ui";
import { IoEyeOutline, IoEyeOffOutline } from "solid-icons/io";
import { setProjectWatchOpts } from "../../queries/project";
import { isApiError } from "../../utils/is-api-error";

interface WatchButtonProps {
  projectUuid: string;
  isWatching: boolean;
  watcherCount: number;
}

export function WatchButton(props: WatchButtonProps) {
  const queryClient = useQueryClient();
  const setWatch = useMutation(() => setProjectWatchOpts());

  const label = createMemo(() => (props.isWatching ? "Watching" : "Watch"));

  const tooltip = createMemo(() =>
    props.isWatching
      ? "You receive notifications about this project. Click to stop."
      : "Get notified about proofs, manufacturing, and messages on this project.",
  );

  function toggleWatch() {
    setWatch.mutate(
      { uuid: props.projectUuid, isWatching: !props.isWatching },
      {
        onSuccess: () => {
          queryClient.invalidateQueries({
            queryKey: ["project", props.projectUuid],
          });
        },
        onError: (error) => {
          showToast({
            title: "Could not update watch state",
            description: isApiError(error)
              ? error.message
              : "An unexpected error occurred",
            variant: "destructive",
          });
        },
      },
    );
  }

  return (
    <Tooltip>
      <TooltipTrigger as="span">
        <Button
          variant={props.isWatching ? "secondary" : "outline"}
          onClick={toggleWatch}
          disabled={setWatch.isPending}
          class="gap-2"
          aria-label={`${label()} this project`}
        >
          <Show
            when={props.isWatching}
            fallback={<IoEyeOffOutline class="size-4" />}
          >
            <IoEyeOutline class="size-4" />
          </Show>
          {label()}
          <span class="text-xs opacity-70">{props.watcherCount}</span>
        </Button>
      </TooltipTrigger>
      <TooltipContent class="max-w-xs text-left">{tooltip()}</TooltipContent>
    </Tooltip>
  );
}
