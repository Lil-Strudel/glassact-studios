import { A, RouteSectionProps } from "@solidjs/router";
import { type Component, For } from "solid-js";
import { IoPersonOutline, IoBusinessOutline } from "solid-icons/io";

const navigationItems = [
  {
    label: "Dealerships",
    icon: IoBusinessOutline,
    path: "/admin/dealerships",
  },
  {
    label: "Users",
    icon: IoPersonOutline,
    path: "/admin/users",
  },
];

const AdminLayout: Component<RouteSectionProps<unknown>> = (props) => {
  return (
    <div>
      <div class="pt-16 lg:flex lg:gap-x-16">
        <aside class="flex overflow-x-auto border-b border-gray-900/5 py-4 lg:block lg:w-64 lg:flex-none lg:border-0 lg:py-20">
          <nav class="flex-none px-4 sm:px-6 lg:px-0">
            <ul
              role="list"
              class="flex gap-x-3 gap-y-1 whitespace-nowrap lg:flex-col"
            >
              <For each={navigationItems}>
                {(item) => {
                  return (
                    <li>
                      <A
                        href={item.path}
                        class="group flex gap-x-3 rounded-md py-2 pl-2 pr-3 text-sm/6 font-semibold"
                        activeClass="bg-gray-50 text-primary"
                        inactiveClass="text-gray-700 hover:bg-gray-50 hover:text-primary"
                      >
                        <item.icon size={24} />
                        {item.label}
                      </A>
                    </li>
                  );
                }}
              </For>
            </ul>
          </nav>
        </aside>

        <div class="mx-auto max-w-2xl space-y-16 sm:space-y-20 lg:mx-0 lg:max-w-none">
          {props.children}
        </div>
      </div>
    </div>
  );
};

export default AdminLayout;
