import { useCallback, useEffect, useMemo, useRef, useState, type CSSProperties, type FormEvent } from "react";

import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
import { request } from "../../shared/api/client";
import { adaptApiError } from "../../shared/api/errorAdapter";
import { EmptyStateCard, InlineError, SectionSkeleton, SoftRefreshWrapper } from "../../shared/ui/feedback";
import { PageContent, PageGrid, PageHeader, PageRoot, PageSidebar } from "../../shared/ui/layout";
import { AsyncStateView, type AsyncState } from "../../shared/ui/states/AsyncStateView";
import { ActionPanel } from "../../shared/ui/surfaces";
import type { EventLogItem, EventLogListResponse } from "../../shared/api/types";

type AuditLogFilters = {
  event_type: string;
  subject_type: string;
  subject_id: string;
};

const DEFAULT_FILTERS: AuditLogFilters = {
  event_type: "",
  subject_type: "",
  subject_id: ""
};

const DEFAULT_LIMIT = 50;

function buildQuery(filters: AuditLogFilters, limit: number, offset: number): string {
  const query = new URLSearchParams();
  if (filters.event_type.trim()) {
    query.set("event_type", filters.event_type.trim());
  }
  if (filters.subject_type.trim()) {
    query.set("subject_type", filters.subject_type.trim());
  }
  if (filters.subject_id.trim()) {
    query.set("subject_id", filters.subject_id.trim());
  }
  query.set("limit", String(limit));
  query.set("offset", String(offset));
  return query.toString();
}

function hasAnyFilter(filters: AuditLogFilters): boolean {
  return Boolean(filters.event_type.trim() || filters.subject_type.trim() || filters.subject_id.trim());
}

export function AuditLogPage() {
  const hasLoadedOnceRef = useRef(false);
  const [hasLoadedOnce, setHasLoadedOnce] = useState(false);
  const [filterDraft, setFilterDraft] = useState<AuditLogFilters>(DEFAULT_FILTERS);
  const [appliedFilters, setAppliedFilters] = useState<AuditLogFilters>(DEFAULT_FILTERS);
  const [items, setItems] = useState<EventLogItem[]>([]);
  const [total, setTotal] = useState(0);
  const [limit] = useState(DEFAULT_LIMIT);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hoveredRowId, setHoveredRowId] = useState<string | null>(null);

  const query = useMemo(() => buildQuery(appliedFilters, limit, offset), [appliedFilters, limit, offset]);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await request<EventLogListResponse>(`/api/v1/eventlog?${query}`);
      setItems(Array.isArray(response.items) ? response.items : []);
      setTotal(typeof response.total === "number" ? response.total : 0);
      hasLoadedOnceRef.current = true;
      setHasLoadedOnce(true);
    } catch (nextError) {
      if (!hasLoadedOnceRef.current) {
        setItems([]);
        setTotal(0);
      }
      setError(adaptApiError(nextError, "Не удалось загрузить журнал аудита.").message);
    } finally {
      setLoading(false);
    }
  }, [query]);

  useEffect(() => {
    void load();
  }, [load]);

  function updateDraftField<K extends keyof AuditLogFilters>(key: K, value: AuditLogFilters[K]) {
    setFilterDraft((current) => ({ ...current, [key]: value }));
  }

  function submitFilters(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setOffset(0);
    setAppliedFilters(filterDraft);
  }

  function resetFilters() {
    setFilterDraft(DEFAULT_FILTERS);
    setAppliedFilters(DEFAULT_FILTERS);
    setOffset(0);
  }

  const canPrev = offset > 0;
  const canNext = offset + limit < total;
  const hasActiveDraftFilters = hasAnyFilter(filterDraft);
  const showSoftRefreshing = hasLoadedOnce && loading;
  const showInlineError = Boolean(error) && items.length > 0;
  const showFullError = Boolean(error) && items.length === 0;

  const rangeStart = total === 0 ? 0 : offset + 1;
  const rangeEnd = total === 0 ? 0 : Math.min(offset + limit, total);
  const initialAuditState: AsyncState = loading ? "loading" : error ? "error" : "success";

  function renderTableRows(loadedItems: EventLogItem[]) {
    return (
      <table
        data-testid="audit-log-table"
        style={{
          width: "100%",
          borderCollapse: "collapse"
        }}
      >
        <thead>
          <tr style={{ background: "var(--color-surface-subtle)", borderBottom: "2px solid var(--color-border)" }}>
            <th style={getHeaderCellStyle(180)}>Дата и время</th>
            <th style={getHeaderCellStyle()}>Тип события</th>
            <th style={getHeaderCellStyle(140)}>Тип объекта</th>
            <th style={getHeaderCellStyle(220)}>ID объекта</th>
            <th style={getHeaderCellStyle(180)}>Пользователь</th>
          </tr>
        </thead>
        <tbody>
          {loadedItems.map((item) => (
            <tr
              data-testid={`audit-log-row-${item.id}`}
              key={item.id}
              onMouseEnter={() => setHoveredRowId(item.id)}
              onMouseLeave={() => setHoveredRowId((current) => (current === item.id ? null : current))}
              style={{ background: hoveredRowId === item.id ? "var(--color-surface-subtle)" : "transparent" }}
            >
              <td style={tableCellStyle}>{item.timestamp}</td>
              <td style={tableCellStyle}>{item.event_type}</td>
              <td style={tableCellStyle}>{item.subject_type}</td>
              <td style={tableCellStyle}>{item.subject_id}</td>
              <td style={tableCellStyle}>{item.actor_id}</td>
            </tr>
          ))}
        </tbody>
      </table>
    );
  }

  const emptyTableView = (
    <>
      <table data-testid="audit-log-table" style={{ width: "100%", borderCollapse: "collapse" }}>
        <thead>
          <tr style={{ background: "var(--color-surface-subtle)", borderBottom: "2px solid var(--color-border)" }}>
            <th style={getHeaderCellStyle(180)}>Дата и время</th>
            <th style={getHeaderCellStyle()}>Тип события</th>
            <th style={getHeaderCellStyle(140)}>Тип объекта</th>
            <th style={getHeaderCellStyle(220)}>ID объекта</th>
            <th style={getHeaderCellStyle(180)}>Пользователь</th>
          </tr>
        </thead>
      </table>
      <EmptyStateCard heading="Нет записей аудита" />
    </>
  );

  return (
    <PageRoot data-testid="audit-log-page">
      <PageHeader title="Журнал аудита" />

      <PageContent>
        <PageGrid>
          <div>
            <section
              style={{
                background: "var(--color-surface)",
                borderRadius: "var(--radius-card)",
                boxShadow: "var(--shadow-sm)",
                padding: "var(--space-4)",
                overflowX: "auto",
                position: "relative"
              }}
            >
              {!hasLoadedOnce ? (
                <AsyncStateView
                  data={true}
                  errorMessage={error}
                  loadingView={<SectionSkeleton rows={5} />}
                  onRetry={() => void load()}
                  state={initialAuditState}
                >
                  {() => null}
                </AsyncStateView>
              ) : (
                <>
                  <InlineError message={showInlineError ? error : null} />

                  {showFullError ? (
                    <EmptyStateCard
                      body={error ?? "Не удалось загрузить журнал аудита."}
                      cta={{ label: "Повторить", onClick: () => void load() }}
                      heading="Ошибка загрузки"
                    />
                  ) : (
                    <SoftRefreshWrapper isRefreshing={showSoftRefreshing}>
                      {items.length === 0 ? emptyTableView : renderTableRows(items)}
                    </SoftRefreshWrapper>
                  )}
                </>
              )}
            </section>

            <div style={{ marginTop: "var(--space-3)", display: "flex", alignItems: "center", gap: "var(--space-4)" }}>
              <Button
                data-testid="audit-pagination-prev"
                disabled={!canPrev}
                onClick={() => setOffset((current) => Math.max(0, current - limit))}
                type="button"
                variant="outline"
              >
                ← Назад
              </Button>
              <span
                data-testid="audit-pagination-state"
                style={{ flex: 1, textAlign: "center", fontSize: "var(--text-sm)", color: "var(--color-text-muted)" }}
              >
                Записи {rangeStart}–{rangeEnd} из {total}
              </span>
              <Button
                data-testid="audit-pagination-next"
                disabled={!canNext}
                onClick={() => setOffset((current) => current + limit)}
                type="button"
                variant="outline"
              >
                Вперёд →
              </Button>
            </div>
          </div>

          <PageSidebar>
            <ActionPanel title="Фильтры">
              <form data-testid="audit-log-filters" onSubmit={submitFilters}>
                <div style={{ display: "grid", gap: "var(--space-3)" }}>
                  <div>
                    <Label htmlFor="audit-event-type">Тип события</Label>
                    <Input
                      data-testid="audit-filter-event_type"
                      id="audit-event-type"
                      onChange={(event) => updateDraftField("event_type", event.target.value)}
                      value={filterDraft.event_type}
                    />
                  </div>

                  <div>
                    <Label htmlFor="audit-subject-type">Тип объекта</Label>
                    <Input
                      data-testid="audit-filter-subject_type"
                      id="audit-subject-type"
                      onChange={(event) => updateDraftField("subject_type", event.target.value)}
                      value={filterDraft.subject_type}
                    />
                  </div>

                  <div>
                    <Label htmlFor="audit-subject-id">ID объекта</Label>
                    <Input
                      data-testid="audit-filter-subject_id"
                      id="audit-subject-id"
                      onChange={(event) => updateDraftField("subject_id", event.target.value)}
                      value={filterDraft.subject_id}
                    />
                  </div>
                </div>

                <Button
                  data-testid="audit-filter-submit"
                  size="default"
                  style={{ marginTop: "var(--space-3)" }}
                  type="submit"
                  variant="primary"
                >
                  Применить
                </Button>

                {hasActiveDraftFilters ? (
                  <Button
                    onClick={resetFilters}
                    style={{ width: "100%", marginTop: "var(--space-2)" }}
                    type="button"
                    variant="ghost"
                  >
                    Сбросить
                  </Button>
                ) : null}
              </form>
            </ActionPanel>
          </PageSidebar>
        </PageGrid>
      </PageContent>
    </PageRoot>
  );
}

function getHeaderCellStyle(width?: number): CSSProperties {
  return {
    width,
    textAlign: "left",
    padding: "var(--space-2) var(--space-3)",
    fontSize: "var(--text-xs)",
    fontWeight: "var(--weight-semibold)",
    color: "var(--color-text-muted)",
    textTransform: "uppercase",
    letterSpacing: "0.06em"
  };
}

const tableCellStyle: CSSProperties = {
  padding: "var(--space-2) var(--space-3)",
  fontSize: "var(--text-sm)",
  color: "var(--color-text-primary)",
  borderBottom: "1px solid var(--color-border)"
};
