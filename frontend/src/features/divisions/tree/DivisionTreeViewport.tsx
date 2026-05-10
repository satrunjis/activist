import { memo, useEffect, useMemo, useRef, useState } from "react";
import type { CSSProperties, ReactNode } from "react";
import type { Edge, Node, NodeProps, NodeTypes, Viewport } from "@xyflow/react";
import {
  Background,
  BackgroundVariant,
  Handle,
  Panel,
  Position,
  ReactFlow,
  useReactFlow
} from "@xyflow/react";

import { zoomLod } from "../explorer/lod/zoomLod";
import { divisionTreeLayout, nodeDimensionsByDepth } from "./divisionTreeLayout";
import type {
  DivisionViewportNode,
  DivisionViewportNodeData,
  DivisionZoomLodTier
} from "./treeTypes";

import "@xyflow/react/dist/style.css";

const HIDDEN_HANDLE_STYLE: CSSProperties = {
  width: 0,
  height: 0,
  opacity: 0,
  pointerEvents: "none",
  background: "transparent",
  border: "none"
};

const SINGLE_LINE_ELLIPSIS: CSSProperties = {
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap",
  minWidth: 0
};

type DepthMetrics = {
  width: number;
  height: number;
  scale: number;
};

const DEPTH_SCALE: Record<number, number> = {
  0: 1.0,
  1: 0.92,
  2: 0.85
};

function nodeDepthMetrics(depth: number): DepthMetrics {
  const dimensions = nodeDimensionsByDepth(depth);
  return {
    width: dimensions.width,
    height: dimensions.height,
    scale: DEPTH_SCALE[depth] ?? 0.8
  };
}

function nodeRootStyle(node: DivisionViewportNode, metrics: DepthMetrics, hovered: boolean): CSSProperties {
  return {
    width: `${metrics.width}px`,
    height: `${metrics.height}px`,
    fontSize: `${metrics.scale * 100}%`,
    background: node.isArchived ? "var(--color-node-archived-bg)" : "var(--color-node-bg)",
    border: "2px solid var(--color-node-border)",
    borderColor: node.selected
      ? "var(--color-node-border-selected)"
      : hovered
        ? "var(--color-node-border-hover)"
        : "var(--color-node-border)",
    borderRadius: "var(--radius-card)",
    boxShadow: node.selected ? "var(--color-node-shadow-selected)" : "none",
    color: node.isArchived ? "var(--color-node-archived-text)" : "var(--color-text-primary)",
    cursor: "pointer",
    padding: 0,
    textAlign: "left",
    transition: "var(--transition-fast)",
    fontFamily: "var(--font-body)",
    overflow: "hidden",
    display: "flex",
    outline: "none",
    appearance: "none"
  };
}

type NodeShellProps = {
  node: DivisionViewportNode;
  depth: number;
  onSelect: (id: string) => void;
  contentStyle?: CSSProperties;
  children: ReactNode;
};

function NodeShell({ node, depth, onSelect, children, contentStyle }: NodeShellProps) {
  const [hovered, setHovered] = useState(false);
  const metrics = nodeDepthMetrics(depth);

  return (
    <button
      className="nodrag nopan"
      data-testid={`division-node-${node.id}`}
      onClick={() => onSelect(node.id)}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={nodeRootStyle(node, metrics, hovered)}
      type="button"
    >
      <div style={{ width: "100%", height: "100%", ...contentStyle }}>
        {children}
      </div>
    </button>
  );
}

function MicroNode({ node, depth, onSelect }: { node: DivisionViewportNode; depth: number; onSelect: (id: string) => void }) {
  return (
    <NodeShell
      contentStyle={{ display: "flex", alignItems: "stretch", justifyContent: "flex-start" }}
      depth={depth}
      node={node}
      onSelect={onSelect}
    >
      <div
        aria-hidden
        style={{
          width: 4,
          height: "100%",
          background: "var(--color-brand-primary)"
        }}
      />
    </NodeShell>
  );
}

function CompactNode({ node, depth, onSelect }: { node: DivisionViewportNode; depth: number; onSelect: (id: string) => void }) {
  const { scale } = nodeDepthMetrics(depth);
  return (
    <NodeShell
      contentStyle={{
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        padding: "var(--space-2) var(--space-3)"
      }}
      depth={depth}
      node={node}
      onSelect={onSelect}
    >
      <span style={{ ...SINGLE_LINE_ELLIPSIS, width: "100%", textAlign: "center", fontWeight: "var(--weight-semibold)", fontSize: `${12 * scale}px` }}>
        {node.shortName}
      </span>
    </NodeShell>
  );
}

function StandardNode({ node, depth, onSelect }: { node: DivisionViewportNode; depth: number; onSelect: (id: string) => void }) {
  const { scale } = nodeDepthMetrics(depth);
  const positionsCount = node.positionsCount ?? 0;
  const membersCount = node.membersCount ?? 0;

  return (
    <NodeShell
      contentStyle={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        padding: "var(--space-2) var(--space-3)"
      }}
      depth={depth}
      node={node}
      onSelect={onSelect}
    >
      <div style={{ ...SINGLE_LINE_ELLIPSIS, fontWeight: "var(--weight-semibold)", fontSize: `${11 * scale}px` }}>
        {node.shortName}
      </div>
      <div
        style={{
          marginTop: "auto",
          display: "flex",
          justifyContent: "space-between",
          gap: "var(--space-2)",
          fontSize: `${10 * scale}px`,
          color: "var(--color-text-muted)"
        }}
      >
        <span style={SINGLE_LINE_ELLIPSIS}>{positionsCount} должн.</span>
        <span style={SINGLE_LINE_ELLIPSIS}>{membersCount} уч.</span>
      </div>
    </NodeShell>
  );
}

function DetailNode({ node, depth, onSelect }: { node: DivisionViewportNode; depth: number; onSelect: (id: string) => void }) {
  const { scale } = nodeDepthMetrics(depth);
  const positionsCount = node.positionsCount ?? 0;
  const membersCount = node.membersCount ?? 0;

  return (
    <NodeShell
      contentStyle={{
        display: "flex",
        flexDirection: "column",
        height: "100%",
        padding: "var(--space-2) var(--space-3)"
      }}
      depth={depth}
      node={node}
      onSelect={onSelect}
    >
      <div style={{ ...SINGLE_LINE_ELLIPSIS, fontWeight: "var(--weight-semibold)", color: "var(--color-text-muted)", fontSize: `${10 * scale}px` }}>
        {node.shortName}
      </div>
      <div style={{ ...SINGLE_LINE_ELLIPSIS, fontWeight: "var(--weight-bold)", fontSize: `${11 * scale}px`, marginTop: "2px" }}>
        {node.fullName ?? node.shortName}
      </div>
      <div
        style={{
          marginTop: "auto",
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
          gap: "var(--space-2)",
          fontSize: `${10 * scale}px`,
          color: "var(--color-text-muted)"
        }}
      >
        {node.isArchived ? (
          <span style={{ ...SINGLE_LINE_ELLIPSIS, display: "inline-flex", alignItems: "center", gap: 6 }}>
            <span
              aria-hidden
              style={{
                width: 6,
                height: 6,
                borderRadius: "var(--radius-full)",
                background: "var(--color-warning)",
                flexShrink: 0
              }}
            />
            Архив
          </span>
        ) : (
          <span />
        )}
        <span style={{ ...SINGLE_LINE_ELLIPSIS, textAlign: "right" }}>
          {positionsCount} должн. · {membersCount} уч.
        </span>
      </div>
    </NodeShell>
  );
}

function DivisionFlowNode({ data }: NodeProps<Node<DivisionViewportNodeData>>) {
  if (data.lod === "micro") {
    return (
      <>
        <Handle position={Position.Top} style={HIDDEN_HANDLE_STYLE} type="target" />
        <MicroNode depth={data.depth} node={data.node} onSelect={data.onSelect} />
        <Handle position={Position.Bottom} style={HIDDEN_HANDLE_STYLE} type="source" />
      </>
    );
  }

  if (data.lod === "compact") {
    return (
      <>
        <Handle position={Position.Top} style={HIDDEN_HANDLE_STYLE} type="target" />
        <CompactNode depth={data.depth} node={data.node} onSelect={data.onSelect} />
        <Handle position={Position.Bottom} style={HIDDEN_HANDLE_STYLE} type="source" />
      </>
    );
  }

  if (data.lod === "detail") {
    return (
      <>
        <Handle position={Position.Top} style={HIDDEN_HANDLE_STYLE} type="target" />
        <DetailNode depth={data.depth} node={data.node} onSelect={data.onSelect} />
        <Handle position={Position.Bottom} style={HIDDEN_HANDLE_STYLE} type="source" />
      </>
    );
  }

  return (
    <>
      <Handle position={Position.Top} style={HIDDEN_HANDLE_STYLE} type="target" />
      <StandardNode depth={data.depth} node={data.node} onSelect={data.onSelect} />
      <Handle position={Position.Bottom} style={HIDDEN_HANDLE_STYLE} type="source" />
    </>
  );
}

const nodeTypes: NodeTypes = { divisionNode: memo(DivisionFlowNode) as NodeTypes[string] };

const defaultEdgeOptions = {
  type: "bezier",
  style: {
    stroke: "#B0BAC4",
    strokeWidth: 1.5,
    pointerEvents: "none"
  },
  interactionWidth: 0
} as const;

const NOOP_SELECT = () => undefined;

type DivisionLayoutDescriptor = {
  id: string;
  parentId?: string;
  depth: number;
};

function toLayoutPlaceholderNode(descriptor: DivisionLayoutDescriptor): DivisionViewportNode {
  return {
    id: descriptor.id,
    parentId: descriptor.parentId,
    depth: descriptor.depth,
    shortName: descriptor.id,
    fullName: descriptor.id,
    description: "",
    isArchived: false,
    hasChildren: false,
    childrenCount: 0,
    selected: false
  };
}

function buildLayoutFlow(descriptors: DivisionLayoutDescriptor[]): {
  nodes: Array<Node<DivisionViewportNodeData>>;
  edges: Edge[];
} {
  const flowNodes: Array<Node<DivisionViewportNodeData>> = descriptors.map((descriptor) => ({
    id: descriptor.id,
    type: "divisionNode",
    position: { x: 0, y: 0 },
    data: {
      node: toLayoutPlaceholderNode(descriptor),
      depth: descriptor.depth,
      lod: "standard",
      onSelect: NOOP_SELECT
    }
  }));

  const flowEdges: Edge[] = descriptors
    .filter((descriptor) => Boolean(descriptor.parentId))
    .map((descriptor) => ({
      id: `edge-${descriptor.parentId}-${descriptor.id}`,
      source: descriptor.parentId as string,
      target: descriptor.id,
      type: "bezier",
      animated: false,
      focusable: false,
      selectable: false,
      reconnectable: false,
      style: {
        stroke: "#B0BAC4",
        strokeWidth: 1.5,
        pointerEvents: "none"
      },
      interactionWidth: 0
    }));

  return { nodes: flowNodes, edges: flowEdges };
}

function fallbackGraph(
  nodes: Array<Node<DivisionViewportNodeData>>,
  edges: Edge[]
): { nodes: Array<Node<DivisionViewportNodeData>>; edges: Edge[]; canvasHeight: number } {
  let y = 0;
  const laidOut = nodes.map((node) => {
    const dims = nodeDimensionsByDepth(node.data.depth);
    const positioned = {
      ...node,
      position: { x: 0, y },
      sourcePosition: Position.Bottom,
      targetPosition: Position.Top
    };
    y += dims.height + 64;
    return positioned;
  });

  return {
    nodes: laidOut,
    edges,
    canvasHeight: Math.max(600, y + 96)
  };
}

function ZoomControlPanel() {
  const { zoomIn, zoomOut } = useReactFlow();

  const buttonStyle: CSSProperties = {
    width: 32,
    height: 32,
    border: "none",
    background: "transparent",
    color: "var(--color-text-secondary)",
    fontSize: "var(--text-md)",
    lineHeight: 1,
    cursor: "pointer",
    display: "inline-flex",
    alignItems: "center",
    justifyContent: "center"
  };

  return (
    <Panel position="bottom-right" style={{ margin: 0, right: 12, bottom: 12 }}>
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          overflow: "hidden",
          border: "1px solid var(--color-border)",
          borderRadius: "var(--radius-card)",
          background: "var(--color-surface)",
          boxShadow: "var(--shadow-sm)"
        }}
      >
        <button
          aria-label="Увеличить"
          onClick={() => void zoomIn({ duration: 120 })}
          style={buttonStyle}
          type="button"
        >
          +
        </button>
        <button
          aria-label="Уменьшить"
          onClick={() => void zoomOut({ duration: 120 })}
          style={{ ...buttonStyle, borderTop: "1px solid var(--color-border)" }}
          type="button"
        >
          −
        </button>
      </div>
    </Panel>
  );
}

type DivisionTreeViewportProps = {
  nodes: DivisionViewportNode[];
  onSelect: (id: string) => void;
  debugZoom?: number;
  panelWidth?: number;
  isRefreshing?: boolean;
  fitResetKey?: number;
};

type ViewportGraph = {
  nodes: Array<Node<DivisionViewportNodeData>>;
  edges: Edge[];
  canvasHeight: number;
};

type InitialTreeFitProps = {
  nodeCount: number;
  panelWidth: number;
  isRefreshing: boolean;
  fitResetKey: number;
  triggerKey: string;
};

function InitialTreeFit({ nodeCount, panelWidth, isRefreshing, fitResetKey, triggerKey }: InitialTreeFitProps) {
  const hasFitOnLoad = useRef(false);
  const { fitView, getViewport, setViewport } = useReactFlow();

  useEffect(() => {
    hasFitOnLoad.current = false;
  }, [fitResetKey]);

  useEffect(() => {
    if (isRefreshing || nodeCount === 0 || hasFitOnLoad.current) {
      return;
    }

    hasFitOnLoad.current = true;

    let cancelled = false;
    let raf = 0;

    raf = requestAnimationFrame(() => {
      void (async () => {
        await fitView({ padding: 0.08, duration: 0 });
        if (cancelled) {
          return;
        }
        const viewport = getViewport();
        await setViewport({ ...viewport, x: viewport.x - panelWidth / 2 }, { duration: 0 });
      })();
    });

    return () => {
      cancelled = true;
      if (raf) {
        cancelAnimationFrame(raf);
      }
    };
  }, [fitView, fitResetKey, getViewport, isRefreshing, nodeCount, panelWidth, setViewport, triggerKey]);

  return null;
}

export function DivisionTreeViewport({
  nodes,
  onSelect,
  debugZoom,
  panelWidth = 380,
  isRefreshing = false,
  fitResetKey = 0
}: DivisionTreeViewportProps) {
  const [lod, setLod] = useState<DivisionZoomLodTier>(() => zoomLod(debugZoom ?? 1));

  const layoutSignature = useMemo(
    () => JSON.stringify(nodes.map((node) => ({ id: node.id, parentId: node.parentId, depth: node.depth }))),
    [nodes]
  );

  const layoutDescriptors = useMemo(
    () => JSON.parse(layoutSignature) as DivisionLayoutDescriptor[],
    [layoutSignature]
  );
  const layoutFlow = useMemo(() => buildLayoutFlow(layoutDescriptors), [layoutDescriptors]);

  const [graph, setGraph] = useState<ViewportGraph | null>(null);

  useEffect(() => {
    let cancelled = false;
    void divisionTreeLayout(layoutFlow.nodes, layoutFlow.edges)
      .then((next) => {
        if (!cancelled) {
          setGraph(next);
        }
      })
      .catch(() => {
        if (!cancelled) {
          setGraph(fallbackGraph(layoutFlow.nodes, layoutFlow.edges));
        }
      });
    return () => {
      cancelled = true;
    };
  }, [layoutFlow]);

  const nodeById = useMemo(() => new Map(nodes.map((node) => [node.id, node])), [nodes]);
  const renderedNodes = useMemo(
    () => (graph?.nodes ?? []).map((graphNode) => {
      const latestNode = nodeById.get(graphNode.id) ?? graphNode.data.node;
      return {
        ...graphNode,
        data: {
          ...graphNode.data,
          node: latestNode,
          depth: latestNode.depth,
          lod,
          onSelect
        }
      };
    }),
    [graph?.nodes, lod, nodeById, onSelect]
  );

  useEffect(() => {
    if (debugZoom !== undefined) {
      setLod(zoomLod(debugZoom));
    }
  }, [debugZoom]);

  function handleMove(_: MouseEvent | TouchEvent | null, viewport: Viewport) {
    if (debugZoom !== undefined) {
      return;
    }
    const nextLod = zoomLod(viewport.zoom);
    setLod((current) => (current === nextLod ? current : nextLod));
  }

  return (
    <div data-testid="division-tree-viewport" style={{ width: "100%", height: "100%" }}>
      {graph ? (
        <div style={{ width: "100%", height: `${graph.canvasHeight}px`, minHeight: "100%" }}>
          <ReactFlow
            defaultEdgeOptions={defaultEdgeOptions}
            edges={graph.edges}
            edgesFocusable={false}
            elementsSelectable={false}
            nodes={renderedNodes}
            nodesConnectable={false}
            nodesDraggable={false}
            nodesFocusable={false}
            nodeTypes={nodeTypes}
            onMove={handleMove}
            onNodeClick={(_, node) => onSelect(node.id)}
            panOnDrag
            proOptions={{ hideAttribution: true }}
            zoomOnScroll
          >
            <Background color="#CBD5E1" gap={24} size={1} variant={BackgroundVariant.Dots} />
            <InitialTreeFit
              fitResetKey={fitResetKey}
              isRefreshing={isRefreshing}
              nodeCount={renderedNodes.length}
              panelWidth={panelWidth}
              triggerKey={layoutSignature}
            />
            <ZoomControlPanel />
          </ReactFlow>
        </div>
      ) : null}
    </div>
  );
}
