import { EmptyStateCard, SectionSkeleton } from "../../shared/ui/feedback";
import { AsyncStateView, type AsyncState } from "../../shared/ui/states/AsyncStateView";
import { ItemRow } from "../../shared/ui/layout";
import { request } from "../../shared/api/client";
import { adaptApiError } from "../../shared/api/errorAdapter";
import type { PositionMemberItem, UserProfile } from "../../shared/api/types";
import { useEffect, useMemo, useState } from "react";

type PositionMembersListProps = {
  positionId: string | null;
  state: "idle" | "loading" | "loaded" | "error";
  errorMessage: string | null;
  items: PositionMemberItem[];
};

function formatFullName(profile: UserProfile): string {
  const parts = [profile.last_name, profile.first_name, profile.middle_name].filter((value) => value && value.trim().length > 0);
  return parts.length > 0 ? parts.join(" ") : profile.first_name;
}

function getInitials(userId: string, profile?: UserProfile): string {
  if (profile) {
    const first = profile.first_name?.trim().charAt(0) ?? "";
    const last = profile.last_name?.trim().charAt(0) ?? "";
    const initials = `${first}${last}`.toUpperCase();
    if (initials) {
      return initials;
    }
  }

  return userId.slice(0, 2).toUpperCase();
}

function profileLine(profile: UserProfile | undefined, fallback: string): string {
  if (!profile) {
    return fallback;
  }

  return profile.login?.trim() ? profile.login : fallback;
}

function memberChips(item: PositionMemberItem, profile?: UserProfile): string[] {
  return [
    item.role_name,
    profile?.group_number,
    profile?.institute
  ].filter((value): value is string => Boolean(value?.trim()));
}

export function PositionMembersList({ positionId, state, errorMessage, items }: PositionMembersListProps) {
  const [profilesById, setProfilesById] = useState<Record<string, UserProfile>>({});
  const [profileError, setProfileError] = useState<string | null>(null);
  const [isProfileLoading, setProfileLoading] = useState(false);

  const memberIdsSignature = useMemo(
    () => items.map((item) => item.user_id).sort().join("|"),
    [items]
  );

  useEffect(() => {
    setProfilesById({});
    setProfileError(null);

    if (!positionId || state !== "loaded" || items.length === 0) {
      setProfileLoading(false);
      return;
    }

    const controller = new AbortController();
    const uniqueUserIds = Array.from(new Set(items.map((item) => item.user_id)));

    setProfileLoading(true);

    void Promise.allSettled(
      uniqueUserIds.map(async (userId) => {
        const profile = await request<UserProfile>(`/api/v1/users/${encodeURIComponent(userId)}`, {
          signal: controller.signal
        });
        return [userId, profile] as const;
      })
    ).then((results) => {
      if (controller.signal.aborted) {
        return;
      }

      const nextProfiles: Record<string, UserProfile> = {};
      let firstError: unknown = null;

      for (const result of results) {
        if (result.status === "fulfilled") {
          const [userId, profile] = result.value;
          nextProfiles[userId] = profile;
        } else if (!firstError) {
          firstError = result.reason;
        }
      }

      setProfilesById(nextProfiles);
      setProfileError(firstError ? adaptApiError(firstError, "Не удалось загрузить часть профилей.").message : null);
    }).finally(() => {
      if (!controller.signal.aborted) {
        setProfileLoading(false);
      }
    });

    return () => controller.abort();
  }, [items, memberIdsSignature, positionId, state]);

  const membersState: AsyncState = !positionId || state === "idle"
    ? "idle"
    : state === "loading"
      ? "loading"
      : state === "error"
        ? "error"
        : items.length === 0
          ? "empty"
          : "success";

  return (
    <section data-testid="position-members-list">
      <AsyncStateView
        data={items}
        emptyView={<EmptyStateCard heading="Нет участников" />}
        errorMessage={errorMessage ?? "Не удалось загрузить участников."}
        idleView={(
          <p
            className="muted-text"
            style={{ fontSize: "var(--text-sm)", color: "var(--color-text-muted)", textAlign: "center" }}
          >
            Выберите должность выше, чтобы увидеть участников
          </p>
        )}
        loadingView={<SectionSkeleton rows={2} />}
        state={membersState}
      >
        {(loadedItems) => (
          <>
            {isProfileLoading ? (
              <p style={{ margin: "0 0 var(--space-2)", fontSize: "var(--text-xs)", color: "var(--color-text-muted)" }}>
                Загрузка профилей…
              </p>
            ) : null}
            {profileError ? (
              <p style={{ margin: "0 0 var(--space-2)", fontSize: "var(--text-xs)", color: "var(--color-error-text)" }}>
                {profileError}
              </p>
            ) : null}
            <ul
              data-testid="position-members-items"
              style={{ listStyle: "none", margin: "0", padding: "0" }}
            >
              {loadedItems.map((item) => {
                const profile = profilesById[item.user_id];
                const chips = memberChips(item, profile);

                return (
                  <ItemRow
                    as="li"
                    data-testid={`position-member-${item.user_id}`}
                    key={`${item.position_id}-${item.user_id}`}
                  >
                    <div style={{ display: "flex", alignItems: "flex-start", gap: "var(--space-3)" }}>
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
                        {getInitials(item.user_id, profile)}
                      </div>

                      <div style={{ flex: 1, minWidth: 0 }}>
                        <p
                          style={{
                            margin: 0,
                            fontSize: "var(--text-sm)",
                            fontWeight: "var(--weight-semibold)",
                            color: "var(--color-text-primary)",
                            overflow: "hidden",
                            textOverflow: "ellipsis",
                            whiteSpace: "nowrap"
                          }}
                        >
                          {profile ? formatFullName(profile) : item.user_id}
                        </p>
                        <p
                          style={{
                            margin: "2px 0 0",
                            fontSize: "var(--text-xs)",
                            color: "var(--color-text-muted)",
                            overflow: "hidden",
                            textOverflow: "ellipsis",
                            whiteSpace: "nowrap"
                          }}
                        >
                          {profileLine(profile, item.user_id)}
                        </p>
                        <p
                          data-testid={`position-member-id-${item.user_id}`}
                          style={{
                            margin: "2px 0 0",
                            fontSize: "var(--text-xs)",
                            color: "var(--color-text-muted)",
                            overflow: "hidden",
                            textOverflow: "ellipsis",
                            whiteSpace: "nowrap"
                          }}
                        >
                          ID: {item.user_id}
                        </p>

                        {chips.length > 0 ? (
                          <div style={{ display: "flex", alignItems: "center", gap: "var(--space-2)", flexWrap: "wrap", marginTop: "var(--space-2)" }}>
                            {chips.map((chip) => (
                              <span
                                key={chip}
                                style={{
                                  fontSize: "var(--text-xs)",
                                  border: "1px solid var(--color-border)",
                                  borderRadius: "var(--radius-pill)",
                                  padding: "2px 8px",
                                  color: "var(--color-text-primary)",
                                  background: "var(--color-surface)",
                                  maxWidth: "100%",
                                  overflow: "hidden",
                                  textOverflow: "ellipsis",
                                  whiteSpace: "nowrap"
                                }}
                              >
                                {chip}
                              </span>
                            ))}
                          </div>
                        ) : null}
                      </div>
                    </div>
                  </ItemRow>
                );
              })}
            </ul>
          </>
        )}
      </AsyncStateView>
    </section>
  );
}
