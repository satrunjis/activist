import { AlertCircle, X } from "lucide-react";
type GlobalApiBannerSeverity = "error" | "warning" | "info";

export type GlobalApiBannerProps = {
  message: string;
  severity?: GlobalApiBannerSeverity;
  retryLabel?: string;
  onRetry?: () => void;
  onDismiss?: () => void;
};

type BannerSeverityStyles = {
  background: string;
  borderColor: string;
  color: string;
  iconColor: string;
};

const severityStyles: Record<GlobalApiBannerSeverity, BannerSeverityStyles> = {
  error: {
    background: "var(--color-error-subtle)",
    borderColor: "var(--color-border-error)",
    color: "var(--color-error-text)",
    iconColor: "var(--color-error)"
  },
  warning: {
    background: "var(--color-warning-subtle)",
    borderColor: "var(--color-warning)",
    color: "var(--color-warning-text)",
    iconColor: "var(--color-warning)"
  },
  info: {
    background: "var(--color-info-subtle)",
    borderColor: "var(--color-info)",
    color: "var(--color-info-text)",
    iconColor: "var(--color-info)"
  }
};

export function GlobalApiBanner({
  message,
  severity = "error",
  retryLabel,
  onRetry,
  onDismiss
}: GlobalApiBannerProps) {
  const styles = severityStyles[severity];
  const showRetry = Boolean(retryLabel && onRetry);

  return (
    <div
      data-testid="global-api-banner"
      role="alert"
      aria-live="polite"
      style={{
        position: "fixed",
        top: "var(--topbar-height)",
        left: 0,
        right: 0,
        zIndex: 90,
        display: "flex",
        alignItems: "center",
        gap: "var(--space-3)",
        padding: "var(--space-3) var(--space-6)",
        borderBottom: `1px solid ${styles.borderColor}`,
        background: styles.background,
        color: styles.color,
        fontSize: "var(--text-sm)",
        lineHeight: "var(--leading-normal)"
      }}
    >
      <AlertCircle aria-hidden size={16} style={{ color: styles.iconColor, flexShrink: 0 }} />
      <span style={{ flex: 1 }}>{message}</span>
      {showRetry ? (
        <button
          type="button"
          onClick={onRetry}
          style={{
            height: 28,
            padding: "0 var(--space-3)",
            borderRadius: "var(--radius-input)",
            border: `1px solid ${styles.borderColor}`,
            background: "transparent",
            color: styles.color,
            fontSize: "var(--text-xs)",
            fontWeight: "var(--weight-semibold)",
            cursor: "pointer",
            flexShrink: 0
          }}
        >
          {retryLabel}
        </button>
      ) : null}
      {onDismiss ? (
        <button
          type="button"
          onClick={onDismiss}
          aria-label="Закрыть"
          style={{
            width: 24,
            height: 24,
            display: "inline-flex",
            alignItems: "center",
            justifyContent: "center",
            borderRadius: "var(--radius-input)",
            border: "none",
            background: "transparent",
            color: styles.color,
            cursor: "pointer",
            flexShrink: 0,
            padding: 0
          }}
        >
          <X aria-hidden size={14} />
        </button>
      ) : null}
    </div>
  );
}
