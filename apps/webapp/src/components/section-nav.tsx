import { Link } from "@tanstack/solid-router";
import { For } from "solid-js";
import type { IconTypes } from "solid-icons";

export interface SectionNavItem {
  label: string;
  icon: IconTypes;
  path: string;
}

interface SectionNavProps {
  label: string;
  items: SectionNavItem[];
}

export function SectionNav(props: SectionNavProps) {
  return (
    <aside class="-mx-4 flex overflow-x-auto border-b border-border px-4 pb-3 sm:-mx-6 sm:px-6 lg:sticky lg:top-0 lg:mx-0 lg:block lg:w-56 lg:flex-none lg:self-start lg:border-0 lg:px-0 lg:pb-0">
      <nav class="flex-none" aria-label={props.label}>
        <p class="eyebrow mb-3 hidden px-3 lg:block">{props.label}</p>
        <ul
          role="list"
          class="flex gap-x-1 gap-y-0.5 whitespace-nowrap lg:flex-col"
        >
          <For each={props.items}>
            {(item) => (
              <li>
                <Link
                  to={item.path}
                  class="flex items-center gap-x-3 rounded-md py-2 pl-3 pr-4 text-sm font-medium transition-colors"
                  activeProps={{
                    class:
                      "bg-accent text-foreground shadow-[inset_2px_0_0_hsl(var(--crimson-light))]",
                  }}
                  inactiveProps={{
                    class:
                      "text-muted-foreground hover:bg-accent/60 hover:text-foreground",
                  }}
                >
                  <item.icon size={18} class="shrink-0" />
                  {item.label}
                </Link>
              </li>
            )}
          </For>
        </ul>
      </nav>
    </aside>
  );
}
