import { Button } from "@glassact/ui";
import type { ParentProps } from "solid-js";
import { createSignal } from "solid-js";

export default function NavBar(props: ParentProps) {
  const [open, setOpen] = createSignal(false);

  const navigation = [
    { title: "Home", path: "/" },
    { title: "Gallery", path: "/gallery" },
    { title: "Our Product", path: "/product" },
    { title: "Catalog", path: "/catalog" },
    { title: "FAQs", path: "/faqs" },
    { title: "Customer Portal", path: "/login" },
  ];

  function toggleMenu() {
    setOpen(!open());
  }

  return (
    <nav class="bg-gray-100 w-full border-b lg:border-0 lg:static">
      <div class="items-center px-4 max-w-screen-xl mx-auto lg:flex lg:px-8">
        <div class="flex items-center justify-between py-3 lg:py-5 lg:block">
          <a href="/">{props.children}</a>
          <div class="lg:hidden">
            <Button
              variant="outline"
              size="icon"
              aria-label="Open Menu"
              onClick={() => toggleMenu()}
            >
              {open() ? (
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  class="h-6 w-6"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                >
                  <path
                    fill-rule="evenodd"
                    d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z"
                    clip-rule="evenodd"
                  />
                </svg>
              ) : (
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  class="h-6 w-6"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                >
                  <path
                    stroke-linecap="round"
                    stroke-linejoin="round"
                    stroke-width={2}
                    d="M4 8h16M4 16h16"
                  />
                </svg>
              )}
            </Button>
          </div>
        </div>
        <div
          class={`flex-1 justify-self-center pb-3 mt-8 lg:block lg:pb-0 lg:mt-0 ${
            open() ? "block" : "hidden"
          }`}
        >
          <ul class="justify-center items-center space-y-8 lg:flex lg:space-x-6 lg:space-y-0">
            {navigation.map((item) => (
              <li class="text-gray-600 hover:text-primary">
                <a href={item.path} class="block w-full">
                  <Button variant="ghost" class="text-md hover:text-primary">
                    {item.title}
                  </Button>
                </a>
              </li>
            ))}
            <li class="inline-block lg:hidden">
              <a href="/contact" class="block w-full">
                <Button variant="ghost" class="text-md hover:text-primary">
                  Contact Us
                </Button>
              </a>
            </li>
          </ul>
        </div>
        <div class="hidden lg:inline-block">
          <a href="/contact">
            <Button>Contact Us</Button>
          </a>
        </div>
      </div>
    </nav>
  );
}
