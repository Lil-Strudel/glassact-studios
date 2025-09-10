import { SolidMutationOptions } from "@tanstack/solid-query";

export function mutationOptions<K, J>(
  opts: {
    mutationFn: (arg: K) => Promise<J>;
  } & SolidMutationOptions<J, Error, K, unknown>,
): SolidMutationOptions<J, Error, K, unknown> {
  return opts;
}
