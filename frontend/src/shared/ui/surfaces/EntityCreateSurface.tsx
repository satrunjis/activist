import { X } from "lucide-react";

import { Button } from "../../../components/ui/button";

export type EntityCreateSurfaceProps = {
  title?: string;
  description?: string;
  children: React.ReactNode; // usually EntityFormFrame
  onClose: () => void;
  showTitle?: boolean;
};

export function EntityCreateSurface({
  title,
  description,
  children,
  onClose,
  showTitle = true
}: EntityCreateSurfaceProps) {
  const showHeaderText = showTitle && Boolean(title || description);

  return (
    <section
      style={{
        border: "1px solid var(--color-border)",
        borderRadius: "var(--radius-card)",
        padding: "var(--space-3)",
        background: "var(--color-surface-subtle)"
      }}
    >
      <header
        style={{
          display: "flex",
          alignItems: "flex-start",
          justifyContent: showHeaderText ? "space-between" : "flex-end",
          gap: "var(--space-2)"
        }}
      >
        {showHeaderText ? (
          <div style={{ minWidth: 0 }}>
            {title ? <h3 style={{ margin: 0, fontSize: "var(--text-sm)", fontWeight: "var(--weight-bold)" }}>{title}</h3> : null}
            {description ? (
              <p style={{ margin: "var(--space-1) 0 0", fontSize: "var(--text-xs)", color: "var(--color-text-muted)" }}>
                {description}
              </p>
            ) : null}
          </div>
        ) : null}
        <Button
          aria-label="Закрыть форму создания"
          onClick={onClose}
          size="sm"
          style={{ width: 28, height: 28, padding: 0 }}
          type="button"
          variant="ghost"
        >
          <X size={14} />
        </Button>
      </header>
      <div style={{ marginTop: "var(--space-3)" }}>{children}</div>
    </section>
  );
}
