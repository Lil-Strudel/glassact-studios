import { createFileRoute } from "@tanstack/solid-router";
import { useQuery } from "@tanstack/solid-query";
import { createMemo, For, Match, Show, Switch } from "solid-js";
import {
  PERMISSION_ACTIONS,
  type GET,
  type PriceGroup,
  type SupportArticle,
  type SupportCategory,
} from "@glassact/data";
import { Card, CardContent } from "@glassact/ui";
import {
  getSupportArticlesOpts,
  getSupportPriceGroupsOpts,
} from "../../queries/support";
import { Can } from "../../components/Can";
import { Markdown } from "../../components/markdown";
import { YoutubeEmbed } from "../../components/youtube-embed";
import { ArticleFormDialog } from "../../components/support/article-form-dialog";
import { ArticleAdminControls } from "../../components/support/article-admin-controls";

export const Route = createFileRoute("/_app/support")({
  component: RouteComponent,
});

const MANAGE_SUPPORT = PERMISSION_ACTIONS.MANAGE_SUPPORT;

interface SectionDef {
  category: SupportCategory;
  title: string;
  description: string;
}

const SECTIONS: SectionDef[] = [
  {
    category: "installation",
    title: "Installation",
    description: "Videos and tips for installing your inlays.",
  },
  {
    category: "ordering",
    title: "Placing Orders",
    description: "How to build a project and place an order.",
  },
  {
    category: "pricing",
    title: "Pricing",
    description: "How pricing works and our current price groups.",
  },
  {
    category: "contact",
    title: "Contact Us",
    description: "Have another question? Here's how to reach us.",
  },
  {
    category: "general",
    title: "General",
    description: "Everything else.",
  },
];

const currencyFormatter = new Intl.NumberFormat("en-US", {
  style: "currency",
  currency: "USD",
});

function formatCents(cents: number): string {
  return currencyFormatter.format(cents / 100);
}

function RouteComponent() {
  const articlesQuery = useQuery(() => getSupportArticlesOpts());
  const priceGroupsQuery = useQuery(() => getSupportPriceGroupsOpts());

  const articlesByCategory = createMemo(() => {
    const grouped = new Map<SupportCategory, GET<SupportArticle>[]>();
    for (const article of articlesQuery.data ?? []) {
      const list = grouped.get(article.category) ?? [];
      list.push(article);
      grouped.set(article.category, list);
    }
    return grouped;
  });

  return (
    <div class="mx-auto max-w-4xl px-4 py-8">
      <div class="mb-8 flex items-start justify-between gap-4">
        <div>
          <h1 class="text-2xl font-semibold text-foreground">Support</h1>
          <p class="mt-1 text-sm text-muted-foreground">
            Guides, tips, and answers for ordering and installing your inlays.
          </p>
        </div>
        <Can permission={MANAGE_SUPPORT}>
          <ArticleFormDialog triggerClass="inline-flex h-9 items-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground shadow hover:bg-primary/90">
            + Add Article
          </ArticleFormDialog>
        </Can>
      </div>

      <Switch>
        <Match when={articlesQuery.isLoading}>
          <div class="space-y-4">
            <For each={[0, 1, 2]}>
              {() => <div class="h-32 animate-pulse rounded-lg bg-muted" />}
            </For>
          </div>
        </Match>
        <Match when={articlesQuery.isError}>
          <div class="rounded-lg border border-destructive/40 bg-destructive/5 p-4 text-sm text-destructive">
            Failed to load support content. Please try again later.
          </div>
        </Match>
        <Match when={articlesQuery.isSuccess}>
          <div class="space-y-10">
            <For each={SECTIONS}>
              {(section) => {
                const articles = () =>
                  articlesByCategory().get(section.category) ?? [];
                const isPricing = section.category === "pricing";
                const hasContent = () =>
                  articles().length > 0 ||
                  (isPricing && (priceGroupsQuery.data?.length ?? 0) > 0);

                return (
                  <Show when={hasContent()}>
                    <section>
                      <div class="mb-4 border-b pb-2">
                        <h2 class="text-lg font-semibold text-foreground">
                          {section.title}
                        </h2>
                        <p class="text-sm text-muted-foreground">
                          {section.description}
                        </p>
                      </div>

                      <div class="space-y-4">
                        <For each={articles()}>
                          {(article) => <ArticleCard article={article} />}
                        </For>

                        <Show when={isPricing}>
                          <PriceGroupsCard
                            priceGroups={priceGroupsQuery.data ?? []}
                          />
                        </Show>
                      </div>
                    </section>
                  </Show>
                );
              }}
            </For>

            <Show
              when={
                (articlesQuery.data?.length ?? 0) === 0 &&
                (priceGroupsQuery.data?.length ?? 0) === 0
              }
            >
              <div class="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">
                No support content has been added yet.
              </div>
            </Show>
          </div>
        </Match>
      </Switch>
    </div>
  );
}

interface ArticleCardProps {
  article: GET<SupportArticle>;
}

function ArticleCard(props: ArticleCardProps) {
  return (
    <Card>
      <CardContent class="space-y-3 p-5">
        <div class="flex items-start justify-between gap-4">
          <h3 class="text-base font-semibold text-foreground">
            {props.article.title}
          </h3>
          <Can permission={MANAGE_SUPPORT}>
            <ArticleAdminControls article={props.article} />
          </Can>
        </div>

        <Show when={props.article.youtube_url}>
          {(url) => <YoutubeEmbed url={url()} title={props.article.title} />}
        </Show>

        <Show when={props.article.body.trim().length > 0}>
          <Markdown content={props.article.body} />
        </Show>
      </CardContent>
    </Card>
  );
}

interface PriceGroupsCardProps {
  priceGroups: GET<PriceGroup>[];
}

function PriceGroupsCard(props: PriceGroupsCardProps) {
  return (
    <Show when={props.priceGroups.length > 0}>
      <Card>
        <CardContent class="p-5">
          <h3 class="mb-3 text-base font-semibold text-foreground">
            Price Groups
          </h3>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b text-left text-muted-foreground">
                  <th class="py-2 pr-4 font-medium">Group</th>
                  <th class="py-2 pr-4 font-medium">Base Price</th>
                  <th class="py-2 font-medium">Description</th>
                </tr>
              </thead>
              <tbody>
                <For each={props.priceGroups}>
                  {(group) => (
                    <tr class="border-b last:border-0">
                      <td class="py-2 pr-4 font-medium text-foreground">
                        {group.name}
                      </td>
                      <td class="py-2 pr-4 text-foreground">
                        {formatCents(group.base_price_cents)}
                      </td>
                      <td class="py-2 text-muted-foreground">
                        {group.description ?? "—"}
                      </td>
                    </tr>
                  )}
                </For>
              </tbody>
            </table>
          </div>
        </CardContent>
      </Card>
    </Show>
  );
}
