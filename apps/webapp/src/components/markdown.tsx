import { createMemo } from "solid-js";
import { marked } from "marked";
import DOMPurify from "dompurify";
import { cn } from "@glassact/ui";

// Force every rendered link to open in a new, opener-isolated tab.
DOMPurify.addHook("afterSanitizeAttributes", (node) => {
  if (node.tagName === "A") {
    node.setAttribute("target", "_blank");
    node.setAttribute("rel", "noopener noreferrer");
  }
});

interface MarkdownProps {
  content: string;
  class?: string;
}

// Renders trusted-admin-authored Markdown. marked converts to HTML and DOMPurify
// strips anything unsafe (defense in depth even though authors are internal
// admins). Links are hardened to open in a new, opener-isolated tab.
export function Markdown(props: MarkdownProps) {
  const html = createMemo(() => {
    const raw = marked.parse(props.content ?? "", { async: false }) as string;
    return DOMPurify.sanitize(raw, {
      ADD_ATTR: ["target", "rel"],
    });
  });

  return (
    <div
      class={cn(
        "max-w-none text-sm leading-relaxed text-foreground [&_a]:text-primary [&_a]:underline [&_h1]:mt-4 [&_h1]:mb-2 [&_h1]:text-lg [&_h1]:font-semibold [&_h2]:mt-4 [&_h2]:mb-2 [&_h2]:text-base [&_h2]:font-semibold [&_h3]:mt-3 [&_h3]:mb-1 [&_h3]:font-semibold [&_ul]:my-2 [&_ul]:list-disc [&_ul]:pl-5 [&_ol]:my-2 [&_ol]:list-decimal [&_ol]:pl-5 [&_li]:my-0.5 [&_p]:my-2 [&_code]:rounded [&_code]:bg-muted [&_code]:px-1 [&_code]:py-0.5 [&_code]:text-xs [&_strong]:font-semibold",
        props.class,
      )}
      // eslint-disable-next-line solid/no-innerhtml -- content is sanitized with DOMPurify above
      innerHTML={html()}
    />
  );
}
