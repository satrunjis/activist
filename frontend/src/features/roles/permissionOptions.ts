import type { PermissionScope } from "../../shared/api/types";

export const PERMISSION_OPTIONS = [
  { value: "can_add_member", label: "Добавление участников", defaultScope: "current_division", category: "Участники" },
  { value: "can_remove_member", label: "Исключение участников", defaultScope: "current_division", category: "Участники" },
  { value: "can_assign_position", label: "Назначение на должности", defaultScope: "current_division", category: "Должности" },
  { value: "can_manage_positions", label: "Управление должностями", defaultScope: "current_division", category: "Должности" },
  { value: "can_edit_division", label: "Редактирование подразделения", defaultScope: "current_division", category: "Подразделения" },
  { value: "can_create_subdivision", label: "Создание подподразделений", defaultScope: "current_and_descendants", category: "Подразделения" },
  { value: "can_archive_division", label: "Архивирование подразделений", defaultScope: "current_and_descendants", category: "Подразделения" },
  { value: "can_view_contacts", label: "Просмотр контактов", defaultScope: "current_division", category: "Профили" },
  { value: "can_edit_self_profile", label: "Редактирование своего профиля", defaultScope: "self", category: "Профили" },
  { value: "can_manage_roles", label: "Управление ролями", defaultScope: "current_and_descendants", category: "Роли" }
] as const;

export type PermissionOption = (typeof PERMISSION_OPTIONS)[number];

// Keep in sync with ScopeMode ordering in activist-backend/src/domain/role/permission.go.
export const SCOPE_OPTIONS: Array<{ value: PermissionScope; label: string }> = [
  { value: "self", label: "Только я" },
  { value: "current_division", label: "Текущее" },
  { value: "current_and_descendants", label: "С потомками" }
];
