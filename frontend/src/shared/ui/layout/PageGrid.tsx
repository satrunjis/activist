import type { CSSProperties } from "react";

export type PageGridProps = {
  children: React.ReactNode;
  sidebarWidth?: number; // default 360
  gap?: number; // default var(--space-6)
};

export function PageGrid({ children, sidebarWidth = 360, gap }: PageGridProps) {
  const style: CSSProperties = {
    display: "grid",
    width: "100%",
    alignItems: "start",
    gridTemplateColumns: `minmax(0, 1fr) ${sidebarWidth}px`,
    gap: gap === undefined ? "var(--space-6)" : `${gap}px`
  };

  return <div style={style}>{children}</div>;
}
