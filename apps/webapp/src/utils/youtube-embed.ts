// Extracts a YouTube video id from the common URL shapes admins paste
// (watch?v=, youtu.be/, /embed/, /shorts/) and returns a nocookie embed URL.
// Returns null when no id can be found so callers can fall back to a plain link.
export function getYoutubeEmbedUrl(url: string): string | null {
  const id = getYoutubeVideoId(url);
  return id ? `https://www.youtube-nocookie.com/embed/${id}` : null;
}

export function getYoutubeVideoId(url: string): string | null {
  if (!url) return null;

  try {
    const parsed = new URL(url.trim());
    const host = parsed.hostname.replace(/^www\./, "");

    if (host === "youtu.be") {
      return sanitizeId(parsed.pathname.slice(1));
    }

    if (host.endsWith("youtube.com") || host.endsWith("youtube-nocookie.com")) {
      const v = parsed.searchParams.get("v");
      if (v) return sanitizeId(v);

      const segments = parsed.pathname.split("/").filter(Boolean);
      // /embed/{id}, /shorts/{id}, /live/{id}
      if (
        segments.length >= 2 &&
        ["embed", "shorts", "live"].includes(segments[0])
      ) {
        return sanitizeId(segments[1]);
      }
    }

    return null;
  } catch {
    return null;
  }
}

function sanitizeId(id: string): string | null {
  return /^[a-zA-Z0-9_-]{11}$/.test(id) ? id : null;
}
