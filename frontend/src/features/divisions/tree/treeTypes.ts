export type DivisionViewportNode = {
  id: string;
  parentId?: string;
  depth: number;
  shortName: string;
  fullName?: string;
  description?: string;
  isArchived: boolean;
  hasChildren: boolean;
  childrenCount: number;
  positionsCount?: number;
  membersCount?: number;
  selected: boolean;
};

/** Four zoom-based detail tiers for semantic zoom */
export type DivisionZoomLodTier = "micro" | "compact" | "standard" | "detail";

export const DIVISION_TREE_MIN_CANVAS_HEIGHT_CSS = "100%";
export const DIVISION_TREE_DESKTOP_MIN_CANVAS_WIDTH_RATIO = 1;

export type DivisionViewportNodeData = {
  node: DivisionViewportNode;
  depth: number;
  lod: DivisionZoomLodTier;
  onSelect: (id: string) => void;
};
