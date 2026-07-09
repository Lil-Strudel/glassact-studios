import { createMemo, Show } from "solid-js";
import { getYoutubeEmbedUrl } from "../utils/youtube-embed";

interface YoutubeEmbedProps {
  url: string;
  title?: string;
}

export function YoutubeEmbed(props: YoutubeEmbedProps) {
  const embedUrl = createMemo(() => getYoutubeEmbedUrl(props.url));

  return (
    <Show
      when={embedUrl()}
      fallback={
        <a
          href={props.url}
          target="_blank"
          rel="noopener noreferrer"
          class="text-primary underline"
        >
          Watch video
        </a>
      }
    >
      {(src) => (
        <div class="relative w-full overflow-hidden rounded-lg border bg-muted pt-[56.25%]">
          <iframe
            src={src()}
            title={props.title ?? "YouTube video"}
            class="absolute inset-0 h-full w-full"
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
            allowfullscreen
            loading="lazy"
          />
        </div>
      )}
    </Show>
  );
}
