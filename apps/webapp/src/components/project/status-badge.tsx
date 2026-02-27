import type { ProjectStatus } from "@glassact/data";
import { Badge } from "@glassact/ui";

const STATUS_CONFIG: Record<
  ProjectStatus,
  { label: string; class: string }
> = {
  draft: {
    label: "Draft",
    class: "bg-gray-100 text-gray-700 border-gray-300",
  },
  designing: {
    label: "Designing",
    class: "bg-blue-100 text-blue-700 border-blue-300",
  },
  "pending-approval": {
    label: "Pending Approval",
    class: "bg-yellow-100 text-yellow-700 border-yellow-300",
  },
  approved: {
    label: "Approved",
    class: "bg-green-100 text-green-700 border-green-300",
  },
  ordered: {
    label: "Ordered",
    class: "bg-indigo-100 text-indigo-700 border-indigo-300",
  },
  "in-production": {
    label: "In Production",
    class: "bg-purple-100 text-purple-700 border-purple-300",
  },
  shipped: {
    label: "Shipped",
    class: "bg-cyan-100 text-cyan-700 border-cyan-300",
  },
  delivered: {
    label: "Delivered",
    class: "bg-teal-100 text-teal-700 border-teal-300",
  },
  invoiced: {
    label: "Invoiced",
    class: "bg-orange-100 text-orange-700 border-orange-300",
  },
  completed: {
    label: "Completed",
    class: "bg-green-100 text-green-700 border-green-300",
  },
  cancelled: {
    label: "Cancelled",
    class: "bg-red-100 text-red-700 border-red-300",
  },
};

interface ProjectStatusBadgeProps {
  status: ProjectStatus;
}

export function ProjectStatusBadge(props: ProjectStatusBadgeProps) {
  const config = () => STATUS_CONFIG[props.status];

  return (
    <Badge variant="outline" class={config().class}>
      {config().label}
    </Badge>
  );
}
