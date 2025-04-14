import { Button, Breadcrumb } from "@glassact/ui";
import { FiPlusCircle } from "solid-icons/fi";
import type { Component } from "solid-js";

const Projects: Component = () => {
  return (
    <div>
      <Breadcrumb crumbs={[{ title: "Projects", href: "/projects" }]} />
      <div class="flex justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
            Lorem Ipsum
          </h1>
          <p class="mt-2 text-sm text-gray-500">lorem ipsum domir alit sit</p>
        </div>
        <Button as="a" href="/projects/create-project">
          Create New Project
          <FiPlusCircle size={20} class="ml-2" />
        </Button>
      </div>
    </div>
  );
};

export default Projects;
