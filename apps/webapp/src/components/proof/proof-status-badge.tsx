import { Badge, type BadgeVariant } from "@glassact/ui";
import type { ProofStatus } from "@glassact/data";
import type { Component } from "solid-js";

interface ProofStatusBadgeProps {
  status: ProofStatus;
  class?: string;
}

const STATUS_CONFIG: Record<
  ProofStatus,
  { label: string; variant: BadgeVariant }
> = {
  pending: { label: "Pending Review", variant: "gilt" },
  approved: { label: "Approved", variant: "success" },
  declined: { label: "Declined", variant: "error" },
  superseded: { label: "Superseded", variant: "outline" },
};

export const ProofStatusBadge: Component<ProofStatusBadgeProps> = (props) => {
  const config = () => STATUS_CONFIG[props.status];

  return (
    <Badge variant={config().variant} lozenge class={props.class}>
      {config().label}
    </Badge>
  );
};
