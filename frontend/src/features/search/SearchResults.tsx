import { AlertCircle, SearchX } from "lucide-react";
import type { SearchUserItem } from "../../shared/api/types";
import { EmptyStateCard, SectionSkeleton } from "../../shared/ui/feedback";
import { ItemRow } from "../../shared/ui/layout";

type SearchResultsProps = {
  items: SearchUserItem[];
  loading: boolean;
  error: string | null;
  total: number;
  hasSearched: boolean;
  selectedUserId?: string | null;
  onSelectUser?: (item: SearchUserItem) => void;
  onRetry: () => void;
};

type MembershipChip = {
  id: string;
  label: string;
};

function formatFullName(item: SearchUserItem): string {
  const parts = [item.last_name, item.first_name, item.middle_name].filter((value) => value && value.trim().length > 0);
  return parts.length > 0 ? parts.join(" ") : item.first_name;
}

function getInitials(item: SearchUserItem): string {
  const first = item.first_name?.trim().charAt(0) ?? "";
  const last = item.last_name?.trim().charAt(0) ?? "";
  const initials = `${first}${last}`.toUpperCase();
  return initials || "??";
}

function getMembershipChips(item: SearchUserItem): MembershipChip[] {
  const memberships = Array.isArray(item.memberships) ? item.memberships : [];
  return memberships.map((membership) => ({
    id: `${membership.position_name}:${membership.division_name}:${membership.role_name}`,
    label: membership.position_name
  }));
}

export function SearchResults({
  items,
  loading,
  error,
  total,
  hasSearched,
  selectedUserId,
  onSelectUser,
  onRetry
}: SearchResultsProps) {
  const showInitialLoading = loading && items.length === 0;

  if (!hasSearched) {
    return null;
  }

  if (showInitialLoading) {
    return (
      <div data-testid="search-results-loading" style={{ marginTop: "var(--space-4)" }}>
        <SectionSkeleton rows={5} />
      </div>
    );
  }

  if (error) {
    return (
      <div data-testid="search-results-error" style={{ marginTop: "var(--space-4)" }}>
        <EmptyStateCard
          body={error}
          cta={{ label: "Повторить", onClick: onRetry }}
          heading="Ошибка поиска"
          icon={<AlertCircle />}
        />
      </div>
    );
  }

  if (items.length === 0) {
    return (
      <div data-testid="search-results-empty" style={{ marginTop: "var(--space-4)" }}>
        <EmptyStateCard body="Попробуйте изменить параметры поиска." heading="Ничего не найдено" icon={<SearchX />} />
      </div>
    );
  }

  return (
    <section data-testid="search-results">
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: "var(--space-3)" }}>
        <p
          style={{
            margin: "var(--space-4) 0 var(--space-2)",
            fontSize: "var(--text-sm)",
            color: "var(--color-text-muted)"
          }}
        >
          Найдено: {total}
        </p>
        {loading ? (
          <span style={{ fontSize: "var(--text-xs)", color: "var(--color-text-muted)" }}>
            Обновление…
          </span>
        ) : null}
      </div>

      <ul
        style={{
          listStyle: "none",
          margin: 0,
          padding: 0,
          opacity: loading ? 0.75 : 1,
          transition: "opacity var(--transition-fast)"
        }}
      >
        {items.map((item) => {
          const chips = getMembershipChips(item);
          const visibleChips = chips.slice(0, 3);
          const overflowCount = Math.max(0, chips.length - visibleChips.length);

          return (
            <ItemRow
              as="li"
              data-testid={`search-result-${item.id}`}
              key={item.id}
              onClick={onSelectUser ? () => onSelectUser(item) : undefined}
              selected={selectedUserId === item.id}
            >
              <div style={{ display: "flex", alignItems: "center", gap: "var(--space-3)" }}>
              <div
                aria-hidden
                style={{
                  width: 36,
                  height: 36,
                  borderRadius: "var(--radius-full)",
                  background: "var(--color-brand-primary)",
                  color: "var(--color-text-inverse)",
                  display: "inline-flex",
                  alignItems: "center",
                  justifyContent: "center",
                  fontSize: "var(--text-xs)",
                  fontWeight: "var(--weight-bold)",
                  flexShrink: 0
                }}
              >
                {getInitials(item)}
              </div>

              <div style={{ flex: 1, minWidth: 0 }}>
                <p style={{ margin: 0, fontSize: "var(--text-sm)", fontWeight: "var(--weight-semibold)", color: "var(--color-text-primary)" }}>
                  {formatFullName(item)}
                </p>
                <p style={{ margin: "2px 0 0", fontSize: "var(--text-xs)", color: "var(--color-text-muted)" }}>
                  {item.login?.trim() ? item.login : "—"}
                </p>
                <p
                  data-testid={`search-result-id-${item.id}`}
                  style={{
                    margin: "2px 0 0",
                    fontSize: "var(--text-xs)",
                    color: "var(--color-text-muted)",
                    overflowWrap: "anywhere"
                  }}
                >
                  ID: {item.id}
                </p>
              </div>

              <div style={{ display: "flex", alignItems: "center", justifyContent: "flex-end", gap: "var(--space-2)", flexWrap: "wrap" }}>
                {visibleChips.map((chip) => (
                  <span
                    key={chip.id}
                    style={{
                      fontSize: "var(--text-xs)",
                      border: "1px solid var(--color-border)",
                      borderRadius: "var(--radius-pill)",
                      padding: "2px 8px",
                      color: "var(--color-text-primary)",
                      background: "var(--color-surface)"
                    }}
                  >
                    {chip.label}
                  </span>
                ))}
                {overflowCount > 0 ? (
                  <span
                    style={{
                      fontSize: "var(--text-xs)",
                      border: "1px solid var(--color-border)",
                      borderRadius: "var(--radius-pill)",
                      padding: "2px 8px",
                      color: "var(--color-text-muted)",
                      background: "var(--color-surface-subtle)"
                    }}
                  >
                    +{overflowCount}
                  </span>
                ) : null}
              </div>
              </div>
            </ItemRow>
          );
        })}
      </ul>
    </section>
  );
}
