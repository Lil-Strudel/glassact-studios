import { Link } from "@tanstack/solid-router";
import { Button } from "./button";
type breadcrumbProps = {
  crumbs: {
    to: string;
    title: string;
  }[];
};
export const Breadcrumb = (props: breadcrumbProps) => {
  return (
    <div class="flex gap-2 items-center">
      {props.crumbs.map((crumb, i) => (
        <>
          <Button as={Link} to={crumb.to} variant="link" class="px-0">
            {crumb.title}
          </Button>
          {i !== props.crumbs.length - 1 && (
            <svg
              fill="currentColor"
              stroke-width="0"
              xmlns="http://www.w3.org/2000/svg"
              viewBox="0 0 24 24"
              style="overflow: visible; color: currentcolor;"
              height="1em"
              width="1em"
            >
              <path
                fill="currentColor"
                d="m13.171 12-4.95-4.95 1.415-1.413L16 12l-6.364 6.364-1.414-1.415 4.95-4.95Z"
              ></path>
            </svg>
          )}
        </>
      ))}
    </div>
  );
};
