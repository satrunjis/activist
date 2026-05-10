import ELK from "elkjs/lib/elk.bundled.js";
import { Position, type Edge, type Node } from "@xyflow/react";

import type { DivisionViewportNodeData } from "../../tree/treeTypes";

const elk = new ELK();

export type ElkNodeSize = {
  width: number;
  height: number;
};

export type ElkLayoutConfig = {
  rankSep: number;
  nodeSep: number;
  getNodeSize: (node: Node<DivisionViewportNodeData>) => ElkNodeSize;
};

const DEFAULT_NODE_SIZE: ElkNodeSize = { width: 200, height: 72 };

const DEFAULT_LAYOUT_CONFIG: ElkLayoutConfig = {
  rankSep: 64,
  nodeSep: 24,
  getNodeSize: () => DEFAULT_NODE_SIZE
};

export type ElkLayoutResult = {
  nodes: Array<Node<DivisionViewportNodeData>>;
  edges: Edge[];
  canvasHeight: number;
};

export async function elkLayoutAdapter(
  nodes: Array<Node<DivisionViewportNodeData>>,
  edges: Edge[],
  config: ElkLayoutConfig = DEFAULT_LAYOUT_CONFIG
): Promise<ElkLayoutResult> {
  const nodeSizeById = new Map<string, ElkNodeSize>();
  for (const node of nodes) {
    nodeSizeById.set(node.id, config.getNodeSize(node));
  }

  const elkLayoutOptions = {
    "elk.algorithm":                              "layered",
    "elk.direction":                              "DOWN",
    "elk.layered.spacing.nodeNodeBetweenLayers":  String(config.rankSep),
    "elk.spacing.nodeNode":                       String(config.nodeSep),
    "elk.edgeRouting":                            "ORTHOGONAL",
    "elk.layered.nodePlacement.strategy":         "NETWORK_SIMPLEX"
  } as const;

  const layout = await elk.layout({
    id: "division-tree",
    layoutOptions: elkLayoutOptions,
    children: nodes.map((node) => ({
      id: node.id,
      width:  nodeSizeById.get(node.id)?.width ?? DEFAULT_NODE_SIZE.width,
      height: nodeSizeById.get(node.id)?.height ?? DEFAULT_NODE_SIZE.height
    })),
    edges: edges.map((edge) => ({
      id:      edge.id,
      sources: [edge.source],
      targets: [edge.target]
    }))
  });

  const positioned = new Map<string, { x: number; y: number }>();
  for (const child of layout.children ?? []) {
    positioned.set(child.id, {
      x: typeof child.x === "number" ? Math.round(child.x) : 0,
      y: typeof child.y === "number" ? Math.round(child.y) : 0
    });
  }

  let maxY = 0;
  const laidOutNodes = nodes.map((node) => {
    const pos = positioned.get(node.id);
    const nodeSize = nodeSizeById.get(node.id) ?? DEFAULT_NODE_SIZE;
    if (!pos) return node;
    maxY = Math.max(maxY, pos.y + nodeSize.height);
    return {
      ...node,
      targetPosition: Position.Top,
      sourcePosition: Position.Bottom,
      position: pos
    };
  });

  return {
    nodes: laidOutNodes,
    edges,
    canvasHeight: Math.max(600, Math.ceil(maxY + 160))
  };
}
