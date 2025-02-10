import { Button } from "@glassact/ui";
import type { Component } from "solid-js";

const PlaceOrder: Component = () => {
  return (
    <div>
      <div class="mx-auto max-w-2xl px-4 py-16 sm:px-6 sm:py-24 lg:px-0">
        <h1 class="text-center text-3xl font-bold tracking-tight text-gray-900 sm:text-4xl">
          New Order
        </h1>

        <div class="mt-12">
          <section>
            <ul
              role="list"
              class="divide-y divide-gray-200 border-b border-t border-gray-200"
            >
              <li class="flex py-6">
                <div class="shrink-0">
                  <img
                    src="https://tailwindui.com/plus-assets/img/ecommerce-images/checkout-page-03-product-04.jpg"
                    alt="Front side of mint cotton t-shirt with wavey lines pattern."
                    class="size-24 rounded-md object-cover sm:size-32"
                  />
                </div>

                <div class="ml-4 flex flex-1 flex-col sm:ml-6">
                  <div>
                    <div class="flex justify-between">
                      <h4 class="text-sm">
                        <a
                          href="#"
                          class="font-medium text-gray-700 hover:text-gray-800"
                        >
                          Artwork Tee
                        </a>
                      </h4>
                      <p class="ml-4 text-sm font-medium text-gray-900">
                        $32.00
                      </p>
                    </div>
                    <p class="mt-1 text-sm text-gray-500">Mint</p>
                    <p class="mt-1 text-sm text-gray-500">Medium</p>
                  </div>

                  <div class="mt-4 flex flex-1 items-end justify-between">
                    <p class="flex items-center space-x-2 text-sm text-gray-700">
                      <svg
                        class="size-5 shrink-0 text-green-500"
                        viewBox="0 0 20 20"
                        fill="currentColor"
                        data-slot="icon"
                      >
                        <path
                          fill-rule="evenodd"
                          d="M16.704 4.153a.75.75 0 0 1 .143 1.052l-8 10.5a.75.75 0 0 1-1.127.075l-4.5-4.5a.75.75 0 0 1 1.06-1.06l3.894 3.893 7.48-9.817a.75.75 0 0 1 1.05-.143Z"
                          clip-rule="evenodd"
                        />
                      </svg>
                      <span>In stock</span>
                    </p>
                    <div class="ml-4">
                      <Button variant="text" class="p-0">
                        Remove
                      </Button>
                    </div>
                  </div>
                </div>
              </li>
              <li class="flex py-6">
                <div class="shrink-0">
                  <img
                    src="https://tailwindui.com/plus-assets/img/ecommerce-images/shopping-cart-page-01-product-02.jpg"
                    alt="Front side of charcoal cotton t-shirt."
                    class="size-24 rounded-md object-cover sm:size-32"
                  />
                </div>

                <div class="ml-4 flex flex-1 flex-col sm:ml-6">
                  <div>
                    <div class="flex justify-between">
                      <h4 class="text-sm">
                        <a
                          href="#"
                          class="font-medium text-gray-700 hover:text-gray-800"
                        >
                          Basic Tee
                        </a>
                      </h4>
                      <p class="ml-4 text-sm font-medium text-gray-900">
                        $32.00
                      </p>
                    </div>
                    <p class="mt-1 text-sm text-gray-500">Charcoal</p>
                    <p class="mt-1 text-sm text-gray-500">Large</p>
                  </div>

                  <div class="mt-4 flex flex-1 items-end justify-between">
                    <p class="flex items-center space-x-2 text-sm text-gray-700">
                      <svg
                        class="size-5 shrink-0 text-gray-300"
                        viewBox="0 0 20 20"
                        fill="currentColor"
                        data-slot="icon"
                      >
                        <path
                          fill-rule="evenodd"
                          d="M10 18a8 8 0 1 0 0-16 8 8 0 0 0 0 16Zm.75-13a.75.75 0 0 0-1.5 0v5c0 .414.336.75.75.75h4a.75.75 0 0 0 0-1.5h-3.25V5Z"
                          clip-rule="evenodd"
                        />
                      </svg>
                      <span>Will ship in 7-8 years</span>
                    </p>
                    <div class="ml-4">
                      <Button variant="text" class="p-0">
                        Remove
                      </Button>
                    </div>
                  </div>
                </div>
              </li>
            </ul>
          </section>
          <section class="grid place-items-center mt-10">
            <Button
              as="a"
              href="/orders/place-order/add-item"
              variant="outline"
            >
              Add Additional Item
            </Button>
          </section>

          <section class="mt-10">
            <div>
              <dl class="space-y-4">
                <div class="flex items-center justify-between">
                  <dt class="text-base font-medium text-gray-900">
                    Estimated Total
                  </dt>
                  <dd class="ml-4 text-base font-medium text-gray-900">
                    $96.00
                  </dd>
                </div>
              </dl>
              <p class="mt-1 text-sm text-gray-500">
                Shipping and taxes will be calculated at time of shipment.
              </p>
            </div>

            <div class="mt-10">
              <Button class="w-full">Place Order</Button>
            </div>
          </section>
        </div>
      </div>
    </div>
  );
};

export default PlaceOrder;
