export type PageHeaderProps = {
  title: string;
  description?: string;
  actions?: React.ReactNode;
  "data-testid"?: string;
};

export function PageHeader({ title, description, actions, "data-testid": dataTestId }: PageHeaderProps) {
  return (
    <header
      data-testid={dataTestId}
      style={{
        display: "flex",
        alignItems: "flex-start",
        justifyContent: "space-between",
        gap: "var(--space-4)"
      }}
    >
      <div style={{ minWidth: 0, flex: 1 }}>
        <h1
          style={{
            margin: 0,
            fontSize: "var(--text-xl)",
            fontWeight: "var(--weight-bold)",
            color: "var(--color-text-primary)"
          }}
        >
          {title}
        </h1>
        {description ? (
          <p
            style={{
              margin: "var(--space-1) 0 0",
              fontSize: "var(--text-sm)",
              color: "var(--color-text-muted)"
            }}
          >
            {description}
          </p>
        ) : null}
      </div>
      {actions ? (
        <div style={{ display: "flex", alignItems: "center", gap: "var(--space-2)", flexShrink: 0 }}>
          {actions}
        </div>
      ) : null}
    </header>
  );
}
