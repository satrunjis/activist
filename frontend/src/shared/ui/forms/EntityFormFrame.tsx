import { createContext, useContext, type FormEvent, type ReactNode } from "react";

import { FormRootError } from "./FormRootError";

type EntityFormFrameProps = {
  formId?: string;
  title?: string;
  description?: string;
  rootError?: string | null;
  actions: ReactNode;
  children: ReactNode;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
};

type EntityFormFrameContextValue = {
  formId?: string;
};

const EntityFormFrameContext = createContext<EntityFormFrameContextValue>({});

export function useEntityFormFrameContext() {
  return useContext(EntityFormFrameContext);
}

export function EntityFormFrame({
  formId,
  title,
  description,
  rootError,
  actions,
  children,
  onSubmit
}: EntityFormFrameProps) {
  const hasRootError = Boolean(rootError?.trim());

  return (
    <EntityFormFrameContext.Provider value={{ formId }}>
      <form data-testid={formId} id={formId} onSubmit={onSubmit}>
        {title || description ? (
          <header style={{ marginBottom: "var(--space-4)" }}>
            {title ? (
              <h2 style={{ margin: 0, fontSize: "var(--text-lg)", fontWeight: "var(--weight-bold)", color: "var(--color-text-primary)" }}>
                {title}
              </h2>
            ) : null}
            {description ? (
              <p style={{ margin: "var(--space-1) 0 0", fontSize: "var(--text-sm)", color: "var(--color-text-muted)" }}>
                {description}
              </p>
            ) : null}
          </header>
        ) : null}

        {children}
        {hasRootError ? (
          <div style={{ marginTop: "var(--space-3)" }}>
            <FormRootError message={rootError} />
          </div>
        ) : null}
        {actions}
      </form>
    </EntityFormFrameContext.Provider>
  );
}

export type { EntityFormFrameProps };
