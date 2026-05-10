/**
 * SectionSkeleton — skeleton composition for list and tree sections.
 *
 * Design contract: D-04, D-06 from 09-CONTEXT.md
 * - Renders N skeleton rows to represent a loading list or tree panel.
 * - Each row can include an optional indent level for tree-like surfaces.
 *
 * Usage:
 *   <SectionSkeleton rows={5} />
 *   <SectionSkeleton rows={3} withHeader />
 */

import { SkeletonBlock } from "./SkeletonBlock";

export interface SectionSkeletonProps {
  /** Number of skeleton rows to render. Defaults to 5. */
  rows?: number;
  /** Whether to show a larger header skeleton above the rows. */
  withHeader?: boolean;
  /** Additional CSS class for the wrapper. */
  className?: string;
}

const ROW_WIDTH_CYCLE = ["100%", "80%", "90%", "65%", "85%"] as const;

export function SectionSkeleton({ rows = 5, withHeader = false, className }: SectionSkeletonProps) {
  return (
    <div
      aria-busy="true"
      aria-label="Загрузка..."
      style={{
        display: "flex",
        flexDirection: "column",
        gap: "var(--space-2)",
      }}
      className={className}
    >
      {withHeader ? <SkeletonBlock height={16} width="45%" /> : null}
      {Array.from({ length: rows }).map((_, i) => (
        <div
          key={i}
          style={{
            display: "flex",
            alignItems: "center",
            gap: "var(--space-2)",
          }}
        >
          <SkeletonBlock height={16} width={ROW_WIDTH_CYCLE[i % ROW_WIDTH_CYCLE.length]} />
        </div>
      ))}
    </div>
  );
}
