import { createFileRoute, Link } from "@tanstack/solid-router";
import { GET, Inlay, Project, ProjectStatus } from "@glassact/data";
import { Button, Breadcrumb } from "@glassact/ui";
import { IoAddCircleOutline, IoCheckmarkCircleOutline } from "solid-icons/io";
import { Component, Index, Show } from "solid-js";
import { useQuery } from "@tanstack/solid-query";
import { getProjectsOpts } from "../../queries/project";
import { useUserContext } from "../../providers/user";

export const Route = createFileRoute("/_app/projects")({
  component: RouteComponent,
});

function RouteComponent() {
  const { user } = useUserContext();
  const query = useQuery(getProjectsOpts());

  function getByStatusi(statusi: ProjectStatus[]): GET<Project>[] {
    if (!query.isSuccess) return [];
    return query.data.filter((project) => statusi.includes(project.status));
  }

  return (
    <div>
      <Breadcrumb crumbs={[{ title: "Projects", to: "/projects" }]} />
      <div>
        <Button as={Link} to="/projects/create-project">
          Create New Project
          <IoAddCircleOutline size={20} class="ml-2" />
        </Button>
      </div>
      <div class="flex flex-col gap-16 mt-4">
        <div>
          <div>
            <h1 class="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
              Needs Action
            </h1>
            <p class="mt-2 text-sm text-gray-500">
              These projects have proofs that are awaiting your approval or
              invoices waiting to be paid.
            </p>
          </div>

          <div class="mt-4">
            <div class="space-y-4"></div>
          </div>
        </div>
      </div>
    </div>
  );
}

interface SectionMessageProps {
  title: string;
  description: string;
}
const SectionMessage: Component<SectionMessageProps> = (props) => {
  return (
    <div class="border-2 border-dashed border-gray-300 rounded-xl p-8">
      <div class="text-center">
        <div class="text-gray-400 text-lg font-medium">{props.title}</div>
        <div class="text-gray-400 text-sm mt-2">{props.description}</div>
      </div>
    </div>
  );
};

interface ProjectCardProps {
  project: () => GET<Project>;
}
const ProjectCard: Component<ProjectCardProps> = (props) => {
  return (
    <div class="border rounded-xl p-4">
      <div class="flex items-center justify-between">
        <span class="text-2xl font-bold">{props.project().name}</span>
        <Button as={Link} to={`/projects/${props.project().uuid}`}>
          View Proofs
        </Button>
      </div>

      <div class="mt-4 w-full">
        {/* <Show */}
        {/*   when={props.project().inlays.length > 0} */}
        {/*   fallback={ */}
        {/*     <SectionMessage */}
        {/*       title="No inlays found for this project" */}
        {/*       description="Please add one before you place an order" */}
        {/*     /> */}
        {/*   } */}
        {/* > */}
        {/*   <table class="w-full text-gray-500"> */}
        {/*     <thead class="text-left text-sm text-gray-500"> */}
        {/*       <tr> */}
        {/*         <th scope="col" class="py-3"> */}
        {/*           Inlay */}
        {/*         </th> */}
        {/*         <th scope="col" class="py-3 text-right"> */}
        {/*           Proof Status */}
        {/*         </th> */}
        {/*       </tr> */}
        {/*     </thead> */}
        {/*     <tbody class="divide-y divide-gray-200 border-y border-gray-200 text-sm"> */}
        {/*       <Index each={props.project().inlays}> */}
        {/*         {(inlay, index) => ( */}
        {/*           <tr> */}
        {/*             <td class="py-4"> */}
        {/*               <div class="flex items-center"> */}
        {/*                 <img */}
        {/*                   src={inlay().preview_url} */}
        {/*                   alt={`${inlay().name} Preview`} */}
        {/*                   class="mr-6 size-16 rounded object-cover" */}
        {/*                 /> */}
        {/*                 <div class="font-medium text-gray-900"> */}
        {/*                   {inlay().name} */}
        {/*                 </div> */}
        {/*               </div> */}
        {/*             </td> */}
        {/*             {index % 2 === 0 && ( */}
        {/*               <td class="text-right">Proof Awaiting Approval</td> */}
        {/*             )} */}
        {/*             {index % 2 === 1 && ( */}
        {/*               <td> */}
        {/*                 <div class="flex justify-end"> */}
        {/*                   Approved */}
        {/*                   <IoCheckmarkCircleOutline size={20} class="ml-2" /> */}
        {/*                 </div> */}
        {/*               </td> */}
        {/*             )} */}
        {/*           </tr> */}
        {/*         )} */}
        {/*       </Index> */}
        {/*     </tbody> */}
        {/*   </table> */}
        {/* </Show> */}
      </div>
    </div>
  );
};
