import { Link } from "@tanstack/solid-router";

const linkClass = "-mb-px border-b-2 pb-2.5 text-sm font-medium transition-colors";
const activeClass = "border-crimson-light text-foreground";
const inactiveClass =
  "border-transparent text-muted-foreground hover:border-input hover:text-foreground";

export function CatalogSectionTabs() {
  return (
    <nav class="flex gap-6 border-b border-border">
      <Link
        to="/catalog"
        activeOptions={{ exact: true }}
        class={linkClass}
        activeProps={{ class: activeClass }}
        inactiveProps={{ class: inactiveClass }}
      >
        Inlays
      </Link>
      <Link
        to="/catalog/backgrounds"
        class={linkClass}
        activeProps={{ class: activeClass }}
        inactiveProps={{ class: inactiveClass }}
      >
        Sandblast backgrounds
      </Link>
    </nav>
  );
}
