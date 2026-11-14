import { cn } from "./cn";
import { type ComponentProps, splitProps } from "solid-js";

export const Table = (props: ComponentProps<"table">) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <div class="w-full overflow-auto">
      <table
        class={cn("w-full caption-bottom text-sm", local.class)}
        {...rest}
      />
    </div>
  );
};

export const TableHeader = (props: ComponentProps<"thead">) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <thead
      class={cn(
        "bg-muted [&_tr]:border-b [&_tr]:border-border [&_tr:hover]:bg-muted",
        local.class,
      )}
      {...rest}
    />
  );
};

export const TableBody = (props: ComponentProps<"tbody">) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <tbody class={cn("[&_tr:last-child]:border-0", local.class)} {...rest} />
  );
};

export const TableFooter = (props: ComponentProps<"tfoot">) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <tbody
      class={cn("border-t border-border bg-muted font-medium text-foreground", local.class)}
      {...rest}
    />
  );
};

export const TableRow = (props: ComponentProps<"tr">) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <tr
      class={cn(
        "border-b border-border transition-colors hover:bg-accent/60 data-[state=selected]:bg-accent",
        local.class,
      )}
      {...rest}
    />
  );
};

export const TableHead = (props: ComponentProps<"th">) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <th
      class={cn(
        "h-10 px-3 text-left align-middle text-xs font-semibold uppercase tracking-label text-muted-foreground [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px]",
        local.class,
      )}
      {...rest}
    />
  );
};

export const TableCell = (props: ComponentProps<"td">) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <td
      class={cn(
        "px-3 py-2.5 align-middle [&:has([role=checkbox])]:pr-0 [&>[role=checkbox]]:translate-y-[2px]",
        local.class,
      )}
      {...rest}
    />
  );
};

export const TableCaption = (props: ComponentProps<"caption">) => {
  const [local, rest] = splitProps(props, ["class"]);

  return (
    <caption
      class={cn("mt-4 text-sm text-faint", local.class)}
      {...rest}
    />
  );
};
