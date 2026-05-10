import type { CSSProperties } from "react";

export type ActionPanelProps = {
  title: React.ReactNode;
  subtitle?: React.ReactNode;
  headerActions?: React.ReactNode;
  children: React.ReactNode;
  width?: number; // default 360
  stickyHeader?: boolean; // default true
};

export function ActionPanel({
  title,
  subtitle,
  headerActions,
  children,
  width = 360,
  stickyHeader = true
}: ActionPanelProps) {
  const rootStyle: CSSProperties = {
    width: `${width}px`,
    maxWidth: "100%",
    background: "var(--color-surface)",
    borderRadius: "var(--radius-card)",
    boxShadow: "var(--shadow-sm)",
    border: "1px solid var(--color-border)",
    overflow: "hidden",
    display: "flex",
    flexDirection: "column",
    maxHeight: "100%",
    minHeight: 0,
    minWidth: 0
  };

  const headerStyle: CSSProperties = {
    padding: "var(--space-3) var(--space-4)",
    borderBottom: "1px solid var(--color-border)",
    background: "var(--color-surface)",
    display: "flex",
    alignItems: "flex-start",
    gap: "var(--space-3)",
    ...(stickyHeader
      ? {
          position: "sticky",
          top: 0,
          zIndex: 1
        }
      : null)
  };

  return (
    <section style={rootStyle}>
      <header style={headerStyle}>
        <div style={{ flex: 1, minWidth: 0 }}>
          <h2 style={{ margin: 0, fontSize: "var(--text-base)", fontWeight: "var(--weight-bold)" }}>{title}</h2>
          {subtitle ? (
            <p style={{ margin: "var(--space-1) 0 0", fontSize: "var(--text-sm)", color: "var(--color-text-muted)" }}>
              {subtitle}
            </p>
          ) : null}
        </div>
        {headerActions ? <div style={{ flexShrink: 0 }}>{headerActions}</div> : null}
      </header>

      <div style={{ padding: "var(--space-4)", overflowY: "auto", flex: 1, minHeight: 0 }}>{children}</div>
    </section>
  );
}
