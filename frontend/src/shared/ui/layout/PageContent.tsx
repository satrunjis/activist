import type { CSSProperties } from "react";

export type PageContentProps = {
  children: React.ReactNode;
  maxWidth?: number | "none"; // default "none"
  centered?: boolean; // default false
};

export function PageContent({ children, maxWidth = "none", centered = false }: PageContentProps) {
  const style: CSSProperties = {
    width: "100%",
    ...(maxWidth !== "none"
      ? {
          maxWidth
        }
      : null),
    ...(centered
      ? {
          marginLeft: "auto",
          marginRight: "auto"
        }
      : null)
  };

  return <div style={style}>{children}</div>;
}
