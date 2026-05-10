import { useRef, useState, type FormEvent } from "react";

import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import { Label } from "../../components/ui/label";
import { request } from "../../shared/api/client";
import { adaptApiError } from "../../shared/api/errorAdapter";
import { SoftRefreshWrapper, type GlobalApiBannerProps } from "../../shared/ui/feedback";
import { PageContent, PageGrid, PageRoot, PageSidebar } from "../../shared/ui/layout";
import { AsyncStateView, type AsyncState } from "../../shared/ui/states/AsyncStateView";
import type { SearchUserItem, UserSearchRequest, UserSearchResponse } from "../../shared/api/types";
import { SearchResults } from "./SearchResults";
import { SearchUserDetailsPanel } from "./SearchUserDetailsPanel";

type SearchFormState = {
  first_name: string;
  last_name: string;
  middle_name: string;
  login: string;
  group_number: string;
  institute: string;
  about: string;
  position_title: string;
  role_name: string;
};

type SearchPageProps = {
  onGlobalError?: (next: Omit<GlobalApiBannerProps, "onDismiss"> | null) => void;
};

const DEFAULT_FORM: SearchFormState = {
  first_name: "",
  last_name: "",
  middle_name: "",
  login: "",
  group_number: "",
  institute: "",
  about: "",
  position_title: "",
  role_name: ""
};

const ALLOWED_FIELDS: Array<keyof SearchFormState> = [
  "first_name",
  "last_name",
  "middle_name",
  "login",
  "group_number",
  "institute",
  "about",
  "position_title",
  "role_name"
];

const DEFAULT_LIMIT = 20;

function buildRequestPayload(form: SearchFormState, includeArchived: boolean, limit: number, offset: number): UserSearchRequest {
  const payload: UserSearchRequest = {
    include_archived: includeArchived,
    limit,
    offset
  };

  for (const key of ALLOWED_FIELDS) {
    const value = form[key].trim();
    if (value.length > 0) {
      payload[key] = value;
    }
  }

  return payload;
}

function buildQueryString(payload: UserSearchRequest): string {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(payload)) {
    if (value === undefined || value === null) {
      continue;
    }
    params.set(key, String(value));
  }
  return params.toString();
}

export function SearchPage({ onGlobalError }: SearchPageProps) {
  const [form, setForm] = useState<SearchFormState>(DEFAULT_FORM);
  const [includeArchived, setIncludeArchived] = useState(false);
  const [items, setItems] = useState<SearchUserItem[]>([]);
  const [total, setTotal] = useState(0);
  const [limit] = useState(DEFAULT_LIMIT);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasSearched, setHasSearched] = useState(false);
  const [selectedUser, setSelectedUser] = useState<SearchUserItem | null>(null);
  const lastPayloadRef = useRef<UserSearchRequest | null>(null);

  function updateField<K extends keyof SearchFormState>(key: K, value: SearchFormState[K]) {
    setForm((current) => ({
      ...current,
      [key]: value
    }));
  }

  async function executeSearch(payload: UserSearchRequest) {
    setLoading(true);
    setError(null);
    setHasSearched(true);
    setSelectedUser(null);
    lastPayloadRef.current = payload;

    try {
      const query = buildQueryString(payload);
      const response = await request<UserSearchResponse>(`/api/v1/search/users?${query}`);
      onGlobalError?.(null);
      setItems(Array.isArray(response.items) ? response.items : []);
      setTotal(typeof response.total === "number" ? response.total : 0);
      setOffset(payload.offset ?? 0);
    } catch (nextError) {
      const apiError = adaptApiError(nextError, "Не удалось выполнить поиск.");
      setItems([]);
      setTotal(0);
      setError(apiError.message);
      onGlobalError?.({
        message: apiError.message,
        severity: "error",
        retryLabel: apiError.retryable ? "Повторить" : undefined,
        onRetry: apiError.retryable
          ? () => {
              const payloadForRetry = lastPayloadRef.current;
              if (payloadForRetry) {
                void executeSearch(payloadForRetry);
              }
            }
          : undefined
      });
    } finally {
      setLoading(false);
    }
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const payload = buildRequestPayload(form, includeArchived, limit, 0);
    void executeSearch(payload);
  }

  const searchState: AsyncState = !hasSearched
    ? "idle"
    : loading && items.length === 0 && !error
      ? "loading"
      : error
        ? "error"
        : items.length === 0
          ? "empty"
          : "success";

  function retryLastSearch() {
    const payloadForRetry = lastPayloadRef.current;
    if (payloadForRetry) {
      void executeSearch(payloadForRetry);
    }
  }

  function goToPage(nextOffset: number) {
    const payload = lastPayloadRef.current ?? buildRequestPayload(form, includeArchived, limit, nextOffset);
    void executeSearch({
      ...payload,
      limit,
      offset: Math.max(0, nextOffset)
    });
  }

  const canPrev = offset > 0;
  const canNext = offset + limit < total;
  const rangeStart = total === 0 ? 0 : offset + 1;
  const rangeEnd = total === 0 ? 0 : Math.min(offset + limit, total);

  return (
    <PageRoot data-testid="search-panel">
      <PageContent centered maxWidth={1280}>
        <h1
          style={{
            margin: "0 0 var(--space-4)",
            fontSize: "var(--text-xl)",
            fontWeight: "var(--weight-bold)",
            color: "var(--color-text-primary)"
          }}
        >
          Поиск пользователей
        </h1>

        <PageGrid sidebarWidth={380}>
          <div>
            <div
              style={{
                background: "var(--color-surface)",
                borderRadius: "var(--radius-card)",
                boxShadow: "var(--shadow-sm)",
                padding: "var(--space-5)"
              }}
            >
              <form onSubmit={handleSubmit}>
                <div style={{ display: "grid", gridTemplateColumns: "repeat(3, 1fr)", gap: "var(--space-4)" }}>
                  <div>
                    <Label htmlFor="search-first-name">Имя</Label>
                    <Input
                      data-testid="search-first_name"
                      id="search-first-name"
                      onChange={(event) => updateField("first_name", event.target.value)}
                      value={form.first_name}
                    />
                  </div>
                  <div>
                    <Label htmlFor="search-last-name">Фамилия</Label>
                    <Input
                      data-testid="search-last_name"
                      id="search-last-name"
                      onChange={(event) => updateField("last_name", event.target.value)}
                      value={form.last_name}
                    />
                  </div>
                  <div>
                    <Label htmlFor="search-middle-name">Отчество</Label>
                    <Input
                      data-testid="search-middle_name"
                      id="search-middle-name"
                      onChange={(event) => updateField("middle_name", event.target.value)}
                      value={form.middle_name}
                    />
                  </div>
                  <div>
                    <Label htmlFor="search-login">Логин</Label>
                    <Input
                      data-testid="search-login"
                      id="search-login"
                      onChange={(event) => updateField("login", event.target.value)}
                      value={form.login}
                    />
                  </div>
                  <div>
                    <Label htmlFor="search-group-number">Группа</Label>
                    <Input
                      data-testid="search-group_number"
                      id="search-group-number"
                      onChange={(event) => updateField("group_number", event.target.value)}
                      value={form.group_number}
                    />
                  </div>
                  <div>
                    <Label htmlFor="search-institute">Институт</Label>
                    <Input
                      data-testid="search-institute"
                      id="search-institute"
                      onChange={(event) => updateField("institute", event.target.value)}
                      value={form.institute}
                    />
                  </div>
                  <div>
                    <Label htmlFor="search-about">О себе</Label>
                    <Input
                      data-testid="search-about"
                      id="search-about"
                      onChange={(event) => updateField("about", event.target.value)}
                      value={form.about}
                    />
                  </div>
                  <div>
                    <Label htmlFor="search-position-title">Должность</Label>
                    <Input
                      data-testid="search-position_title"
                      id="search-position-title"
                      onChange={(event) => updateField("position_title", event.target.value)}
                      value={form.position_title}
                    />
                  </div>
                  <div>
                    <Label htmlFor="search-role-name">Роль</Label>
                    <Input
                      data-testid="search-role_name"
                      id="search-role-name"
                      onChange={(event) => updateField("role_name", event.target.value)}
                      value={form.role_name}
                    />
                  </div>
                </div>

                <label
                  htmlFor="search-include-archived"
                  style={{ display: "inline-flex", alignItems: "center", gap: "var(--space-2)", marginTop: "var(--space-3)", fontSize: "var(--text-sm)" }}
                >
                  <input
                    checked={includeArchived}
                    data-testid="search-include_archived"
                    id="search-include-archived"
                    onChange={(event) => setIncludeArchived(event.target.checked)}
                    style={{ width: 16, height: 16, accentColor: "var(--color-brand-primary)" }}
                    type="checkbox"
                  />
                  Включить архивные назначения
                </label>

                <div style={{ display: "flex", justifyContent: "flex-end", marginTop: "var(--space-4)" }}>
                  <Button
                    data-testid="search-submit"
                    disabled={loading}
                    isLoading={loading}
                    size="default"
                    type="submit"
                    variant="primary"
                  >
                    Найти
                  </Button>
                </div>
              </form>
            </div>

            <AsyncStateView
              data={{ items, total }}
              emptyView={(
                <SearchResults
                  error={null}
                  hasSearched
                  items={[]}
                  loading={false}
                  onRetry={retryLastSearch}
                  total={total}
                />
              )}
              errorMessage={error}
              idleView={(
                <SearchResults
                  error={null}
                  hasSearched={false}
                  items={[]}
                  loading={false}
                  onRetry={retryLastSearch}
                  total={0}
                />
              )}
              loadingView={(
                <SearchResults
                  error={null}
                  hasSearched
                  items={[]}
                  loading
                  onRetry={retryLastSearch}
                  total={0}
                />
              )}
              onRetry={retryLastSearch}
              state={searchState}
            >
              {(resultData) => (
                <SoftRefreshWrapper isRefreshing={loading && hasSearched}>
                  <SearchResults
                    error={null}
                    hasSearched
                    items={resultData.items}
                    loading={false}
                    onRetry={retryLastSearch}
                    onSelectUser={setSelectedUser}
                    selectedUserId={selectedUser?.id}
                    total={resultData.total}
                  />
                </SoftRefreshWrapper>
              )}
            </AsyncStateView>

            {hasSearched ? (
              <div style={{ marginTop: "var(--space-3)", display: "flex", alignItems: "center", gap: "var(--space-4)" }}>
                <Button
                  data-testid="search-pagination-prev"
                  disabled={!canPrev || loading}
                  onClick={() => goToPage(offset - limit)}
                  type="button"
                  variant="outline"
                >
                  ← Назад
                </Button>
                <span
                  data-testid="search-pagination-state"
                  style={{ flex: 1, textAlign: "center", fontSize: "var(--text-sm)", color: "var(--color-text-muted)" }}
                >
                  Записи {rangeStart}–{rangeEnd} из {total}
                </span>
                <Button
                  data-testid="search-pagination-next"
                  disabled={!canNext || loading}
                  onClick={() => goToPage(offset + limit)}
                  type="button"
                  variant="outline"
                >
                  Вперёд →
                </Button>
              </div>
            ) : null}
          </div>

          <PageSidebar sticky>
            <SearchUserDetailsPanel user={selectedUser} />
          </PageSidebar>
        </PageGrid>
      </PageContent>
    </PageRoot>
  );
}
