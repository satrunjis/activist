import { request } from "../../shared/api/client";
import type { PositionMembersResponse } from "../../shared/api/types";

export function getPositionMembers(positionId: string): Promise<PositionMembersResponse> {
  return request<PositionMembersResponse>(`/api/v1/positions/${encodeURIComponent(positionId)}/members`);
}
