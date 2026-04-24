import { queryOptions } from "@tanstack/solid-query";
import api from "./api";
import type { DealershipDashboard, InternalDashboard } from "@glassact/data";

export async function getDealershipDashboard(): Promise<DealershipDashboard> {
  const res = await api.get("/dashboard/dealership");
  return res.data;
}

export function getDealershipDashboardOpts() {
  return queryOptions({
    queryKey: ["dashboard", "dealership"],
    queryFn: getDealershipDashboard,
  });
}

export async function getInternalDashboard(): Promise<InternalDashboard> {
  const res = await api.get("/dashboard/internal");
  return res.data;
}

export function getInternalDashboardOpts() {
  return queryOptions({
    queryKey: ["dashboard", "internal"],
    queryFn: getInternalDashboard,
  });
}
