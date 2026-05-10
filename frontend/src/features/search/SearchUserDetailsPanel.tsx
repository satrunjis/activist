import { useCallback, useEffect, useState } from "react";
import { MousePointerClick } from "lucide-react";

import { request } from "../../shared/api/client";
import { adaptApiError } from "../../shared/api/errorAdapter";
import type { MembershipSummary, SearchUserItem } from "../../shared/api/types";
import { InlineError, SoftRefreshWrapper } from "../../shared/ui/feedback";
import { ItemRow, MetaBlock } from "../../shared/ui/layout";
import { ActionPanel } from "../../shared/ui/surfaces";

type MembershipsResponse = {
  items: MembershipSummary[];
};

type SearchUserDetailsPanelProps = {
  user: SearchUserItem | null;
};

function fieldValue(value: string | undefined): string {
  if (!value || !value.trim()) {
    return "—";
  }
  return value;
}

function formatFullName(user: SearchUserItem): string {
  const parts = [user.last_name, user.first_name, user.middle_name].filter((value) => value && value.trim().length > 0);
  return parts.length > 0 ? parts.join(" ") : user.first_name;
}

export function SearchUserDetailsPanel({ user }: SearchUserDetailsPanelProps) {
  const [memberships, setMemberships] = useState<MembershipSummary[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [membershipsUserId, setMembershipsUserId] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  const loadMemberships = useCallback(async (id: string, signal?: AbortSignal) => {
    setIsLoading(true);
    setErrorMessage(null);

    try {
      const loadedMemberships = await request<MembershipsResponse>(`/api/v1/users/${encodeURIComponent(id)}/memberships`, { signal });
      setMemberships(Array.isArray(loadedMemberships.items) ? loadedMemberships.items : []);
      setMembershipsUserId(id);
    } catch (error) {
      if (error instanceof DOMException && error.name === "AbortError") {
        return;
      }
      const apiError = adaptApiError(error, "Не удалось загрузить назначения пользователя.");
      setErrorMessage(apiError.message);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!user) {
      setMemberships([]);
      setMembershipsUserId(null);
      setErrorMessage(null);
      setIsLoading(false);
      return;
    }

    const controller = new AbortController();
    void loadMemberships(user.id, controller.signal);
    return () => controller.abort();
  }, [loadMemberships, user]);

  const title = user ? formatFullName(user) : "Детали пользователя";
  const subtitle = user?.login?.trim() || user?.id;

  return (
    <ActionPanel subtitle={subtitle} title={title}>
      <InlineError message={errorMessage} />

      {!user ? (
        <div
          style={{
            alignItems: "center",
            display: "flex",
            flexDirection: "column",
            justifyContent: "center",
            minHeight: 240,
            padding: "var(--space-8) var(--space-4)",
            textAlign: "center"
          }}
        >
          <span
            aria-hidden
            style={{
              alignItems: "center",
              color: "var(--color-text-muted)",
              display: "inline-flex",
              height: 36,
              justifyContent: "center",
              marginBottom: "var(--space-2)",
              width: 36
            }}
          >
            <MousePointerClick size={28} />
          </span>
          <p style={{ margin: 0, fontSize: "var(--text-sm)", color: "var(--color-text-muted)" }}>
            Выберите пользователя в результатах поиска, чтобы открыть карточку.
          </p>
        </div>
      ) : null}

      {user ? (
        <div data-testid="search-user-details-panel">
          <MetaBlock
            data-testid="search-user-details-meta"
            fields={[
              { label: "ID", value: user.id },
              { label: "Логин", value: fieldValue(user.login) },
              { label: "Имя", value: fieldValue(user.first_name) },
              { label: "Фамилия", value: fieldValue(user.last_name) },
              { label: "Отчество", value: fieldValue(user.middle_name) },
              { label: "Номер зачётки", value: fieldValue(user.gradebook_number) },
              { label: "Группа", value: fieldValue(user.group_number) },
              { label: "Институт", value: fieldValue(user.institute) },
              { label: "Дата рождения", value: fieldValue(user.birth_date) },
              { label: "Телефон", value: fieldValue(user.phone) },
              { label: "О себе", value: fieldValue(user.about) }
            ]}
          />

          <div style={{ marginTop: "var(--space-5)" }}>
            <h3 style={{ margin: "0 0 var(--space-2)", fontSize: "var(--text-sm)", fontWeight: "var(--weight-semibold)", color: "var(--color-text-primary)" }}>
              Назначения
            </h3>
            {isLoading && membershipsUserId !== user.id ? (
              <p style={{ margin: 0, fontSize: "var(--text-sm)", color: "var(--color-text-muted)" }}>
                Загрузка назначений…
              </p>
            ) : (
              <SoftRefreshWrapper isRefreshing={isLoading}>
                {memberships.length === 0 ? (
                  <p style={{ margin: 0, fontSize: "var(--text-sm)", color: "var(--color-text-muted)" }}>
                    Назначений нет.
                  </p>
                ) : (
                  <div style={{ border: "1px solid var(--color-border)", borderRadius: "var(--radius-card)", overflow: "hidden" }}>
                    {memberships.map((membership) => (
                      <ItemRow
                        data-testid={`search-user-membership-${membership.position_id}`}
                        key={`${membership.position_id}:${membership.division_id}:${membership.role_name}`}
                      >
                        <div>
                          <div style={{ fontSize: "var(--text-sm)", fontWeight: "var(--weight-semibold)", color: "var(--color-text-primary)" }}>
                            {membership.position_name}
                          </div>
                          <div style={{ marginTop: 2, fontSize: "var(--text-xs)", color: "var(--color-text-muted)" }}>
                            {membership.division_name} · {membership.role_name}
                          </div>
                        </div>
                      </ItemRow>
                    ))}
                  </div>
                )}
              </SoftRefreshWrapper>
            )}
          </div>
        </div>
      ) : null}
    </ActionPanel>
  );
}
