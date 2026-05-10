import type { Edge, Node } from "@xyflow/react";

import { elkLayoutAdapter } from "../explorer/layout/elkLayoutAdapter";
import type { DivisionViewportNodeData } from "./treeTypes";

export type DivisionTreeLayoutResult = {
  nodes: Array<Node<DivisionViewportNodeData>>;
  edges: Edge[];
  canvasHeight: number;
};

type NodeDimension = { width: number; height: number };

const DEPTH_DIMENSIONS: Record<number, NodeDimension> = {
  0: { width: 200, height: 72 },
  1: { width: 176, height: 64 },
  2: { width: 156, height: 58 }
};

export function nodeDimensionsByDepth(depth: number): NodeDimension {
  return DEPTH_DIMENSIONS[depth] ?? { width: 140, height: 52 };
}

export async function divisionTreeLayout(
  nodes: Array<Node<DivisionViewportNodeData>>,
  edges: Edge[]
): Promise<DivisionTreeLayoutResult> {
  return elkLayoutAdapter(nodes, edges, {
    nodeSep: 24,
    rankSep: 64,
    getNodeSize: (node) => nodeDimensionsByDepth(node.data.depth)
  });
}
