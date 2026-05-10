import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import { Button } from "../../components/ui/button";
import { request } from "../../shared/api/client";
import { adaptApiError } from "../../shared/api/errorAdapter";
import { EmptyStateCard, SectionSkeleton, SoftRefreshWrapper } from "../../shared/ui/feedback";
import type { PositionItem, RoleItem } from "../../shared/api/types";

type PositionListPanelProps = {
  divisionId: string;
  selectedPositionId?: string | null;
  onStartCreatePosition: (divisionId: string) => void;
  onStartMembershipCreate: (divisionId: string, positionId: string) => void;
  onSelectPosition?: (position: PositionItem) => void;
};

export function PositionListPanel({
  divisionId,
  selectedPositionId,
  onStartCreatePosition,
  onStartMembershipCreate,
  onSelectPosition
}: PositionListPanelProps) {
  const [positions, setPositions] = useState<PositionItem[]>([]);
  const [roles, setRoles]         = useState<RoleItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const showInitialLoading = isLoading && positions.length === 0 && !errorMessage;
  const showSoftRefreshing = isLoading && positions.length > 0;
  const abortControllerRef = useRef<AbortController | null>(null);

  const roleNameById = useMemo(() => {
    const map = new Map<string, string>();
    for (const role of roles) map.set(role.id, role.name);
    return map;
  }, [roles]);

  const loadData = useCallback(async () => {
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
    const controller = new AbortController();
    abortControllerRef.current = controller;
    const { signal } = controller;

    setIsLoading(true);
    setErrorMessage(null);
    try {
      const [posRes, rolesRes] = await Promise.all([
        request<{ items: PositionItem[] }>(`/api/v1/divisions/${encodeURIComponent(divisionId)}/positions`, { signal }),
        request<{ items: RoleItem[] }>("/api/v1/roles", { signal })
      ]);
      setPositions(posRes.items);
      setRoles(rolesRes.items);
      abortControllerRef.current = null;
    } catch (error) {
      if (error instanceof DOMException && error.name === "AbortError") return;
      setErrorMessage(adaptApiError(error, "Не удалось загрузить должности.").message);
    } finally {
      setIsLoading(false);
    }
  }, [divisionId]);

  useEffect(() => {
    void loadData();
    return () => { abortControllerRef.current?.abort(); };
  }, [loadData]);

  return (
    <div data-testid="position-list-panel">
      {showInitialLoading && <SectionSkeleton rows={3} />}

      {!showInitialLoading && errorMessage && (
        <EmptyStateCard
          body="Не удалось загрузить должности. Повторите попытку."
          cta={{ label: "Повторить", onClick: () => void loadData() }}
          heading="Ошибка загрузки"
        />
      )}

      {!showInitialLoading && !errorMessage && positions.length === 0 && (
        <EmptyStateCard
          body="Создайте первую должность для этого подразделения."
          cta={{ label: "+ Добавить должность", onClick: () => onStartCreatePosition(divisionId) }}
          heading="Должностей пока нет"
        />
      )}

      {!showInitialLoading && !errorMessage && positions.length > 0 && (
        <SoftRefreshWrapper isRefreshing={showSoftRefreshing}>
          <div>
          {positions.map((position) => {
            const isSelected = selectedPositionId === position.id;
            return (
              <div
                className={`position-row${isSelected ? " position-row--selected" : ""}`}
                data-testid={`position-row-${position.id}`}
                key={position.id}
                onClick={() => onSelectPosition?.(position)}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "var(--space-2)",
                  marginBottom: "var(--space-2)",
                  border: "1px solid var(--color-border)",
                  borderLeft: isSelected
                    ? "3px solid var(--color-brand-primary)"
                    : "3px solid transparent",
                  borderRadius: "var(--radius-card)",
                  padding: "var(--space-2)",
                  background: isSelected ? "var(--color-success-subtle)" : "var(--color-surface)",
                  boxShadow: isSelected ? "var(--shadow-sm)" : "none",
                  cursor: "pointer",
                  transition: "background var(--transition-fast), border-color var(--transition-fast), box-shadow var(--transition-fast)"
                }}
              >
                <button
                  onClick={(event) => {
                    event.stopPropagation();
                    onSelectPosition?.(position);
                  }}
                  style={{
                    flex: 1,
                    minWidth: 0,
                    display: "flex",
                    alignItems: "center",
                    gap: "var(--space-2)",
                    padding: "0 var(--space-1)",
                    textAlign: "left",
                    border: 0,
                    background: "transparent",
                    color: "inherit",
                    cursor: "pointer"
                  }}
                  type="button"
                >
                  <div className="position-row__content" style={{ flex: 1, minWidth: 0 }}>
                    <div className="position-row__title" style={{ fontSize: "var(--text-sm)", fontWeight: "var(--weight-semibold)" }}>
                      {position.title}
                    </div>
                    <div className="position-row__meta" style={{ fontSize: "var(--text-xs)", color: "var(--color-text-muted)" }}>
                      {roleNameById.get(position.role_id) ?? position.role_id}
                    </div>
                  </div>
                </button>

                <Button
                  aria-label={`Добавить участника на должность ${position.title}`}
                  data-testid={`assign-member-${position.id}`}
                  onClick={(event) => {
                    event.stopPropagation();
                    onStartMembershipCreate(divisionId, position.id);
                  }}
                  size="sm"
                  style={{ flexShrink: 0 }}
                  title="Добавить участника"
                  type="button"
                  variant="primary"
                >
                  +
                </Button>
              </div>
            );
          })}
          </div>
        </SoftRefreshWrapper>
      )}

      {positions.length > 0 && (
        <div style={{ marginTop: "4px" }}>
          <Button
            data-testid="create-position"
            onClick={() => onStartCreatePosition(divisionId)}
            size="sm"
            type="button"
            variant="ghost"
          >
            + Добавить должность
          </Button>
        </div>
      )}
    </div>
  );
}
