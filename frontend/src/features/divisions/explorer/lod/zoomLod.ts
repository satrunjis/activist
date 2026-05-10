import type { DivisionZoomLodTier } from "../../tree/treeTypes";

export function zoomLod(zoom: number): DivisionZoomLodTier {
  if (zoom < 0.35) return "micro";
  if (zoom <= 0.65) return "compact";
  if (zoom <= 1.10) return "standard";
  return "detail";
}
