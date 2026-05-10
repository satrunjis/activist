type FormRootErrorProps = {
  message?: string | null;
};

export function FormRootError({ message }: FormRootErrorProps) {
  const normalized = message?.trim();
  if (!normalized) {
    return null;
  }

  return (
    <div
      className="[background:var(--color-error-subtle)] [border:1px_solid_var(--color-border-error)] [border-radius:var(--radius-input)] px-[var(--space-3)] py-[var(--space-2)] [font-size:var(--text-sm)] [color:var(--color-error-text)]"
      role="alert"
    >
      {normalized}
    </div>
  );
}

export type { FormRootErrorProps };
