import type { GET, SandblastBackground } from "@glassact/data";

export function matchesBackgroundSearch(background: GET<SandblastBackground>, query: string): boolean {
  const needle = query.trim().toLowerCase();
  if (!needle) return true;
  return (
    background.name.toLowerCase().includes(needle) ||
    background.code.toLowerCase().includes(needle) ||
    background.tags.some((tag) => tag.includes(needle))
  );
}

export function backgroundTagOptions(backgrounds: GET<SandblastBackground>[]): string[] {
  return [...new Set(backgrounds.flatMap((background) => background.tags))].sort();
}
