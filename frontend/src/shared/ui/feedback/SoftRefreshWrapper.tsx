export type SoftRefreshWrapperProps = {
  isRefreshing: boolean;
  children: React.ReactNode;
  label?: string;
};

export function SoftRefreshWrapper({
  isRefreshing,
  children,
  label = "Обновление…"
}: SoftRefreshWrapperProps) {
  if (!isRefreshing) {
    return <>{children}</>;
  }

  return (
    <div style={{ position: "relative" }}>
      <div
        style={{
          fontSize: "var(--text-xs)",
          color: "var(--color-text-muted)",
          marginBottom: "var(--space-1)"
        }}
      >
        {label}
      </div>
      <div style={{ opacity: 0.5, pointerEvents: "none" }}>
        {children}
      </div>
    </div>
  );
}
