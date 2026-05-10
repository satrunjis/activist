import { request } from "../../../shared/api/client";
import type { DivisionTreeNode } from "../../../shared/api/types";

function normalizeTree(node: DivisionTreeNode): DivisionTreeNode {
  return {
    ...node,
    children: Array.isArray(node.children) ? node.children.map(normalizeTree) : []
  };
}

export async function fetchFullDivisionTree(signal?: AbortSignal): Promise<DivisionTreeNode> {
  const response = await request<DivisionTreeNode>("/api/v1/divisions/tree?depth=99&include_archived=false", { signal });
  return normalizeTree(response);
}
