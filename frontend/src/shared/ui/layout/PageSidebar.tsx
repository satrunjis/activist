import type { CSSProperties } from "react";

export type PageSidebarProps = {
  children: React.ReactNode;
  sticky?: boolean; // default false
};

export function PageSidebar({ children, sticky = false }: PageSidebarProps) {
  const style: CSSProperties = {
    display: "flex",
    flexDirection: "column",
    gap: "var(--space-4)",
    minWidth: 0,
    alignSelf: "start",
    ...(sticky
      ? {
          position: "sticky",
          top: "calc(var(--topbar-height) + var(--space-6))"
        }
      : null)
  };

  return <aside style={style}>{children}</aside>;
}
