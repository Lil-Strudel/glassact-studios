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
import { Alert, buttonVariants, Card, CardContent } from "@glassact/ui";
import { IoAlertCircleOutline } from "solid-icons/io";
import {
  getSupportArticlesOpts,
  getSupportPriceGroupsOpts,
} from "../../queries/support";
import { Can } from "../../components/Can";
import { PriceCaveat } from "../../components/price-caveat";
import { InstallationKitNote } from "../../components/installation-kit-note";
import { Markdown } from "../../components/markdown";
import { YoutubeEmbed } from "../../components/youtube-embed";
import { ArticleFormDialog } from "../../components/support/article-form-dialog";
import { ArticleAdminControls } from "../../components/support/article-admin-controls";
import { PageHeader } from "@glassact/ui";
import { EmptyState } from "@glassact/ui";

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
    <div class="mx-auto flex max-w-4xl flex-col gap-10">
      <PageHeader
        eyebrow="Help"
        title="Support"
        description="Guides, tips, and answers for ordering and installing your inlays."
        actions={
          <Can permission={MANAGE_SUPPORT}>
            <ArticleFormDialog triggerClass={buttonVariants()}>
              + Add Article
            </ArticleFormDialog>
          </Can>
        }
      />

      <Switch>
        <Match when={articlesQuery.isLoading}>
          <div class="space-y-4">
            <For each={[0, 1, 2]}>
              {() => <div class="h-32 animate-pulse rounded-lg bg-card" />}
            </For>
          </div>
        </Match>
        <Match when={articlesQuery.isError}>
          <Alert variant="destructive">
            <IoAlertCircleOutline size={18} />
            <p>Failed to load support content. Please try again later.</p>
          </Alert>
        </Match>
        <Match when={articlesQuery.isSuccess}>
          <div class="space-y-12">
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
                      <div class="mb-5 space-y-1 border-b border-border pb-3">
                        <h2 class="text-2xl">{section.title}</h2>
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
              <EmptyState
                class="rounded-lg border border-dashed border-input"
                title="No support content has been added yet."
              />
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
          <h3 class="text-xl">{props.article.title}</h3>
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
          <h3 class="mb-3 text-xl">Price Groups</h3>
          <div class="overflow-x-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="border-b border-border text-left">
                  <th class="eyebrow py-2 pr-4">Group</th>
                  <th class="eyebrow py-2 pr-6 text-right">Base Price</th>
                  <th class="eyebrow py-2">Description</th>
                </tr>
              </thead>
              <tbody>
                <For each={props.priceGroups}>
                  {(group) => (
                    <tr class="border-b border-border last:border-0">
                      <td class="py-2 pr-4 font-medium text-foreground">
                        {group.name}
                      </td>
                      <td class="py-2 pr-6 text-right tabular-nums text-foreground">
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
          <PriceCaveat class="mt-3" />
          <InstallationKitNote class="mt-1" />
        </CardContent>
      </Card>
    </Show>
  );
}
