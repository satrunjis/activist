import { describe, expect, it, vi } from "vitest";
import type { Edge, Node } from "@xyflow/react";

import { elkLayoutAdapter } from "./elkLayoutAdapter";
import type { DivisionViewportNodeData } from "../../tree/treeTypes";

function buildGraph(): {
  nodes: Array<Node<DivisionViewportNodeData>>;
  edges: Edge[];
} {
  const onSelect = vi.fn();

  return {
    nodes: [
      {
        id: "root",
        type: "divisionNode",
        position: { x: 0, y: 0 },
        data: {
          lod: "standard",
          depth: 0,
          onSelect,
          node: {
            id: "root",
            depth: 0,
            shortName: "Root",
            hasChildren: true,
            childrenCount: 2,
            isArchived: false,
            selected: true
          }
        }
      },
      {
        id: "a",
        type: "divisionNode",
        position: { x: 0, y: 0 },
        data: {
          lod: "standard",
          depth: 1,
          onSelect,
          node: {
            id: "a",
            parentId: "root",
            depth: 1,
            shortName: "A",
            hasChildren: false,
            childrenCount: 0,
            isArchived: false,
            selected: false
          }
        }
      }
    ],
    edges: [{ id: "edge-root-a", source: "root", target: "a" }]
  };
}

describe("elkLayoutAdapter", () => {
  it("returns positioned nodes and canvas dimensions", async () => {
    const graph = buildGraph();
    const result = await elkLayoutAdapter(graph.nodes, graph.edges);

    expect(result.nodes).toHaveLength(2);
    expect(result.canvasHeight).toBeGreaterThanOrEqual(460);
    for (const node of result.nodes) {
      expect(Number.isFinite(node.position.x)).toBe(true);
      expect(Number.isFinite(node.position.y)).toBe(true);
    }
  });

  it("preserves edge wiring and node ids", async () => {
    const graph = buildGraph();
    const result = await elkLayoutAdapter(graph.nodes, graph.edges);

    expect(result.edges).toEqual(graph.edges);
    expect(result.nodes.map((node) => node.id)).toEqual(["root", "a"]);
  });

  it("is deterministic for same input", async () => {
    const graph = buildGraph();
    const first = await elkLayoutAdapter(graph.nodes, graph.edges);
    const second = await elkLayoutAdapter(graph.nodes, graph.edges);

    expect(first.nodes.map((node) => node.position)).toEqual(second.nodes.map((node) => node.position));
  });
});
