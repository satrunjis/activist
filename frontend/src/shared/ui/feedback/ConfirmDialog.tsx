import { useEffect, useId } from "react";
import { createPortal } from "react-dom";

export type ConfirmDialogProps = {
  open: boolean;
  title: string;
  body: string;
  confirmLabel: string;
  destructive: boolean;
  onConfirm: () => void;
  onCancel: () => void;
};

export function ConfirmDialog({
  open,
  title,
  body,
  confirmLabel,
  destructive,
  onConfirm,
  onCancel
}: ConfirmDialogProps) {
  const titleId = useId();
  const bodyId = useId();

  useEffect(() => {
    if (!open) {
      return undefined;
    }

    function onKeyDown(event: KeyboardEvent) {
      if (event.key === "Escape") {
        onCancel();
      }
    }

    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [onCancel, open]);

  if (!open) {
    return null;
  }

  return createPortal(
    <div
      aria-hidden={false}
      onClick={onCancel}
      style={{
        position: "fixed",
        inset: 0,
        zIndex: 200,
        background: "var(--color-surface-overlay)",
        animation: "fadeIn 150ms ease",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        padding: "var(--space-6)"
      }}
    >
      <div
        aria-describedby={bodyId}
        aria-labelledby={titleId}
        aria-modal="true"
        onClick={(event) => event.stopPropagation()}
        role="dialog"
        style={{
          width: "min(100%, var(--modal-width))",
          background: "var(--color-surface)",
          borderRadius: "var(--radius-card)",
          boxShadow: "var(--shadow-modal)",
          padding: "var(--space-6)"
        }}
      >
        <h2
          id={titleId}
          style={{
            margin: 0,
            marginBottom: "var(--space-2)",
            fontSize: "var(--text-md)",
            fontWeight: "var(--weight-bold)",
            color: "var(--color-text-primary)",
            lineHeight: "var(--leading-tight)"
          }}
        >
          {title}
        </h2>
        <p
          id={bodyId}
          style={{
            margin: 0,
            marginBottom: "var(--space-6)",
            fontSize: "var(--text-sm)",
            color: "var(--color-text-secondary)",
            lineHeight: "var(--leading-normal)"
          }}
        >
          {body}
        </p>
        <div
          style={{
            display: "flex",
            justifyContent: "flex-end",
            gap: "var(--space-3)"
          }}
        >
          <button
            onClick={onCancel}
            style={{
              border: "1px solid var(--color-border)",
              background: "var(--color-surface)",
              color: "var(--color-text-primary)",
              borderRadius: "var(--radius-card)",
              height: 36,
              padding: "0 var(--space-4)",
              fontSize: "var(--text-sm)",
              fontWeight: "var(--weight-semibold)",
              cursor: "pointer"
            }}
            type="button"
          >
            Отмена
          </button>
          <button
            onClick={onConfirm}
            style={{
              border: "none",
              background: destructive ? "var(--color-error)" : "var(--color-brand-primary)",
              color: "var(--color-text-inverse)",
              borderRadius: "var(--radius-card)",
              height: 36,
              padding: "0 var(--space-4)",
              fontSize: "var(--text-sm)",
              fontWeight: "var(--weight-semibold)",
              cursor: "pointer"
            }}
            type="button"
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>,
    document.body
  );
}
