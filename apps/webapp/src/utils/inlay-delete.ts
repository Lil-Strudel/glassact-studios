import type { InlayDeleteBlocker } from "@glassact/data";

// Explains an undeletable inlay in dealership terms: what already happened to
// it, and what to do instead. Mirrors inlayDeleteBlockedMessage in
// apps/api/modules/inlay/inlayHandlers.go — keep the two in sync.
export function inlayDeleteBlockedMessage(
  blockers: InlayDeleteBlocker[],
): string {
  if (blockers.includes("order")) {
    return "This inlay is part of an order you've already placed, so it can't be removed.";
  }

  return "This inlay is already in production, so it can't be removed.";
}
