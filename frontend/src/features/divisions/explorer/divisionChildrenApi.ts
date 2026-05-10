import { request } from "../../../shared/api/client";
import type { DivisionChildItem, DivisionChildrenResponse } from "../../../shared/api/types";

/**
 * Fetch direct children of a division using the lazy-load endpoint.
 *
 * - parentId omitted or "root" => root-level divisions
 * - parentId set              => direct children of that division
 */
export async function fetchDivisionChildren(parentId?: string, signal?: AbortSignal): Promise<DivisionChildItem[]> {
  const param = parentId && parentId !== "root" ? `?parent_id=${encodeURIComponent(parentId)}` : "?parent_id=root";
  const response = await request<DivisionChildrenResponse>(`/api/v1/divisions${param}`, { signal });
  return Array.isArray(response.items) ? response.items : [];
}
