import { StandardTable } from "./helpers";

export type SupportCategory =
  | "installation"
  | "ordering"
  | "pricing"
  | "contact"
  | "general";

export const SUPPORT_CATEGORIES: SupportCategory[] = [
  "installation",
  "ordering",
  "pricing",
  "contact",
  "general",
];

export type SupportArticle = StandardTable<{
  category: SupportCategory;
  title: string;
  body: string;
  youtube_url: string | null;
  sort_order: number;
  is_published: boolean;
}>;
