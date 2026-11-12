import { Link } from "@tanstack/solid-router";

const linkClass = "rounded-md px-3 py-1.5 text-sm font-medium";

export function CatalogSectionTabs() {
  return (
    <nav class="flex gap-1 rounded-lg bg-gray-100 p-1 self-start">
      <Link
        to="/catalog"
        activeOptions={{ exact: true }}
        class={linkClass}
        activeProps={{ class: "bg-white text-gray-900 shadow-sm" }}
        inactiveProps={{ class: "text-gray-600 hover:text-gray-900" }}
      >
        Inlays
      </Link>
      <Link
        to="/catalog/backgrounds"
        class={linkClass}
        activeProps={{ class: "bg-white text-gray-900 shadow-sm" }}
        inactiveProps={{ class: "text-gray-600 hover:text-gray-900" }}
      >
        Sandblast backgrounds
      </Link>
    </nav>
  );
}
