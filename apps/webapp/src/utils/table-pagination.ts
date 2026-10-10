import {
  functionalUpdate,
  type OnChangeFn,
  type PaginationState,
} from "@tanstack/solid-table";
import type { Accessor } from "solid-js";
import { TABLE_PAGE_SIZE } from "./list-search";

export function tablePaginationState(page: number): PaginationState {
  return { pageIndex: page - 1, pageSize: TABLE_PAGE_SIZE };
}

export function onTablePaginationChange(
  page: Accessor<number>,
  setPage: (page: number) => void,
): OnChangeFn<PaginationState> {
  return (updater) => {
    const next = functionalUpdate(updater, tablePaginationState(page()));
    setPage(next.pageIndex + 1);
  };
}
