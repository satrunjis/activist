import { AlertCircle, Plus, Trash2, Users } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { CSSProperties, ReactNode } from "react";

import { Button } from "../../components/ui/button";
import { Input } from "../../components/ui/input";
import { request } from "../../shared/api/client";
import { adaptApiError } from "../../shared/api/errorAdapter";
import type { SocialLink, UserProfile } from "../../shared/api/types";
import { ConfirmDialog, EmptyStateCard, SectionSkeleton, useToast } from "../../shared/ui/feedback";
import { ItemRow, MetaBlock, PageContent, PageGrid, PageHeader, PageRoot, PageSidebar } from "../../shared/ui/layout";
import { AsyncStateView } from "../../shared/ui/states/AsyncStateView";

type MembershipItem = {
  position_id: string;
  position_name: string;
  division_id: string;
  division_name: string;
  role_name: string;
};

type MembershipsResponse = {
  items: MembershipItem[];
};

type ProfilePageProps = {
  userId: string;
};

type ProfileDraft = {
  first_name: string;
  last_name: string;
  middle_name: string;
  birth_date: string;
  gradebook_number: string;
  group_number: string;
  institute: string;
  phone: string;
  about: string;
  social_links: SocialLink[];
};

type RequiredProfileField = "first_name" | "birth_date" | "gradebook_number" | "group_number" | "institute";

type ProfileValidationErrors = Partial<Record<RequiredProfileField, string>>;
type SocialLinkValidationErrors = Record<number, Partial<Record<keyof SocialLink, string>>>;

function getMembershipKey(item: MembershipItem, userId: string): string {
  return `${item.position_id}:${userId}`;
}

function stringifyValue(value: string | undefined): string {
  if (!value || !value.trim()) {
    return "—";
  }
  return value;
}

function formatProfileDate(value: string | undefined): string {
  const trimmed = value?.trim() ?? "";
  const dateMatch = trimmed.match(/^(\d{4}-\d{2}-\d{2})/);
  return dateMatch?.[1] ?? trimmed;
}

function normalizeSocialLinks(links: SocialLink[] | undefined): SocialLink[] {
  return (links ?? [])
    .map((entry) => ({
      platform: entry.platform.trim(),
      value: entry.value.trim()
    }))
    .filter((entry) => entry.platform.length > 0 && entry.value.length > 0);
}

function requiredLabel(label: string): ReactNode {
  return (
    <>
      {label}
      <span aria-hidden="true" style={{ color: "var(--color-error)" }}> *</span>
    </>
  );
}

const readonlyFieldValueStyle: CSSProperties = {
  minHeight: 36,
  display: "flex",
  alignItems: "center",
  padding: "0 var(--space-3)",
  border: "1px solid transparent",
  borderRadius: "var(--radius-input)",
  fontSize: "var(--text-sm)",
  color: "var(--color-text-primary)"
};

const validationMessageStyle: CSSProperties = {
  color: "var(--color-error-text)",
  fontSize: "var(--text-xs)",
  fontWeight: "var(--weight-medium)",
  margin: "4px 0 0"
};

function toProfileDraft(profile: UserProfile): ProfileDraft {
  return {
    first_name: profile.first_name ?? "",
    last_name: profile.last_name ?? "",
    middle_name: profile.middle_name ?? "",
    birth_date: formatProfileDate(profile.birth_date),
    gradebook_number: profile.gradebook_number ?? "",
    group_number: profile.group_number ?? "",
    institute: profile.institute ?? "",
    phone: profile.phone ?? "",
    about: profile.about ?? "",
    social_links: normalizeSocialLinks(profile.social_links)
  };
}

export function ProfilePage({ userId }: ProfilePageProps) {
  const { showToast } = useToast();
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [memberships, setMemberships] = useState<MembershipItem[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isMembershipsLoading, setIsMembershipsLoading] = useState(true);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [draft, setDraft] = useState<ProfileDraft | null>(null);
  const [validationErrors, setValidationErrors] = useState<ProfileValidationErrors>({});
  const [socialLinkErrors, setSocialLinkErrors] = useState<SocialLinkValidationErrors>({});
  const [isSaving, setIsSaving] = useState(false);
  const [confirmRemoveId, setConfirmRemoveId] = useState<string | null>(null);
  const [fadingOutKeys, setFadingOutKeys] = useState<string[]>([]);
  const fadeTimersRef = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map());

  const loadProfile = useCallback(async () => {
    setErrorMessage(null);
    setIsLoading(true);
    setIsMembershipsLoading(true);

    try {
      const [loadedProfile, loadedMemberships] = await Promise.all([
        request<UserProfile>(`/api/v1/users/${encodeURIComponent(userId)}`),
        request<MembershipsResponse>(`/api/v1/users/${encodeURIComponent(userId)}/memberships`)
      ]);
      setProfile(loadedProfile);
      setMemberships(Array.isArray(loadedMemberships.items) ? loadedMemberships.items : []);
    } catch (error) {
      setErrorMessage(adaptApiError(error, "Не удалось загрузить профиль.").message);
    } finally {
      setIsLoading(false);
      setIsMembershipsLoading(false);
    }
  }, [userId]);

  useEffect(() => {
    void loadProfile();
  }, [loadProfile]);

  useEffect(() => {
    const timers = fadeTimersRef.current;
    return () => {
      for (const timer of timers.values()) {
        clearTimeout(timer);
      }
      timers.clear();
    };
  }, []);

  async function handleRemove(positionId: string, userIdValue: string) {
    const key = `${positionId}:${userIdValue}`;

    try {
      await request(`/api/v1/positions/${encodeURIComponent(positionId)}/members/${encodeURIComponent(userIdValue)}`, {
        method: "DELETE"
      });

      setFadingOutKeys((current) => (current.includes(key) ? current : [...current, key]));

      const timer = setTimeout(() => {
        setMemberships((current) => current.filter((item) => getMembershipKey(item, userId) !== key));
        setFadingOutKeys((current) => current.filter((item) => item !== key));
        fadeTimersRef.current.delete(key);
      }, 150);
      fadeTimersRef.current.set(key, timer);

      showToast("success", "Участие удалено");
    } catch (error) {
      const apiError = adaptApiError(error, "Не удалось исключить из должности.");
      const message = apiError.code?.startsWith("access.scope")
        ? "Операция недоступна в текущем контексте."
        : apiError.kind === "forbidden" || apiError.code?.startsWith("access.")
          ? "Операция запрещена."
          : apiError.message;
      showToast("error", "Не удалось исключить из должности", message);
    }
  }

  const fullName = useMemo(() => {
    if (!profile) {
      return "";
    }
    const source = isEditing && draft ? draft : profile;
    const parts = [source.last_name, source.first_name, source.middle_name].filter(
      (value) => typeof value === "string" && value.trim().length > 0
    );
    return parts.length > 0 ? parts.join(" ") : source.first_name;
  }, [draft, isEditing, profile]);

  function updateDraft<K extends keyof ProfileDraft>(key: K, value: ProfileDraft[K]) {
    setDraft((current) => (current ? { ...current, [key]: value } : current));
    if (key in validationErrors) {
      setValidationErrors((current) => {
        const next = { ...current };
        delete next[key as RequiredProfileField];
        return next;
      });
    }
    if (key === "social_links") {
      setSocialLinkErrors({});
    }
  }

  function validateDraft(nextDraft: ProfileDraft): ProfileValidationErrors {
    const nextErrors: ProfileValidationErrors = {};
    if (!nextDraft.first_name.trim()) {
      nextErrors.first_name = "Имя обязательно.";
    }
    if (!nextDraft.birth_date.trim()) {
      nextErrors.birth_date = "Дата рождения обязательна.";
    }
    if (!nextDraft.gradebook_number.trim()) {
      nextErrors.gradebook_number = "Зачётная книжка обязательна.";
    }
    if (!nextDraft.group_number.trim()) {
      nextErrors.group_number = "Номер группы обязателен.";
    }
    if (!nextDraft.institute.trim()) {
      nextErrors.institute = "Институт обязателен.";
    }
    return nextErrors;
  }

  function validateSocialLinks(links: SocialLink[]): SocialLinkValidationErrors {
    return links.reduce<SocialLinkValidationErrors>((nextErrors, link, index) => {
      const hasPlatform = link.platform.trim().length > 0;
      const hasValue = link.value.trim().length > 0;
      if (!hasPlatform && !hasValue) {
        return nextErrors;
      }
      if (!hasPlatform) {
        nextErrors[index] = { ...nextErrors[index], platform: "Укажите платформу." };
      }
      if (!hasValue) {
        nextErrors[index] = { ...nextErrors[index], value: "Укажите ссылку." };
      }
      return nextErrors;
    }, {});
  }

  function hasProfileChanges(currentProfile: UserProfile, nextDraft: ProfileDraft): boolean {
    const normalizeOptional = (value: string | undefined) => value?.trim() ?? "";
    const normalizeLinks = (links: SocialLink[] | undefined) =>
      JSON.stringify(normalizeSocialLinks(links));

    return (
      normalizeOptional(currentProfile.first_name) !== nextDraft.first_name.trim() ||
      normalizeOptional(currentProfile.last_name) !== nextDraft.last_name.trim() ||
      normalizeOptional(currentProfile.middle_name) !== nextDraft.middle_name.trim() ||
      formatProfileDate(currentProfile.birth_date) !== nextDraft.birth_date.trim() ||
      normalizeOptional(currentProfile.gradebook_number) !== nextDraft.gradebook_number.trim() ||
      normalizeOptional(currentProfile.group_number) !== nextDraft.group_number.trim() ||
      normalizeOptional(currentProfile.institute) !== nextDraft.institute.trim() ||
      normalizeOptional(currentProfile.phone) !== nextDraft.phone.trim() ||
      normalizeOptional(currentProfile.about) !== nextDraft.about.trim() ||
      normalizeLinks(currentProfile.social_links) !== normalizeLinks(nextDraft.social_links)
    );
  }

  async function handleSaveProfile() {
    if (!profile || !draft || isSaving) {
      return;
    }

    const nextErrors = validateDraft(draft);
    const nextSocialLinkErrors = validateSocialLinks(draft.social_links);
    setValidationErrors(nextErrors);
    setSocialLinkErrors(nextSocialLinkErrors);
    if (Object.keys(nextErrors).length > 0) {
      showToast("error", "Заполните обязательные поля", Object.values(nextErrors)[0]);
      return;
    }
    if (Object.keys(nextSocialLinkErrors).length > 0) {
      showToast("error", "Проверьте социальные сети", "У каждой добавленной ссылки должны быть платформа и ссылка.");
      return;
    }
    if (!hasProfileChanges(profile, draft)) {
      showToast("info", "Нет изменений для сохранения");
      return;
    }

    setIsSaving(true);

    try {
      const socialLinks = normalizeSocialLinks(draft.social_links);

      const updated = await request<UserProfile>(`/api/v1/users/${encodeURIComponent(userId)}`, {
        method: "PATCH",
        body: {
          first_name: draft.first_name.trim(),
          gradebook_number: draft.gradebook_number.trim(),
          group_number: draft.group_number.trim(),
          institute: draft.institute.trim(),
          birth_date: draft.birth_date.trim(),
          last_name: draft.last_name.trim(),
          middle_name: draft.middle_name.trim(),
          phone: draft.phone.trim(),
          about: draft.about.trim(),
          social_links: socialLinks
        }
      });
      setProfile(updated);
      setDraft(null);
      setValidationErrors({});
      setSocialLinkErrors({});
      setIsEditing(false);
      showToast("success", "Профиль обновлён");
    } catch (error) {
      showToast("error", "Не удалось обновить профиль", adaptApiError(error, "Не удалось обновить профиль.").message);
    } finally {
      setIsSaving(false);
    }
  }

  function fieldValue(text: string | undefined) {
    return (
      <span style={readonlyFieldValueStyle}>
        {stringifyValue(text)}
      </span>
    );
  }

  function fieldError(message: string | undefined) {
    return message ? <p style={validationMessageStyle}>{message}</p> : null;
  }

  if (isLoading) {
    return (
      <PageRoot>
        <PageHeader title="Профиль" />
        <PageContent>
          <AsyncStateView
            data={true}
            loadingView={<SectionSkeleton rows={6} withHeader />}
            state="loading"
          >
            {() => null}
          </AsyncStateView>
        </PageContent>
      </PageRoot>
    );
  }

  if (errorMessage || !profile) {
    return (
      <PageRoot>
        <PageHeader title="Профиль" />
        <PageContent>
          <EmptyStateCard
            body={errorMessage ?? "Не удалось загрузить профиль."}
            cta={{ label: "Повторить", onClick: () => void loadProfile() }}
            heading="Ошибка загрузки"
            icon={<AlertCircle />}
          />
        </PageContent>
      </PageRoot>
    );
  }

  const profileSocialLinks = normalizeSocialLinks(profile.social_links);

  return (
    <PageRoot>
      <PageHeader title="Профиль" />
      <PageContent>
        <PageGrid>
          <div>
            <section
              style={{
                background: "var(--color-surface)",
                borderRadius: "var(--radius-card)",
                boxShadow: "var(--shadow-sm)",
                padding: "var(--space-6)"
              }}
            >
              <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: "var(--space-4)" }}>
                <h2 style={{ margin: 0, fontSize: "var(--text-lg)", fontWeight: "var(--weight-bold)", color: "var(--color-text-primary)" }}>
                  {fullName}
                </h2>
                {!isEditing ? (
                  <Button
                    onClick={() => {
                      setDraft(toProfileDraft(profile));
                      setValidationErrors({});
                      setSocialLinkErrors({});
                      setIsEditing(true);
                    }}
                    size="default"
                    type="button"
                    variant="primary"
                  >
                    Редактировать
                  </Button>
                ) : (
                  <div style={{ display: "flex", alignItems: "center", gap: "var(--space-2)" }}>
                    <Button
                      disabled={isSaving}
                      onClick={() => void handleSaveProfile()}
                      size="default"
                      type="button"
                      variant="primary"
                    >
                      Сохранить
                    </Button>
                    <Button
                      disabled={isSaving}
                      onClick={() => {
                        setDraft(null);
                        setValidationErrors({});
                        setSocialLinkErrors({});
                        setIsEditing(false);
                      }}
                      size="default"
                      type="button"
                      variant="outline"
                    >
                      Отмена
                    </Button>
                  </div>
                )}
              </div>

              <div style={{ marginTop: "var(--space-4)" }}>
                <MetaBlock
                  columns={2}
                  fields={[
                    { label: "Логин", value: fieldValue(profile.login) },
                    {
                      label: requiredLabel("Имя"),
                      value: isEditing
                        ? (
                          <div>
                            <Input
                              aria-invalid={Boolean(validationErrors.first_name)}
                              onChange={(event) => updateDraft("first_name", event.target.value)}
                              value={draft?.first_name ?? ""}
                            />
                            {fieldError(validationErrors.first_name)}
                          </div>
                        )
                        : fieldValue(profile.first_name)
                    },
                    {
                      label: "Фамилия",
                      value: isEditing
                        ? (
                          <Input
                            onChange={(event) => updateDraft("last_name", event.target.value)}
                            value={draft?.last_name ?? ""}
                          />
                        )
                        : fieldValue(profile.last_name)
                    },
                    {
                      label: "Отчество",
                      value: isEditing
                        ? (
                          <Input
                            onChange={(event) => updateDraft("middle_name", event.target.value)}
                            value={draft?.middle_name ?? ""}
                          />
                        )
                        : fieldValue(profile.middle_name)
                    },
                    {
                      label: requiredLabel("Дата рождения"),
                      value: isEditing
                        ? (
                          <div>
                            <Input
                              aria-invalid={Boolean(validationErrors.birth_date)}
                              onChange={(event) => updateDraft("birth_date", event.target.value)}
                              placeholder="ГГГГ-ММ-ДД"
                              type="text"
                              value={draft?.birth_date ?? ""}
                            />
                            {fieldError(validationErrors.birth_date)}
                          </div>
                        )
                        : fieldValue(formatProfileDate(profile.birth_date))
                    },
                    {
                      label: requiredLabel("Зачётная книжка"),
                      value: isEditing
                        ? (
                          <div>
                            <Input
                              aria-invalid={Boolean(validationErrors.gradebook_number)}
                              onChange={(event) => updateDraft("gradebook_number", event.target.value)}
                              value={draft?.gradebook_number ?? ""}
                            />
                            {fieldError(validationErrors.gradebook_number)}
                          </div>
                        )
                        : fieldValue(profile.gradebook_number)
                    },
                    {
                      label: requiredLabel("Группа"),
                      value: isEditing
                        ? (
                          <div>
                            <Input
                              aria-invalid={Boolean(validationErrors.group_number)}
                              onChange={(event) => updateDraft("group_number", event.target.value)}
                              value={draft?.group_number ?? ""}
                            />
                            {fieldError(validationErrors.group_number)}
                          </div>
                        )
                        : fieldValue(profile.group_number)
                    },
                    {
                      label: requiredLabel("Институт"),
                      value: isEditing
                        ? (
                          <div>
                            <Input
                              aria-invalid={Boolean(validationErrors.institute)}
                              onChange={(event) => updateDraft("institute", event.target.value)}
                              value={draft?.institute ?? ""}
                            />
                            {fieldError(validationErrors.institute)}
                          </div>
                        )
                        : fieldValue(profile.institute)
                    },
                    {
                      label: "Телефон",
                      value: isEditing
                        ? (
                          <Input
                            onChange={(event) => updateDraft("phone", event.target.value)}
                            value={draft?.phone ?? ""}
                          />
                        )
                        : fieldValue(profile.phone)
                    },
                    {
                      label: "О себе",
                      value: isEditing
                        ? (
                          <Input
                            onChange={(event) => updateDraft("about", event.target.value)}
                            value={draft?.about ?? ""}
                          />
                        )
                        : (
                          <span style={{ ...readonlyFieldValueStyle, whiteSpace: "pre-wrap" }}>
                            {stringifyValue(profile.about)}
                          </span>
                        )
                    },
                    {
                      label: "Социальные сети",
                      value: isEditing ? (
                        <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-2)" }}>
                          <div style={{ minHeight: 36, display: "flex", alignItems: "center" }}>
                            <Button
                              onClick={() => updateDraft("social_links", [...(draft?.social_links ?? []), { platform: "", value: "" }])}
                              size="sm"
                              type="button"
                              variant="ghost"
                            >
                              <Plus aria-hidden size={14} />
                              <span>Добавить ссылку</span>
                            </Button>
                          </div>
                          {(draft?.social_links ?? []).map((link, index) => (
                            <div
                              key={index}
                              style={{
                                alignItems: "flex-start",
                                display: "flex",
                                gap: "var(--space-2)"
                              }}
                            >
                              <div style={{ width: 120 }}>
                                <Input
                                  aria-invalid={Boolean(socialLinkErrors[index]?.platform)}
                                  aria-label="Платформа"
                                  onChange={(event) => {
                                    const nextLinks = [...(draft?.social_links ?? [])];
                                    nextLinks[index] = { ...nextLinks[index], platform: event.target.value };
                                    updateDraft("social_links", nextLinks);
                                  }}
                                  placeholder="Платформа"
                                  value={link.platform}
                                />
                                {fieldError(socialLinkErrors[index]?.platform)}
                              </div>
                              <div style={{ flex: 1 }}>
                                <Input
                                  aria-invalid={Boolean(socialLinkErrors[index]?.value)}
                                  aria-label="Ссылка"
                                  onChange={(event) => {
                                    const nextLinks = [...(draft?.social_links ?? [])];
                                    nextLinks[index] = { ...nextLinks[index], value: event.target.value };
                                    updateDraft("social_links", nextLinks);
                                  }}
                                  placeholder="https://..."
                                  type="url"
                                  value={link.value}
                                />
                                {fieldError(socialLinkErrors[index]?.value)}
                              </div>
                              <Button
                                aria-label="Удалить ссылку"
                                onClick={() => updateDraft("social_links", (draft?.social_links ?? []).filter((_, itemIndex) => itemIndex !== index))}
                                style={{ marginTop: 1, width: 36 }}
                                type="button"
                                variant="ghost"
                              >
                                <Trash2 size={14} />
                              </Button>
                            </div>
                          ))}
                        </div>
                      ) : (
                        profileSocialLinks.length > 0 ? (
                          <div style={{ minHeight: 36, display: "flex", alignItems: "center", flexWrap: "wrap", gap: "var(--space-2)" }}>
                            {profileSocialLinks.map((link) => (
                              <a
                                href={link.value}
                                key={`${link.platform}:${link.value}`}
                                rel="noreferrer"
                                style={{
                                  fontSize: "var(--text-xs)",
                                  border: "1px solid var(--color-border)",
                                  borderRadius: "var(--radius-pill)",
                                  padding: "2px 10px",
                                  color: "var(--color-text-link)",
                                  textDecoration: "none"
                                }}
                                target="_blank"
                              >
                                {link.platform}
                              </a>
                            ))}
                          </div>
                        ) : fieldValue(undefined)
                      )
                    }
                  ]}
                />
              </div>
            </section>
          </div>

          <PageSidebar>
            <div
              data-testid="memberships-card"
              style={{
                background: "var(--color-surface)",
                borderRadius: "var(--radius-card)",
                boxShadow: "var(--shadow-sm)",
                padding: "var(--space-4)"
              }}
            >
              <h3 style={{ margin: 0, fontSize: "var(--text-base)", fontWeight: "var(--weight-bold)", color: "var(--color-text-primary)" }}>
                Членства
              </h3>

              <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-2)", marginTop: "var(--space-3)" }}>
                {isMembershipsLoading ? <SectionSkeleton rows={3} /> : null}

                {!isMembershipsLoading && memberships.length === 0 ? (
                  <EmptyStateCard
                    body="Назначения появятся после добавления в должность."
                    heading="Нет членств"
                    icon={<Users />}
                  />
                ) : null}

                {!isMembershipsLoading && memberships.length > 0 ? (
                  <ul style={{ listStyle: "none", margin: 0, padding: 0, display: "flex", flexDirection: "column", gap: "var(--space-2)" }}>
                    {memberships.map((item) => {
                      const key = getMembershipKey(item, userId);
                      const isFadingOut = fadingOutKeys.includes(key);

                      return (
                        <ItemRow
                          as="li"
                          key={key}
                          style={{ opacity: isFadingOut ? 0 : 1, transition: "opacity 150ms ease" }}
                        >
                          <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-2)" }}>
                          <div style={{ display: "flex", alignItems: "flex-start", justifyContent: "space-between", gap: "var(--space-3)" }}>
                            <div>
                              <p style={{ margin: 0, fontSize: "var(--text-sm)", fontWeight: "var(--weight-semibold)", color: "var(--color-text-primary)" }}>
                                {item.position_name}
                              </p>
                              <p style={{ margin: "2px 0 0", fontSize: "var(--text-xs)", color: "var(--color-text-muted)" }}>
                                {item.division_name} · {item.role_name}
                              </p>
                            </div>
                            <Button
                              aria-label={`Исключить из ${item.position_name}`}
                              data-testid={`remove-membership-${item.position_id}-${userId}`}
                              onClick={() => setConfirmRemoveId(item.position_id)}
                              size="default"
                              type="button"
                              variant="destructive"
                            >
                              <Trash2 size={14} />
                            </Button>
                          </div>

                          </div>
                        </ItemRow>
                      );
                    })}
                  </ul>
                ) : null}
              </div>
            </div>
          </PageSidebar>
        </PageGrid>
      </PageContent>
      <ConfirmDialog
        body="Вы будете исключены из этой должности. Это действие можно отменить позже."
        confirmLabel="Исключить"
        destructive={true}
        onCancel={() => setConfirmRemoveId(null)}
        onConfirm={() => {
          if (confirmRemoveId === null) {
            return;
          }
          const targetPositionId = confirmRemoveId;
          setConfirmRemoveId(null);
          void handleRemove(targetPositionId, userId);
        }}
        open={confirmRemoveId !== null}
        title="Исключить из членства?"
      />
    </PageRoot>
  );
}
