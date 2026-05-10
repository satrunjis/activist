/**
 * EmptyStateCard — consistent empty-state container with explanatory copy and optional CTA.
 *
 * Design contract: D-11, D-12, D-13 from 09-CONTEXT.md
 * - Every empty state MUST include explanatory text (heading + body).
 * - CTA slot is rendered ONLY when caller passes CTA props (i.e., actor has creation permission).
 * - Targets: divisions tree, positions, memberships, roles, search results.
 *
 * Copywriting contract (09-UI-SPEC.md):
 *   heading default: "Пока ничего нет"
 *   body default:    "В этом разделе еще нет данных. Добавьте первую запись, если у вас есть права, или измените фильтры."
 *
 * Usage (no CTA — viewer has no create permission):
 *   <EmptyStateCard />
 *
 * Usage (with CTA — actor has create permission):
 *   <EmptyStateCard
 *     ctaLabel="Создать подразделение"
 *     onCta={() => openCreateDialog()}
 *   />
 *
 * Usage (custom copy):
 *   <EmptyStateCard
 *     heading="Нет назначений"
 *     body="У этого пользователя пока нет активных назначений."
 *   />
 */
import { CircleAlert } from "lucide-react";
import { cloneElement, isValidElement, type ReactNode } from "react";

type EmptyStateIconProps = {
  size?: number;
  color?: string;
  "aria-hidden"?: boolean;
};

export interface EmptyStateCardProps {
  /** Icon shown above heading. */
  icon?: ReactNode;
  /**
   * Primary heading for the empty state.
   * Default: "Пока ничего нет"
   */
  heading?: string;
  /**
   * Explanatory body text shown below the heading.
   * Default: UI-SPEC empty state body copy.
   */
  body?: string;
  /**
   * Optional call-to-action button.
   */
  cta?: {
    label: string;
    onClick: () => void;
  };
  /** @deprecated Use cta={{ label, onClick }}. */
  ctaLabel?: string;
  /**
   * @deprecated Use cta={{ label, onClick }}.
   */
  onCta?: () => void;
  /** Additional CSS class for the card wrapper. */
  className?: string;
}

const DEFAULT_HEADING = "Пока ничего нет";
const DEFAULT_BODY =
  "В этом разделе еще нет данных. Добавьте первую запись, если у вас есть права, или измените фильтры.";

export function EmptyStateCard({
  icon,
  heading = DEFAULT_HEADING,
  body = DEFAULT_BODY,
  cta,
  ctaLabel,
  onCta,
  className,
}: EmptyStateCardProps) {
  const normalizedCta = cta ?? (ctaLabel && onCta ? { label: ctaLabel, onClick: onCta } : undefined);
  const iconNode = icon ?? <CircleAlert />;

  return (
    <div
      role="status"
      aria-label={heading}
      className={className}
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        padding: "var(--space-12) var(--space-6)",
        textAlign: "center",
      }}
    >
      <div
        aria-hidden
        style={{
          color: "var(--color-text-muted)",
          width: 36,
          height: 36,
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
        }}
      >
        {isValidElement<EmptyStateIconProps>(iconNode)
          ? cloneElement<EmptyStateIconProps>(iconNode, { size: 36, color: "var(--color-text-muted)", "aria-hidden": true })
          : iconNode}
      </div>

      <p
        style={{
          margin: 0,
          marginTop: "var(--space-3)",
          fontSize: "var(--text-base)",
          fontWeight: "var(--weight-semibold)",
          lineHeight: "var(--leading-normal)",
          color: "var(--color-text-primary)",
        }}
      >
        {heading}
      </p>

      <p
        style={{
          margin: 0,
          marginTop: "var(--space-1)",
          fontSize: "var(--text-sm)",
          lineHeight: "var(--leading-normal)",
          color: "var(--color-text-muted)",
          maxWidth: "40ch",
        }}
      >
        {body}
      </p>

      {normalizedCta ? (
        <button
          type="button"
          onClick={normalizedCta.onClick}
          style={{
            marginTop: "var(--space-4)",
            padding: "0 var(--space-4)",
            height: 36,
            backgroundColor: "var(--color-brand-primary)",
            color: "var(--color-text-inverse)",
            border: "none",
            borderRadius: "var(--radius-pill)",
            fontSize: "var(--text-sm)",
            fontWeight: "var(--weight-semibold)",
            cursor: "pointer",
          }}
        >
          {normalizedCta.label}
        </button>
      ) : null}
    </div>
  );
}
