import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { GET, Invoice } from "@glassact/data";
import { mutationOptions } from "../utils/mutation-options";

export async function getProjectInvoice(
  projectUuid: string,
): Promise<GET<Invoice>> {
  const res = await api.get(`/project/${projectUuid}/invoice`);
  return res.data;
}

export function getProjectInvoiceOpts(projectUuid: string) {
  return queryOptions({
    queryKey: ["project", projectUuid, "invoice"],
    queryFn: () => getProjectInvoice(projectUuid),
  });
}

export async function getInvoice(uuid: string): Promise<GET<Invoice>> {
  const res = await api.get(`/invoice/${uuid}`);
  return res.data;
}

export function getInvoiceOpts(uuid: string) {
  return queryOptions({
    queryKey: ["invoice", uuid],
    queryFn: () => getInvoice(uuid),
  });
}

export interface AttachInvoiceRequest {
  projectUuid: string;
  invoiceUrl: string;
}

export async function postProjectInvoice(
  request: AttachInvoiceRequest,
): Promise<GET<Invoice>> {
  const res = await api.post(`/project/${request.projectUuid}/invoice`, {
    invoice_url: request.invoiceUrl,
  });
  return res.data;
}

export function postProjectInvoiceOpts() {
  return mutationOptions({
    mutationFn: postProjectInvoice,
  });
}

export async function postMarkInvoicePaid(
  invoiceUuid: string,
): Promise<GET<Invoice>> {
  const res = await api.post(`/invoice/${invoiceUuid}/mark-paid`);
  return res.data;
}

export function postMarkInvoicePaidOpts() {
  return mutationOptions({
    mutationFn: postMarkInvoicePaid,
  });
}

export async function postVoidInvoice(
  invoiceUuid: string,
): Promise<GET<Invoice>> {
  const res = await api.post(`/invoice/${invoiceUuid}/void`);
  return res.data;
}

export function postVoidInvoiceOpts() {
  return mutationOptions({
    mutationFn: postVoidInvoice,
  });
}
