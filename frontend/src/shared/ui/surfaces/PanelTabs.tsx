import type { CSSProperties } from "react";

export type PanelTab = {
  id: string;
  label: string;
};

export type PanelTabsProps = {
  tabs: PanelTab[];
  activeTab: string;
  onTabChange: (id: string) => void;
};

export function PanelTabs({ tabs, activeTab, onTabChange }: PanelTabsProps) {
  const containerStyle: CSSProperties = {
    display: "flex",
    width: "100%",
    height: "36px",
    borderBottom: "1px solid var(--color-border)",
    background: "var(--color-surface)",
  };

  return (
    <div style={containerStyle}>
      {tabs.map((tab) => {
        const isActive = tab.id === activeTab;
        const tabStyle: CSSProperties = {
          flex: 1,
          height: "100%",
          border: "none",
          borderBottom: isActive
            ? "2px solid var(--color-brand-primary)"
            : "2px solid transparent",
          background: "transparent",
          cursor: "pointer",
          fontSize: "var(--text-sm)",
          fontWeight: isActive ? "var(--weight-semibold)" : "normal",
          color: isActive ? "var(--color-text-primary)" : "var(--color-text-muted)",
          padding: 0,
          borderRadius: 0,
          textAlign: "center",
        };

        return (
          <button
            data-testid={`panel-tab-${tab.id}`}
            key={tab.id}
            onClick={() => onTabChange(tab.id)}
            style={tabStyle}
            type="button"
          >
            {tab.label}
          </button>
        );
      })}
    </div>
  );
}
