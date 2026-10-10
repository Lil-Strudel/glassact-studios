import { Accessor, createEffect, on } from "solid-js";

function scrollContainerTop(element: HTMLElement): number {
  for (
    let parent = element.parentElement;
    parent;
    parent = parent.parentElement
  ) {
    const { overflowY } = getComputedStyle(parent);
    if (overflowY === "auto" || overflowY === "scroll") {
      return parent.getBoundingClientRect().top;
    }
  }
  return 0;
}

// Only scrolls when the top of the list is out of view, so paging a list that
// already fits on screen doesn't push the filters above it off screen.
export function useScrollToTopOnChange(
  trigger: Accessor<unknown>,
  target: () => HTMLElement | undefined,
) {
  createEffect(
    on(
      trigger,
      () => {
        const element = target();
        if (!element) return;
        if (element.getBoundingClientRect().top < scrollContainerTop(element)) {
          element.scrollIntoView({ block: "start" });
        }
      },
      { defer: true },
    ),
  );
}
