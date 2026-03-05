import { Badge, cn } from "@glassact/ui";
import type { ProofStatus } from "@glassact/data";
import type { Component } from "solid-js";

interface ProofStatusBadgeProps {
  status: ProofStatus;
  class?: string;
}

const STATUS_CONFIG: Record<
  ProofStatus,
  { label: string; bg: string; text: string; border: string }
> = {
  pending: {
    label: "Pending Review",
    bg: "bg-yellow-100",
    text: "text-yellow-800",
    border: "border-yellow-300",
  },
  approved: {
    label: "Approved",
    bg: "bg-green-100",
    text: "text-green-800",
    border: "border-green-300",
  },
  declined: {
    label: "Declined",
    bg: "bg-red-100",
    text: "text-red-800",
    border: "border-red-300",
  },
  superseded: {
    label: "Superseded",
    bg: "bg-gray-100",
    text: "text-gray-600",
    border: "border-gray-300",
  },
};

export const ProofStatusBadge: Component<ProofStatusBadgeProps> = (props) => {
  const config = () => STATUS_CONFIG[props.status];

  return (
    <Badge
      variant="outline"
      class={cn(
        config().bg,
        config().text,
        config().border,
        "font-medium",
        props.class,
      )}
    >
      {config().label}
    </Badge>
  );
};
