import type { CSSProperties } from "react";

export type PageRootProps = {
  children: React.ReactNode;
  "data-testid"?: string;
  fullHeight?: boolean; // false by default; true for viewport-bound pages
};

export function PageRoot({ children, "data-testid": dataTestId, fullHeight = false }: PageRootProps) {
  const style: CSSProperties = {
    display: "flex",
    flexDirection: "column",
    gap: "var(--space-6)",
    padding: "var(--space-6)",
    ...(fullHeight
      ? {
          minHeight: "calc(100vh - var(--topbar-height))"
        }
      : null)
  };

  return (
    <section data-testid={dataTestId} style={style}>
      {children}
    </section>
  );
}
