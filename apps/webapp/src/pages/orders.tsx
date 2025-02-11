import { Button, Breadcrumb } from "@glassact/ui";
import { FiPlusCircle } from "solid-icons/fi";
import type { Component } from "solid-js";

const Orders: Component = () => {
  return (
    <div>
      <Breadcrumb crumbs={[{ title: "Orders", href: "/orders" }]} />
      <div class="flex justify-between">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
            Current Projects
          </h1>
          <p class="mt-2 text-sm text-gray-500">
            View current projects and take any action if needed.
          </p>
        </div>
        <Button as="a" href="/orders/place-order">
          Place New Order
          <FiPlusCircle size={20} class="ml-2" />
        </Button>
      </div>

      <div>
        <div>
          <table class="w-full text-gray-500 sm:mt-6">
            <thead class="sr-only text-left text-sm text-gray-500 sm:not-sr-only">
              <tr>
                <th scope="col" class="py-3 pr-8 font-normal sm:w-2/5 lg:w-1/3">
                  Product
                </th>
                <th
                  scope="col"
                  class="hidden w-1/5 py-3 pr-8 font-normal sm:table-cell"
                >
                  Price
                </th>
                <th
                  scope="col"
                  class="hidden py-3 pr-8 font-normal sm:table-cell"
                >
                  Status
                </th>
                <th scope="col" class="w-0 py-3 text-right font-normal">
                  Info
                </th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 border-b border-gray-200 text-sm sm:border-t">
              <tr>
                <td class="py-6 pr-8">
                  <div class="flex items-center">
                    <img
                      src="https://tailwindui.com/plus-assets/img/ecommerce-images/order-history-page-02-product-01.jpg"
                      alt="Detail of mechanical pencil tip with machined black steel shaft and chrome lead tip."
                      class="mr-6 size-16 rounded object-cover"
                    />
                    <div>
                      <div class="font-medium text-gray-900">
                        Machined Pen and Pencil Set
                      </div>
                      <div class="mt-1 sm:hidden">$70.00</div>
                    </div>
                  </div>
                </td>
                <td class="hidden py-6 pr-8 sm:table-cell">$70.00</td>
                <td class="hidden py-6 pr-8 sm:table-cell">
                  Delivered Jan 25, 2021
                </td>
                <td class="whitespace-nowrap py-6 text-right font-medium">
                  <Button variant="text">View Product</Button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
      <div class="mt-16">
        <h1 class="text-2xl font-bold tracking-tight text-gray-900 sm:text-3xl">
          Order history
        </h1>
        <p class="mt-2 text-sm text-gray-500">
          Check the status of recent orders, manage returns, and download
          invoices.
        </p>
      </div>

      <div class="mt-16">
        <div class="space-y-20">
          <div>
            <div class="rounded-lg bg-gray-50 px-4 py-6 sm:flex sm:items-center sm:justify-between sm:space-x-6 sm:px-6 lg:space-x-8">
              <dl class="flex-auto divide-y divide-gray-200 text-sm text-gray-600 sm:grid sm:grid-cols-3 sm:gap-x-6 sm:divide-y-0 lg:w-1/2 lg:flex-none lg:gap-x-8">
                <div class="max-sm:flex max-sm:justify-between max-sm:py-6 max-sm:first:pt-0 max-sm:last:pb-0">
                  <dt class="font-medium text-gray-900">Date placed</dt>
                  <dd class="sm:mt-1">
                    <time datetime="2021-01-22">January 22, 2021</time>
                  </dd>
                </div>
                <div class="max-sm:flex max-sm:justify-between max-sm:py-6 max-sm:first:pt-0 max-sm:last:pb-0">
                  <dt class="font-medium text-gray-900">Order number</dt>
                  <dd class="sm:mt-1">WU88191111</dd>
                </div>
                <div class="max-sm:flex max-sm:justify-between max-sm:py-6 max-sm:first:pt-0 max-sm:last:pb-0">
                  <dt class="font-medium text-gray-900">Total amount</dt>
                  <dd class="font-medium text-gray-900 sm:mt-1">$238.00</dd>
                </div>
              </dl>
              <Button variant="outline">View Invoice</Button>
            </div>

            <table class="mt-4 w-full text-gray-500 sm:mt-6">
              <thead class="sr-only text-left text-sm text-gray-500 sm:not-sr-only">
                <tr>
                  <th
                    scope="col"
                    class="py-3 pr-8 font-normal sm:w-2/5 lg:w-1/3"
                  >
                    Product
                  </th>
                  <th
                    scope="col"
                    class="hidden w-1/5 py-3 pr-8 font-normal sm:table-cell"
                  >
                    Price
                  </th>
                  <th
                    scope="col"
                    class="hidden py-3 pr-8 font-normal sm:table-cell"
                  >
                    Status
                  </th>
                  <th scope="col" class="w-0 py-3 text-right font-normal">
                    Info
                  </th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-200 border-b border-gray-200 text-sm sm:border-t">
                <tr>
                  <td class="py-6 pr-8">
                    <div class="flex items-center">
                      <img
                        src="https://tailwindui.com/plus-assets/img/ecommerce-images/order-history-page-02-product-01.jpg"
                        alt="Detail of mechanical pencil tip with machined black steel shaft and chrome lead tip."
                        class="mr-6 size-16 rounded object-cover"
                      />
                      <div>
                        <div class="font-medium text-gray-900">
                          Machined Pen and Pencil Set
                        </div>
                        <div class="mt-1 sm:hidden">$70.00</div>
                      </div>
                    </div>
                  </td>
                  <td class="hidden py-6 pr-8 sm:table-cell">$70.00</td>
                  <td class="hidden py-6 pr-8 sm:table-cell">
                    Delivered Jan 25, 2021
                  </td>
                  <td class="whitespace-nowrap py-6 text-right font-medium">
                    <Button variant="text">View Product</Button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Orders;
