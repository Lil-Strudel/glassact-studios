import type { ProjectStatus } from "@glassact/data";
import { Badge } from "@glassact/ui";
import {
  PROJECT_STATUS_LABELS,
  PROJECT_STATUS_TONES,
} from "../../utils/project-status";

interface ProjectStatusBadgeProps {
  status: ProjectStatus;
  class?: string;
}

export function ProjectStatusBadge(props: ProjectStatusBadgeProps) {
  return (
    <Badge
      variant={PROJECT_STATUS_TONES[props.status]}
      lozenge
      class={props.class}
    >
      {PROJECT_STATUS_LABELS[props.status]}
    </Badge>
  );
}
