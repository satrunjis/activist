import "./ItemRow.css";

import type { CSSProperties, ReactNode } from "react";

export type ItemRowProps = {
  as?: "li" | "div";
  children: ReactNode;
  onClick?: () => void;
  selected?: boolean;
  disabled?: boolean;
  style?: CSSProperties;
  "data-testid"?: string;
};

export function ItemRow({
  as: Tag = "div",
  children,
  onClick,
  selected,
  disabled,
  style,
  "data-testid": testId,
}: ItemRowProps) {
  return (
    <Tag
      className="item-row"
      data-clickable={onClick ? "true" : undefined}
      data-disabled={disabled ? "true" : undefined}
      data-selected={selected ? "true" : undefined}
      data-testid={testId}
      onClick={onClick}
      style={style}
    >
      {children}
    </Tag>
  );
}
