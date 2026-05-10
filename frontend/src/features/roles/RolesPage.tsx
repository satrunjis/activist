import { MousePointerClick, ShieldOff, Trash2 } from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { ReactNode } from "react";

import { Button } from "../../components/ui/button";
import { request } from "../../shared/api/client";
import { adaptApiError } from "../../shared/api/errorAdapter";
import type { RoleListResponse, RoleItem, RoleWithPermissions } from "../../shared/api/types";
import { ConfirmDialog, EmptyStateCard, InlineError, SectionSkeleton, SoftRefreshWrapper, useToast } from "../../shared/ui/feedback";
import { ItemRow, PageContent, PageGrid, PageHeader, PageRoot, PageSidebar } from "../../shared/ui/layout";
import { AsyncStateView, type AsyncState } from "../../shared/ui/states/AsyncStateView";
import { ActionPanel, EntityCreateSurface, EntityEditSurface } from "../../shared/ui/surfaces";
import { PERMISSION_OPTIONS, SCOPE_OPTIONS, type PermissionOption } from "./permissionOptions";
import { RoleCreateForm } from "./RoleCreateForm";
import { RoleEditForm } from "./RoleEditForm";

function mapRoleError(error: { code?: string; kind: string; message: string }): string {
  if (error.code === "validation.role_in_use") {
    return "Роль назначена одной или нескольким должностям.";
  }
  if (error.code?.startsWith("access.scope")) {
    return "Операция недоступна в текущем контексте.";
  }
  if (error.kind === "forbidden" || error.code?.startsWith("access.")) {
    return "Операция запрещена.";
  }
  return error.message;
}

function toRoleWithPermissions(item: RoleItem): RoleWithPermissions {
  return {
    id: item.id,
    name: item.name,
    permissions: item.permissions
  };
}

type PermissionDetailsItem = {
  code: string;
  label: string;
  scopeLabel: string;
};

type PermissionDetailsGroup = {
  category: string;
  items: PermissionDetailsItem[];
};

type RolesPageProps = {
  permissions: readonly string[];
};

type RoleEditorView = "idle" | "details" | "create" | "edit";

const DEFAULT_ROLE_LIMIT = 100;
const SYSTEM_ADMIN_PERMISSION = "system_admin";
const PERMISSION_OPTION_BY_CODE = new Map<string, PermissionOption>(
  PERMISSION_OPTIONS.map((option) => [option.value, option])
);
const SCOPE_LABEL_BY_VALUE = new Map(SCOPE_OPTIONS.map((scope) => [scope.value, scope.label]));

function getPermissionDetailsGroups(role: RoleWithPermissions): PermissionDetailsGroup[] {
  const permissionsByCode = new Map((role.permissions ?? []).map((permission) => [permission.code, permission]));
  const grouped = new Map<string, PermissionDetailsItem[]>();

  for (const option of PERMISSION_OPTIONS) {
    const permission = permissionsByCode.get(option.value);
    if (!permission) {
      continue;
    }
    const items = grouped.get(option.category) ?? [];
    items.push({
      code: permission.code,
      label: option.label,
      scopeLabel: SCOPE_LABEL_BY_VALUE.get(permission.scope) ?? permission.scope
    });
    grouped.set(option.category, items);
    permissionsByCode.delete(option.value);
  }

  for (const permission of permissionsByCode.values()) {
    const option = PERMISSION_OPTION_BY_CODE.get(permission.code);
    const category = option?.category ?? "Другое";
    const items = grouped.get(category) ?? [];
    items.push({
      code: permission.code,
      label: option?.label ?? permission.code,
      scopeLabel: SCOPE_LABEL_BY_VALUE.get(permission.scope) ?? permission.scope
    });
    grouped.set(category, items);
  }

  return Array.from(grouped.entries()).map(([category, items]) => ({ category, items }));
}

export function RolesPage({ permissions }: RolesPageProps) {
  const { showToast } = useToast();
  const canManageRoles = permissions.includes(SYSTEM_ADMIN_PERMISSION);
  const [roles, setRoles] = useState<RoleWithPermissions[]>([]);
  const [total, setTotal] = useState(0);
  const [limit] = useState(DEFAULT_ROLE_LIMIT);
  const [offset, setOffset] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeView, setActiveView] = useState<RoleEditorView>("idle");
  const [selectedRoleId, setSelectedRoleId] = useState<string | null>(null);
  const [dialogRole, setDialogRole] = useState<RoleWithPermissions | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const selectedRole = useMemo(
    () => roles.find((role) => role.id === selectedRoleId) ?? null,
    [roles, selectedRoleId]
  );
  const selectedRolePermissionGroups = useMemo(
    () => selectedRole ? getPermissionDetailsGroups(selectedRole) : [],
    [selectedRole]
  );

  const selectedRoleIdRef = useRef(selectedRoleId);
  const roleIdsBeforeCreateRef = useRef<Set<string>>(new Set());
  useEffect(() => { selectedRoleIdRef.current = selectedRoleId; }, [selectedRoleId]);

  const loadRoles = useCallback(async (preferredSelectionId?: string | null, requestedOffset = offset): Promise<RoleWithPermissions[]> => {
    setIsLoading(true);
    setError(null);
    try {
      const params = new URLSearchParams({
        limit: String(limit),
        offset: String(requestedOffset)
      });
      const response = await request<RoleListResponse>(`/api/v1/roles?${params.toString()}`);
      const nextRoles = response.items.map(toRoleWithPermissions);
      setRoles(nextRoles);
      setTotal(typeof response.total === "number" ? response.total : nextRoles.length);
      setOffset(typeof response.offset === "number" ? response.offset : requestedOffset);
      if (preferredSelectionId !== undefined) {
        setSelectedRoleId(preferredSelectionId && nextRoles.some((item) => item.id === preferredSelectionId) ? preferredSelectionId : null);
      } else if (selectedRoleIdRef.current) {
        setSelectedRoleId(nextRoles.some((item) => item.id === selectedRoleIdRef.current) ? selectedRoleIdRef.current : null);
      }
      return nextRoles;
    } catch (loadError) {
      setError(adaptApiError(loadError, "Не удалось загрузить роли.").message);
      return [];
    } finally {
      setIsLoading(false);
    }
  }, [limit, offset]);

  useEffect(() => {
    void loadRoles();
  }, [loadRoles]);

  useEffect(() => {
    if (!canManageRoles) {
      if (activeView === "create" || activeView === "edit") {
        setActiveView("idle");
      }
      setDialogRole(null);
      setDeleteError(null);
    }
  }, [activeView, canManageRoles]);

  useEffect(() => {
    if ((activeView === "details" || activeView === "edit") && !selectedRole) {
      setActiveView("idle");
    }
  }, [activeView, selectedRole]);

  async function handleDeleteConfirmed() {
    if (!canManageRoles || !dialogRole || isDeleting) {
      return;
    }

    setDeleteError(null);
    setIsDeleting(true);
    try {
      await request(`/api/v1/roles/${encodeURIComponent(dialogRole.id)}`, { method: "DELETE" });
      showToast("success", "Роль удалена");
      setDialogRole(null);
      setSelectedRoleId(null);
      setActiveView("idle");
      const nextOffset = roles.length === 1 ? Math.max(0, offset - limit) : offset;
      await loadRoles(null, nextOffset);
    } catch (deleteErrorValue) {
      const apiError = adaptApiError(deleteErrorValue, "Не удалось удалить роль.");
      setDeleteError(mapRoleError(apiError));
    } finally {
      setIsDeleting(false);
    }
  }

  async function handleCreateSuccess() {
    setDeleteError(null);
    const previousIds = roleIdsBeforeCreateRef.current;
    const nextRoles = await loadRoles(undefined, 0);
    const createdRole = nextRoles.find((role) => !previousIds.has(role.id));
    const nextSelection = createdRole?.id ?? nextRoles[0]?.id ?? null;
    setSelectedRoleId(nextSelection);
    setActiveView(nextSelection ? "details" : "idle");
  }

  async function handleEditSuccess() {
    setDeleteError(null);
    const preferredSelectionId = selectedRoleIdRef.current;
    const nextRoles = await loadRoles(preferredSelectionId, offset);
    const nextSelection = preferredSelectionId && nextRoles.some((item) => item.id === preferredSelectionId)
      ? preferredSelectionId
      : null;
    setSelectedRoleId(nextSelection);
    setActiveView(nextSelection ? "details" : "idle");
  }
  const showInitialRolesLoading = isLoading && roles.length === 0 && !error;
  const showSoftRefreshing = isLoading && roles.length > 0;
  const rolesState: AsyncState = showInitialRolesLoading
    ? "loading"
    : error
      ? "error"
      : roles.length === 0
        ? "empty"
        : "success";
  const canPrev = offset > 0;
  const canNext = offset + limit < total;
  const rangeStart = total === 0 ? 0 : offset + 1;
  const rangeEnd = total === 0 ? 0 : Math.min(offset + limit, total);

  function goToPage(nextOffset: number) {
    setSelectedRoleId(null);
    setActiveView("idle");
    setDeleteError(null);
    setOffset(Math.max(0, nextOffset));
  }
  const rolePanelHeader = useMemo<{ title: ReactNode; subtitle?: ReactNode }>(() => {
    switch (activeView) {
      case "idle":
        return { title: "Роли" };
      case "details":
        return { title: selectedRole?.name ?? "Роли" };
      case "create":
        return { title: "Новая роль" };
      case "edit":
        return {
          title: "Редактирование",
          subtitle: selectedRole?.name
        };
      default:
        return { title: "Роли" };
    }
  }, [activeView, selectedRole?.name]);

  return (
    <PageRoot data-testid="role-management-page">
      <PageHeader title="Управление ролями" />

      <PageContent>
        <PageGrid>
          <section
            style={{
              background: "var(--color-surface)",
              borderRadius: "var(--radius-card)",
              boxShadow: "var(--shadow-sm)",
              padding: "var(--space-4)"
            }}
          >
            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: "var(--space-3)" }}>
              <h2 style={{ margin: 0, fontSize: "var(--text-base)", fontWeight: "var(--weight-bold)" }}>Роли</h2>
              <div style={{ display: "flex", alignItems: "center", gap: "var(--space-2)" }}>
                <Button
                  data-testid="refresh-roles"
                  onClick={() => void loadRoles()}
                  size="default"
                  type="button"
                  variant="outline"
                >
                  Обновить
                </Button>
                {canManageRoles ? (
                  <Button
                    data-testid="create-role-btn"
                    onClick={() => {
                      roleIdsBeforeCreateRef.current = new Set(roles.map((role) => role.id));
                      setSelectedRoleId(null);
                      setActiveView("create");
                      setDeleteError(null);
                    }}
                    size="default"
                    type="button"
                    variant="primary"
                  >
                    Создать роль
                  </Button>
                ) : null}
              </div>
            </div>

            <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-2)", marginTop: "var(--space-3)" }}>
              <AsyncStateView
                data={roles}
                emptyView={(
                  <EmptyStateCard
                    body="Создайте первую роль для назначения должностям."
                    cta={canManageRoles ? {
                      label: "Создать роль",
                      onClick: () => {
                        roleIdsBeforeCreateRef.current = new Set(roles.map((role) => role.id));
                        setSelectedRoleId(null);
                        setActiveView("create");
                      }
                    } : undefined}
                    heading="Роли не созданы"
                    icon={<ShieldOff />}
                  />
                )}
                errorMessage={error}
                loadingView={<SectionSkeleton rows={5} />}
                onRetry={() => void loadRoles()}
                state={rolesState}
              >
                {(roleItems) => (
                  <SoftRefreshWrapper isRefreshing={showSoftRefreshing}>
                    <ul
                      data-testid="role-list"
                      style={{
                        listStyle: "none",
                        margin: 0,
                        padding: 0,
                        display: "grid",
                        gap: "var(--space-2)"
                      }}
                    >
                      {roleItems.map((role) => {
                        const selected = activeView !== "create" && selectedRoleId === role.id;
                        return (
                          <ItemRow
                            as="li"
                            data-testid={`role-item-${role.id}`}
                            onClick={() => {
                              setSelectedRoleId(role.id);
                              setActiveView("details");
                              setDeleteError(null);
                            }}
                            selected={selected}
                            key={role.id}
                          >
                            <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between" }}>
                              <div style={{ minWidth: 0 }}>
                                <p
                                  data-testid={`role-name-${role.id}`}
                                  style={{
                                    margin: 0,
                                    fontSize: "var(--text-sm)",
                                    fontWeight: "var(--weight-semibold)",
                                    color: "var(--color-text-primary)"
                                  }}
                                >
                                  {role.name}
                                </p>
                              </div>

                              {canManageRoles ? (
                                <div style={{ display: "flex", alignItems: "center", gap: "var(--space-1)" }}>
                                  <Button
                                    aria-label={`Удалить роль ${role.name}`}
                                    data-testid={`delete-role-${role.id}`}
                                    onClick={(event) => {
                                      event.stopPropagation();
                                      setDialogRole(role);
                                    }}
                                    size="default"
                                    type="button"
                                    variant="destructive"
                                  >
                                    <Trash2 size={14} />
                                  </Button>
                                </div>
                              ) : null}
                            </div>
                          </ItemRow>
                        );
                      })}
                    </ul>
                  </SoftRefreshWrapper>
                )}
              </AsyncStateView>
            </div>

            <div style={{ marginTop: "var(--space-3)", display: "flex", alignItems: "center", gap: "var(--space-4)" }}>
              <Button
                data-testid="roles-pagination-prev"
                disabled={!canPrev || isLoading}
                onClick={() => goToPage(offset - limit)}
                type="button"
                variant="outline"
              >
                ← Назад
              </Button>
              <span
                data-testid="roles-pagination-state"
                style={{ flex: 1, textAlign: "center", fontSize: "var(--text-sm)", color: "var(--color-text-muted)" }}
              >
                Записи {rangeStart}–{rangeEnd} из {total}
              </span>
              <Button
                data-testid="roles-pagination-next"
                disabled={!canNext || isLoading}
                onClick={() => goToPage(offset + limit)}
                type="button"
                variant="outline"
              >
                Вперёд →
              </Button>
            </div>
          </section>

          <PageSidebar sticky>
            <ActionPanel subtitle={rolePanelHeader.subtitle} title={rolePanelHeader.title}>
              <InlineError message={deleteError} />

              {canManageRoles && activeView === "create" ? (
                <div style={{ marginTop: "var(--space-3)" }}>
                  <EntityCreateSurface onClose={() => setActiveView("idle")} showTitle={false}>
                    <RoleCreateForm
                      onCancel={() => setActiveView("idle")}
                      onSuccess={() => {
                        void handleCreateSuccess();
                      }}
                    />
                  </EntityCreateSurface>
                </div>
              ) : null}

              {canManageRoles && activeView === "edit" && selectedRole ? (
                <div style={{ marginTop: "var(--space-3)" }}>
                  <EntityEditSurface onClose={() => setActiveView("details")} showTitle={false}>
                    <RoleEditForm
                      onCancel={() => setActiveView("details")}
                      onSuccess={() => {
                        void handleEditSuccess();
                      }}
                      role={selectedRole}
                    />
                  </EntityEditSurface>
                </div>
              ) : null}

              {activeView === "details" && selectedRole ? (
                <div style={{ marginTop: "var(--space-3)", display: "grid", gap: "var(--space-3)" }}>
                  <div>
                    <p style={{ margin: 0, fontSize: "var(--text-xs)", color: "var(--color-text-muted)", textTransform: "uppercase", letterSpacing: "0.06em" }}>
                      Разрешения
                    </p>
                    {selectedRolePermissionGroups.length > 0 ? (
                      <div style={{ display: "grid", gap: "var(--space-1)", marginTop: "var(--space-2)" }}>
                        {selectedRolePermissionGroups.map((group, groupIndex) => (
                          <div key={group.category}>
                            <p
                              style={{
                                margin: groupIndex === 0 ? "0 0 var(--space-1)" : "var(--space-3) 0 var(--space-1)",
                                fontSize: "var(--text-xs)",
                                fontWeight: "var(--weight-semibold)",
                                color: "var(--color-text-muted)",
                                textTransform: "uppercase",
                                letterSpacing: "0.06em"
                              }}
                            >
                              {group.category}
                            </p>
                            <div style={{ display: "grid", gap: "var(--space-1)" }}>
                              {group.items.map((permission) => (
                                <div
                                  key={permission.code}
                                  style={{
                                    alignItems: "center",
                                    display: "flex",
                                    gap: "var(--space-3)",
                                    justifyContent: "space-between",
                                    padding: "var(--space-1) 0"
                                  }}
                                >
                                  <span style={{ color: "var(--color-text-primary)", flex: 1, fontSize: "var(--text-sm)" }}>
                                    {permission.label}
                                  </span>
                                  <span
                                    style={{
                                      border: "1px solid var(--color-border)",
                                      borderRadius: "var(--radius-input)",
                                      color: "var(--color-text-muted)",
                                      flexShrink: 0,
                                      fontSize: "var(--text-xs)",
                                      minHeight: 28,
                                      padding: "5px var(--space-2)"
                                    }}
                                  >
                                    {permission.scopeLabel}
                                  </span>
                                </div>
                              ))}
                            </div>
                          </div>
                        ))}
                      </div>
                    ) : (
                      <p style={{ margin: "4px 0 0", fontSize: "var(--text-sm)", color: "var(--color-text-muted)" }}>—</p>
                    )}
                  </div>

                  {canManageRoles ? (
                    <div style={{ display: "flex", alignItems: "center", gap: "var(--space-2)" }}>
                      <Button
                        onClick={() => setActiveView("edit")}
                        size="default"
                        type="button"
                        variant="outline"
                      >
                        Редактировать
                      </Button>
                      <Button
                        onClick={() => setDialogRole(selectedRole)}
                        size="default"
                        type="button"
                        variant="destructive"
                      >
                        Удалить
                      </Button>
                    </div>
                  ) : null}
                </div>
              ) : null}

              {activeView === "idle" ? (
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
                    {canManageRoles
                      ? "Выберите роль для редактирования или создайте новую."
                      : "Управление ролями доступно только администраторам."}
                  </p>
                </div>
              ) : null}
            </ActionPanel>
          </PageSidebar>
        </PageGrid>
      </PageContent>

      <ConfirmDialog
        body={
          dialogRole
            ? `Роль «${dialogRole.name}» будет удалена безвозвратно. Убедитесь, что она не назначена ни одной должности.`
            : ""
        }
        confirmLabel={isDeleting ? "Удаление…" : "Удалить"}
        destructive
        onCancel={() => {
          if (!isDeleting) {
            setDialogRole(null);
          }
        }}
        onConfirm={() => {
          void handleDeleteConfirmed();
        }}
        open={Boolean(dialogRole)}
        title="Удалить роль?"
      />
    </PageRoot>
  );
}
