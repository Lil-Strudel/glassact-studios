import { Accessor, createEffect } from "solid-js";

// A delete, a narrower filter or a hand-edited URL can leave the page past the
// end of the list; wait for real data before judging, or every load would
// snap back to page 1.
export function useClampPage(options: {
  page: Accessor<number>;
  pageCount: Accessor<number>;
  hasSettled: Accessor<boolean>;
  setPage: (page: number) => void;
}) {
  createEffect(() => {
    if (!options.hasSettled()) return;
    const lastPage = Math.max(1, options.pageCount());
    if (options.page() > lastPage) options.setPage(lastPage);
  });
}
