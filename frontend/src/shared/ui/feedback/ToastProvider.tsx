import { CheckCircle2, Info, X, XCircle } from "lucide-react";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode
} from "react";

export type ToastVariant = "success" | "error" | "info";

type ToastRecord = {
  id: number;
  variant: ToastVariant;
  title: string;
  body?: string;
  exiting: boolean;
};

type ToastContextValue = {
  showToast: (variant: ToastVariant, title: string, body?: string) => void;
};

const MAX_TOASTS = 4;
const AUTO_DISMISS_MS = 4000;
const EXIT_DURATION_MS = 150;

const ToastContext = createContext<ToastContextValue | undefined>(undefined);

const variantStyles: Record<
ToastVariant,
{ accent: string; iconColor: string; Icon: typeof CheckCircle2 }
> = {
  success: {
    accent: "var(--color-success)",
    iconColor: "var(--color-success)",
    Icon: CheckCircle2
  },
  error: {
    accent: "var(--color-error)",
    iconColor: "var(--color-error)",
    Icon: XCircle
  },
  info: {
    accent: "var(--color-info)",
    iconColor: "var(--color-info)",
    Icon: Info
  }
};

type ToastProviderProps = {
  children: ReactNode;
};

type ToastItemProps = {
  toast: ToastRecord;
  onDismiss: (id: number) => void;
};

function ToastItem({ toast, onDismiss }: ToastItemProps) {
  const [entered, setEntered] = useState(false);
  const variant = variantStyles[toast.variant];

  useEffect(() => {
    const frame = window.requestAnimationFrame(() => setEntered(true));
    return () => window.cancelAnimationFrame(frame);
  }, []);

  useEffect(() => {
    if (toast.exiting) return undefined;
    const timer = window.setTimeout(() => onDismiss(toast.id), AUTO_DISMISS_MS);
    return () => window.clearTimeout(timer);
  }, [onDismiss, toast.exiting, toast.id]);

  const transform = toast.exiting
    ? "translateX(calc(100% + 16px))"
    : entered
      ? "translateX(0)"
      : "translateX(calc(100% + 16px))";

  return (
    <div
      role="status"
      style={{
        width: "var(--toast-width)",
        background: "var(--color-surface)",
        borderRadius: "var(--radius-card)",
        boxShadow: "var(--shadow-toast)",
        overflow: "hidden",
        transform,
        opacity: toast.exiting ? 0 : 1,
        transition: "transform 200ms ease, opacity 150ms ease",
        display: "flex"
      }}
    >
      <div
        aria-hidden
        style={{
          width: 4,
          background: variant.accent,
          borderTopLeftRadius: "var(--radius-card)",
          borderBottomLeftRadius: "var(--radius-card)",
          flexShrink: 0
        }}
      />
      <div
        style={{
          display: "flex",
          alignItems: "flex-start",
          gap: "var(--space-2)",
          padding: "var(--space-3) var(--space-3) var(--space-3) var(--space-4)",
          flex: 1
        }}
      >
        <variant.Icon aria-hidden size={20} style={{ color: variant.iconColor, flexShrink: 0 }} />
        <div style={{ flex: 1, minWidth: 0 }}>
          <p
            style={{
              margin: 0,
              fontSize: "var(--text-sm)",
              fontWeight: "var(--weight-semibold)",
              color: "var(--color-text-primary)"
            }}
          >
            {toast.title}
          </p>
          {toast.body ? (
            <p
              style={{
                margin: "var(--space-1) 0 0",
                fontSize: "var(--text-xs)",
                color: "var(--color-text-muted)",
                lineHeight: "var(--leading-normal)"
              }}
            >
              {toast.body}
            </p>
          ) : null}
        </div>
        <button
          aria-label="Закрыть уведомление"
          onClick={() => onDismiss(toast.id)}
          style={{
            width: 20,
            height: 20,
            display: "inline-flex",
            alignItems: "center",
            justifyContent: "center",
            border: "none",
            background: "transparent",
            color: "var(--color-text-muted)",
            cursor: "pointer",
            borderRadius: "var(--radius-input)",
            padding: 0,
            flexShrink: 0
          }}
          type="button"
        >
          <X aria-hidden size={14} />
        </button>
      </div>
    </div>
  );
}

export function ToastProvider({ children }: ToastProviderProps) {
  const [toasts, setToasts] = useState<ToastRecord[]>([]);
  const nextIdRef = useRef(0);
  const removalTimersRef = useRef(new Map<number, number>());
  const toastsRef = useRef<ToastRecord[]>([]);

  useEffect(() => {
    toastsRef.current = toasts;
  }, [toasts]);

  useEffect(() => {
    const timers = removalTimersRef.current;
    return () => {
      for (const timer of timers.values()) {
        window.clearTimeout(timer);
      }
      timers.clear();
    };
  }, []);

  const dismissToast = useCallback((id: number) => {
    const target = toastsRef.current.find((item) => item.id === id);
    if (!target || target.exiting || removalTimersRef.current.has(id)) {
      return;
    }

    setToasts((prev) =>
      prev.map((item) => (item.id === id ? { ...item, exiting: true } : item))
    );

    const timer = window.setTimeout(() => {
      setToasts((prev) => prev.filter((item) => item.id !== id));
      removalTimersRef.current.delete(id);
    }, EXIT_DURATION_MS);

    removalTimersRef.current.set(id, timer);
  }, []);

  const showToast = useCallback((variant: ToastVariant, title: string, body?: string) => {
    const id = nextIdRef.current++;

    setToasts((prev) => {
      const next = [...prev, { id, variant, title, body, exiting: false }];
      if (next.length <= MAX_TOASTS) {
        return next;
      }
      const overflowCount = next.length - MAX_TOASTS;
      return next.slice(overflowCount);
    });
  }, []);

  const contextValue = useMemo<ToastContextValue>(() => ({ showToast }), [showToast]);

  return (
    <ToastContext.Provider value={contextValue}>
      {children}
      <div
        aria-live="polite"
        aria-atomic="true"
        style={{
          position: "fixed",
          top: 16,
          right: 16,
          zIndex: 300,
          display: "flex",
          flexDirection: "column",
          gap: "var(--space-2)",
          pointerEvents: "none"
        }}
      >
        {toasts.map((toast) => (
          <div key={toast.id} style={{ pointerEvents: "auto" }}>
            <ToastItem onDismiss={dismissToast} toast={toast} />
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}

export function useToast() {
  const context = useContext(ToastContext);
  if (!context) {
    throw new Error("useToast must be used within ToastProvider.");
  }
  return context;
}
