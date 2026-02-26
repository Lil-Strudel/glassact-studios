import { Accessor, createEffect, createSignal } from "solid-js";

export function useDebounce<T>(
  signal: Accessor<T>,
  delayMs: number = 300,
): Accessor<T> {
  const [debouncedValue, setDebouncedValue] = createSignal<T>(signal());
  let timeoutId: ReturnType<typeof setTimeout> | null = null;
  let latestValue = signal();

  createEffect(() => {
    latestValue = signal();

    if (timeoutId) {
      clearTimeout(timeoutId);
    }

    timeoutId = setTimeout(() => {
      setDebouncedValue(() => latestValue);
    }, delayMs);
  });

  return debouncedValue;
}
