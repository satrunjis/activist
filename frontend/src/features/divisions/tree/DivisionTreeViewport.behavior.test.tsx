import { render, screen } from "@testing-library/react";
import { beforeAll, describe, expect, it, vi } from "vitest";

import { DivisionTreeViewport } from "./DivisionTreeViewport";
import type { DivisionViewportNode } from "./treeTypes";

beforeAll(() => {
  if (!window.ResizeObserver) {
    class ResizeObserverMock {
      observe() {}
      unobserve() {}
      disconnect() {}
    }
    window.ResizeObserver = ResizeObserverMock as unknown as typeof ResizeObserver;
  }
});

function buildNodes(): DivisionViewportNode[] {
  return [
    {
      id: "root",
      depth: 0,
      shortName: "Central Committee",
      fullName: "Central Committee",
      hasChildren: true,
      childrenCount: 1,
      isArchived: false,
      selected: true
    },
    {
      id: "ops",
      parentId: "root",
      depth: 1,
      shortName: "Operations",
      fullName: "Operations Department",
      hasChildren: false,
      childrenCount: 0,
      isArchived: false,
      selected: false
    }
  ];
}

describe("DivisionTreeViewport behavior", () => {
  it("renders hierarchy nodes and pan/zoom controls", async () => {
    const onSelect = vi.fn();

    render(<DivisionTreeViewport debugZoom={1} nodes={buildNodes()} onSelect={onSelect} />);

    expect(await screen.findByTestId("division-tree-viewport")).toBeInTheDocument();
    expect(await screen.findByTestId("division-node-root")).toBeInTheDocument();
    expect(await screen.findByTestId("division-node-ops")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Увеличить" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Уменьшить" })).toBeInTheDocument();
  });

  it("calls onSelect when a node is clicked", async () => {
    const onSelect = vi.fn();

    render(<DivisionTreeViewport debugZoom={1} nodes={buildNodes()} onSelect={onSelect} />);

    const rootNode = await screen.findByTestId("division-node-root");
    rootNode.click();
    expect(onSelect).toHaveBeenCalledWith("root");
  });

  it("applies accent border styling to selected node", async () => {
    render(<DivisionTreeViewport debugZoom={1} nodes={buildNodes()} onSelect={vi.fn()} />);

    const rootNodeButton = await screen.findByTestId("division-node-root");
    expect(rootNodeButton).toBeInTheDocument();
    expect(rootNodeButton.getAttribute("style")).toContain("border-color: var(--color-node-border-selected)");
  });

  it("suppresses browser default focus outline on nodes", async () => {
    render(<DivisionTreeViewport debugZoom={1} nodes={buildNodes()} onSelect={vi.fn()} />);

    const rootNodeButton = await screen.findByTestId("division-node-root");
    expect(rootNodeButton.getAttribute("style")).toContain("outline: none");
  });

  it("renders micro LOD at very low zoom", async () => {
    render(<DivisionTreeViewport debugZoom={0.2} nodes={buildNodes()} onSelect={vi.fn()} />);

    expect(await screen.findByTestId("division-node-root")).toBeInTheDocument();
  });

  it("shows children count at standard LOD", async () => {
    render(<DivisionTreeViewport debugZoom={0.8} nodes={buildNodes()} onSelect={vi.fn()} />);

    const matches = await screen.findAllByText(/должн\./i);
    expect(matches.length).toBeGreaterThan(0);
  });

  it("shows hint text at detail LOD", async () => {
    render(<DivisionTreeViewport debugZoom={1.3} nodes={buildNodes()} onSelect={vi.fn()} />);

    const labels = await screen.findAllByText(/Central Committee/i);
    expect(labels.length).toBeGreaterThan(0);
  });
});
