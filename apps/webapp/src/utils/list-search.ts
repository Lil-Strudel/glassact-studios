import { z } from "zod";
import { lastVisitedSearch } from "@glassact/ui";

// .default keeps the params optional for links and navigate; .catch turns a
// hand-mangled URL into page 1 instead of an error page.
export const listSearchSchema = z.object({
  page: z.number().int().min(1).default(1).catch(1),
  q: z.string().default("").catch(""),
});

export const listSearchDefaults = { page: 1, q: "" };

export type ListSearchInput = z.input<typeof listSearchSchema>;

export const adminCatalogSearchSchema = listSearchSchema.extend({
  inactive: z.boolean().default(false).catch(false),
});

export const adminCatalogSearchDefaults = {
  ...listSearchDefaults,
  inactive: false,
};

export type AdminCatalogSearchInput = z.input<typeof adminCatalogSearchSchema>;

export function lastAdminCatalogSearch() {
  return lastVisitedSearch<AdminCatalogSearchInput>("/admin/catalog");
}

export const catalogBrowseSearchSchema = listSearchSchema.extend({
  category: z.string().default("").catch(""),
});

export const catalogBrowseSearchDefaults = {
  ...listSearchDefaults,
  category: "",
};

export type CatalogBrowseSearchInput = z.input<
  typeof catalogBrowseSearchSchema
>;

export const TABLE_PAGE_SIZE = 10;
