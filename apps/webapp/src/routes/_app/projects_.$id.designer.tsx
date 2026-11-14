import { createFileRoute } from '@tanstack/solid-router'

export const Route = createFileRoute('/_app/projects_/$id/designer')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/_app/projects_/$id/designer"!</div>
}
