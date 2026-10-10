import { Accessor, createEffect, createSignal, on } from "solid-js";
import { useDebounce } from "./use-debounce";

// The input owns its text and only the debounced value reaches the URL:
// binding the input to the URL directly drops keystrokes, because router
// navigation is async.
export function useDebouncedSearchInput(
  urlValue: Accessor<string>,
  commit: (value: string) => void,
  delayMs: number = 300,
) {
  const [text, setText] = createSignal(urlValue());
  const debouncedText = useDebounce(text, delayMs);

  createEffect(
    on(
      debouncedText,
      (value) => {
        if (value !== urlValue()) commit(value);
      },
      { defer: true },
    ),
  );

  return [text, setText] as const;
}
