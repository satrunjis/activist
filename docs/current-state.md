# ТЕКУЩЕЕ_СОСТОЯНИЕ

## Часть 1 — Поверхность Backend API

> Примечание по scope: регистрация routes также монтирует profile endpoints `/api/v1/users/*`. Они упоминаются в Части 2 (поведение route `/profile`) и для полноты перечислены в коротком наборе строк appendix в конце этой части.

### Базовый контракт Backend

- API prefix: `/api/v1`.
- Формат обмена: `application/json; charset=utf-8`.
- Error envelope: `{ "error": { "code": "...", "message": "...", "details": { ... } }, "request_id": "..." }`.
- Pagination для list/search использует `limit` и `offset`; default `limit` — `50`, maximum `limit` — `100`.
- API timestamps возвращаются как ISO 8601 / RFC 3339 UTC.

### Auth

| Метод | Путь | Требуется auth | Что делает (одно предложение) | Поля request body | Форма response (top-level fields) |
|---|---|---|---|---|---|
| POST | `/api/v1/auth/register` | Нет | Создает user account и запускает session через secure cookie. | `login`, `password`, `first_name`, `gradebook_number`, `group_number`, `institute`, `birth_date`, optional `last_name`, `middle_name`, `phone`, `social_links`, `about` | `user`, `session` |
| POST | `/api/v1/auth/login` | Нет | Authenticates existing user и запускает session через secure cookie. | `login`, `password` | `user`, `session` |
| GET | `/api/v1/auth/session` | Да | Возвращает текущие session/user info, CSRF token и resolved permission codes. | None | `user`, `session`, `csrf_token`, `permissions` |
| POST | `/api/v1/auth/logout` | Да | Invalidates текущий session token и очищает session cookie. | None | `204 No Content` |
| POST | `/api/v1/auth/logout-all` | Да | Invalidates все sessions текущего пользователя и очищает session cookie. | None | `204 No Content` |

Заметки по реализации auth/session:

- Authentication использует cookie-based backend sessions.
- Cookie name: `__Host-session`.
- Cookie flags: `Secure`, `HttpOnly`, `SameSite=Lax`, `Path=/`, без `Domain`.
- Cookie хранит только opaque session token; database хранит только его hash.
- Session policy: idle timeout `30m`, absolute timeout `8h`.
- `GET /api/v1/auth/session` возвращает CSRF token; frontend хранит его in memory и отправляет `X-CSRF-Token` в mutating requests.
- Passwords хешируются с Argon2id.

### Divisions

| Метод | Путь | Требуется auth | Что делает (одно предложение) | Поля request body | Форма response (top-level fields) |
|---|---|---|---|---|---|
| GET | `/api/v1/divisions` | Да | Возвращает direct child divisions для root или выбранного parent (`parent_id` query). | None (`parent_id` query) | `items` |
| POST | `/api/v1/divisions` | Да | Создает новое division (root или child) после access checks. | `parent_id` (optional), `short_name`, `full_name`, `description`, `regulation_url`, `media_links` | `id`, `short_name`, `full_name`, `description`, `is_archived`, optional `parent_id`, `regulation_url`, `media_links` |
| PATCH | `/api/v1/divisions/{division_id}` | Да | Обновляет поля division и может переназначить parent с cycle checks. | Optional `parent_id`, `short_name`, `full_name`, `description`, `regulation_url`, `media_links` | `id`, `short_name`, `full_name`, `description`, `is_archived`, optional `parent_id`, `regulation_url`, `media_links` |
| POST | `/api/v1/divisions/{division_id}/archive` | Да | Archives division (cascade path в command transaction). | None | `id`, `short_name`, `full_name`, `description`, `is_archived`, optional `parent_id`, `regulation_url`, `media_links` |
| GET | `/api/v1/divisions/tree` | Да | Возвращает hierarchical division tree с optional depth limit (`depth` query). | None (`depth` query) | `id`, `parent_id`, `short_name`, `is_archived`, `has_children`, `children_count`, `children`, optional `full_name`, `description`, `regulation_url`, `media_links` |

### Roles

| Метод | Путь | Требуется auth | Что делает (одно предложение) | Поля request body | Форма response (top-level fields) |
|---|---|---|---|---|---|
| POST | `/api/v1/roles` | Да | Создает role с набором permission code/scope. | `name`, `permissions[]` (`code`, `scope`; legacy string format also accepted) | `id`, `name`, `kind`, `permissions` |
| GET | `/api/v1/roles` | Да | Возвращает список всех roles. | None | `items` |
| PATCH | `/api/v1/roles/{roleID}` | Да | Редактирует role name и/или permissions. | Optional `name`, optional `permissions[]` | `id`, `name`, `kind`, `permissions` |
| DELETE | `/api/v1/roles/{roleID}` | Да | Удаляет role, если validation это допускает (например, role не используется). | None | `204 No Content` |

Заметки по permission model:

- `Permission` codes определены в backend code, а не как user-managed records.
- `Role.permissions` хранит permission/scope pairs в versioned binary payload.
- Scope modes: `SELF`, `CURRENT_DIVISION` и `CURRENT_AND_DESCENDANTS`.
- `system_admin` — special permission, хранящийся в `Role.permissions`; он не выдается через обычный role-management UI.
- Backend является source of truth для permission checks, scope checks, anti-escalation и business invariants.

### Positions

| Метод | Путь | Требуется auth | Что делает (одно предложение) | Поля request body | Форма response (top-level fields) |
|---|---|---|---|---|---|
| POST | `/api/v1/divisions/{division_id}/positions` | Да | Создает position в division для выбранной role. | `title`, `role_id`, optional `max_count` | `id`, `title`, `role_id`, `division_id`, optional `max_count`, `is_archived` |
| GET | `/api/v1/divisions/{division_id}/positions` | Да | Возвращает список positions для division. | None | `items` |
| GET | `/api/v1/positions/{position_id}/members` | Да | Возвращает members, назначенных на position (permission-gated). | None | `items` |
| POST | `/api/v1/positions/{position_id}/archive` | Да | Archives position и удаляет memberships в transaction flow. | None | `id`, `title`, `role_id`, `division_id`, optional `max_count`, `is_archived` |

### Memberships

| Метод | Путь | Требуется auth | Что делает (одно предложение) | Поля request body | Форма response (top-level fields) |
|---|---|---|---|---|---|
| POST | `/api/v1/memberships` | Да | Назначает user на position после role/scope admissibility checks. | `user_id`, `position_id` | `user_id`, `position_id` |
| DELETE | `/api/v1/positions/{position_id}/members/{user_id}` | Да | Удаляет user из position membership. | None | `204 No Content` |

### Search

| Метод | Путь | Требуется auth | Что делает (одно предложение) | Поля request body | Форма response (top-level fields) |
|---|---|---|---|---|---|
| GET | `/api/v1/search/users` | Да | Выполняет multi-field user search с pagination и optional archived inclusion. | None (`first_name`, `last_name`, `middle_name`, `login`, `group_number`, `institute`, `about`, `position_title`, `role_name`, `include_archived`, `limit`, `offset` query params) | `items`, `total` |

Заметки по реализации search:

- Search ориентирован на substring; full-text search не требуется для первой MVP version.
- PostgreSQL `pg_trgm` — intended baseline для indexed substring search.
- Search rows используют normalized `search_text`; B-tree indexes зарезервированы для exact filters и sorting, а не как замена trigram search.

### Audit

| Метод | Путь | Требуется auth | Что делает (одно предложение) | Поля request body | Форма response (top-level fields) |
|---|---|---|---|---|---|
| GET | `/api/v1/eventlog` | Да | Возвращает paginated audit log с optional event/subject filters. | None (`event_type`, `subject_type`, `subject_id`, `limit`, `offset` query params) | `items`, `total`, `limit`, `offset` |

Заметки по event log:

- Event log writes выполняются backend application/service layer в том же use case, что и business mutation.
- Mutating business operations и их event-log write используют одну database transaction.
- Если event-log write завершается ошибкой, subject mutation откатывается.
- Event payloads включают `payload_version`; MVP payload version — `1`.
- `EventLog` является append-only и не имеет обычного UI/API path для physical deletion или editing.

### Appendix (mounted, используется `/profile`)

| Метод | Путь | Требуется auth | Что делает (одно предложение) | Поля request body | Форма response (top-level fields) |
|---|---|---|---|---|---|
| GET | `/api/v1/users/{user_id}` | Да | Возвращает profile fields, видимые actor. | None | `id`, `first_name`, `last_name`, `middle_name`, optional `login`, `gradebook_number`, `group_number`, `institute`, `birth_date`, `phone`, `social_links`, `about` |
| PATCH | `/api/v1/users/{user_id}` | Да | Обновляет editable profile fields, разрешенные actor. | Optional `first_name`, `last_name`, `middle_name`, `gradebook_number`, `group_number`, `institute`, `birth_date`, `phone`, `social_links`, `about` | Same as GET profile response |
| GET | `/api/v1/users/{user_id}/memberships` | Да | Возвращает user memberships с position/division context. | None | `items` |

Операционные заметки:

- Database migrations используют Goose SQL files в `backend/migrations`.
- Production rollback для данных основан на backup/PITR; `goose down` не считается надежной rollback strategy для complex migrations.
- Incompatible schema changes должны использовать expand/contract approach минимум через два releases.
- `system_seed` — technical first-run seed и должен быть idempotent.
- `demo_seed` предназначен для demonstration/test data и может пересоздавать non-production data, но не определяет production retention policy.
- Session rows могут физически очищаться после revocation или expiry плюс retention window `30` дней.

---

## Часть 2 — Frontend-возможности по Top-Nav Route

### `/` (Org Structure)

- Что пользователь может делать:
  - Выбрать division node в tree.
  - Обновить tree data.
  - Создать division.
  - Редактировать выбранное division.
  - Archive selected division (confirm dialog).
  - Просмотреть positions в выбранном division.
  - Создать position для выбранного division.
  - Выбрать position и загрузить его members.
  - Назначить user на выбранную position.
- Отображаемые данные:
  - Полное organization tree (root children рекурсивно загружаются через children endpoint).
  - Metadata выбранного division: names, description, archive flag, breadcrumb.
  - Position list для выбранного division с role labels и kind badges.
  - Selected position members list (`user_id`, `role_name`, `division_name`).
- Формы и поля:
  - Division form (create/edit): `full_name`, `short_name`, `description`, `parent_id`, `regulation_url`, `media_links[].platform`, `media_links[].value`.
  - Position create form: `title`, `role_id`, `max_count`.
  - Assign member form: `user_id` (position id подразумевается из panel state).
- Поведение success/error:
  - Tree load success: renders tree + side panel.
  - Tree load error: AsyncState error card with retry.
  - Division create/edit success: toast + full tree reload + selection moves to saved node.
  - Division create/edit forbidden: показывается scoped forbidden text, panel mode canceled.
  - Division create/edit validation/server errors: inline field errors или root form error.
  - Archive success: toast + tree reload around parent scope.
  - Archive forbidden: forbidden text в panel.
  - Position list load error: empty-state error card в positions section.
  - Position create success: form closes via mode change and position list reloads.
  - Position create forbidden: form root error + parent forbidden message.
  - Assign member success: list reload; conflict показывает specific duplicate message.
  - Position members load: skeleton while loading, error text on failure, empty text when no members.

### `/roles`

- Что пользователь может делать:
  - Просматривать roles.
  - (Только admin) создавать role.
  - (Только admin) выбирать role и редактировать role.
  - (Только admin) удалять role через confirmation dialog.
- Отображаемые данные:
  - Role list (`name`, `kind`, permissions-derived details через editor form).
  - Sidebar editor surface (create/edit/empty hints).
  - Deletion error alert, если delete fails.
- Формы и поля:
  - Create role form: `name`, `permissions[]` (checkboxes), per-permission `scope` select.
  - Edit role form: те же поля, что в create, предварительно заполнены из selected role.
- Поведение success/error:
  - Roles load success: list visible; soft refresh fades list opacity.
  - Roles load error: AsyncState error card with retry.
  - Create success: toast, form reset, list reload.
  - Edit success: toast, list reload preserving current selection where possible.
  - Delete success: toast, dialog closes, list reloads.
  - Delete error: custom alert in sidebar with mapped human-readable error.
  - Create/edit error: root form error with mapped access/conflict messaging.
  - Non-admin: editor actions hidden, explanatory text shown.

### `/profile`

- Что пользователь может делать:
  - Просматривать детали собственного profile.
  - Открыть edit form и сохранить profile.
  - Просматривать текущие memberships.
  - Удалить membership (inline confirmation per item).
- Отображаемые данные:
  - Profile fields (зависят от server visibility): names, login, study/contact fields, about, social links.
  - Membership list: `position_name`, `division_name`, `role_name`.
- Формы и поля:
  - Profile form (conditionally rendered per visible field): `first_name`, `last_name`, `middle_name`, `gradebook_number`, `group_number`, `institute`, `birth_date`, `phone`, `about`, `social_links[].platform`, `social_links[].value`.
- Поведение success/error:
  - Initial load: centered spinner + loading text.
  - Load error: empty-state error card with retry.
  - Save success: exits edit mode, updates local profile state, toast.
  - Save error: root form error.
  - Membership remove success: row fades out (150ms), removed from list, toast.
  - Membership remove error: per-item inline error text, row remains.
  - Membership list loading: section skeleton; empty memberships uses empty-state card.

### `/search`

- Что пользователь может делать:
  - Ввести multi-field user search criteria.
  - Toggle archived inclusion.
  - Submit search.
  - Retry last search from local or global retry actions.
- Отображаемые данные:
  - Search results count (`total`).
  - Result cards с initials, full name, login fallback, membership chips (up to 3 + overflow badge).
- Формы и поля:
  - Search form: `first_name`, `last_name`, `middle_name`, `login`, `group_number`, `institute`, `about`, `position_title`, `role_name`, `include_archived`.
- Поведение success/error:
  - Success: results list shown, global banner cleared.
  - Empty result: “Ничего не найдено” empty-state card.
  - Error: local error state + global API banner (optional retry button when retryable).
  - Loading: initial skeleton result block; soft-refresh text + dimmed results on repeated search.

### `/audit-log`

- Что пользователь может делать:
  - Просматривать audit records.
  - Фильтровать по event type, subject type, subject id.
  - Apply/reset filters.
  - Paginate forward/back.
  - Retry failed load.
- Отображаемые данные:
  - Audit table rows (`timestamp`, `event_type`, `subject_type`, `subject_id`, `actor_id`).
  - Pagination range summary и total count.
- Формы и поля:
  - Filters form: `event_type`, `subject_type`, `subject_id`.
- Поведение success/error:
  - Success: table rows rendered; hover-highlight per row.
  - Empty: table header + single empty row with icon/message.
  - Error: AsyncState error card with retry.
  - Loading: initial section skeleton; soft-refresh badge and reduced table opacity on refetch.

---

## Часть 3 — Текущий визуальный инвентарь (по Route)

### Shared shell на всех 5 routes

- Dominant colors из shell/nav: `var(--color-canvas)`, `var(--color-nav-bg)`, `var(--color-nav-text)`, `var(--color-nav-text-active)`, `var(--color-nav-active-indicator)`, `var(--color-brand-primary)`, `var(--color-text-inverse)`.
- Typography в shell: classes `text-base`, `text-sm`, custom SVG title `fontSize="11"`.
- Spacing в shell: `var(--space-6)` и fixed px (`12`, `14`, `56` via class utilities and inline).

### `/` (Divisions + positions + memberships panel)

- Dominant colors, найденные в page/children:
  - Hex: `#B0BAC4`, `#CBD5E1`.
  - CSS vars: `var(--color-border)`, `var(--color-brand-primary)`, `var(--color-error)`, `var(--color-node-archived-bg)`, `var(--color-node-archived-text)`, `var(--color-node-bg)`, `var(--color-node-border)`, `var(--color-node-border-hover)`, `var(--color-node-border-selected)`, `var(--color-node-shadow-selected)`, `var(--color-success-subtle)`, `var(--color-surface)`, `var(--color-text-muted)`, `var(--color-text-primary)`, `var(--color-text-secondary)`, `var(--color-warning)`, `var(--color-warning-subtle)`, `var(--color-warning-text)`.
- Найденные font sizes:
  - Hardcoded px: `12px`, `13px`.
  - Typography vars: `var(--text-xs)`, `var(--text-sm)`, `var(--text-md)`, `var(--text-lg)`.
- Spacing pattern:
  - Активное использование `var(--space-1..6)`.
  - Fixed px для micro-adjustments (`1px`, `2px`, `3px`, `4px`, `8px`, `32px`).
  - Mix of `%` dimensions (`100%`) для viewport/tree containers.
- Shared UI components, которые реально используются:
  - From `src/shared/ui/`: `AsyncStateView`, `ConfirmDialog`, `EmptyStateCard`, `SectionSkeleton`, `useToast`, `FullScreenPage`, `ActionPanel`, `EntityCreateSurface`, `EntityEditSurface`, `EntityFormFrame`, `FormActions`.
  - From `src/components/ui/`: `Button`, `Form`, `FormControl`, `FormField`, `FormItem`, `FormLabel`, `FormMessage`, `Input`, `Select`, `SelectTrigger`, `SelectValue`, `SelectContent`, `SelectItem`.
- Inline styles vs className:
  - В основном inline style objects для layout, cards, metadata blocks, nodes, panels.
  - Class usage частичный/utility-based: e.g. `position-row*`, `badge*`, `member-row*`, `muted-text`, `nodrag nopan`, `w-9 px-0`, and button utility class `[border-radius:var(--radius-pill)]`.

### `/roles`

- Dominant colors, найденные в page/children:
  - CSS vars: `var(--color-border)`, `var(--color-border-error)`, `var(--color-brand-primary)`, `var(--color-error-subtle)`, `var(--color-error-text)`, `var(--color-success-subtle)`, `var(--color-surface)`, `var(--color-surface-subtle)`, `var(--color-text-muted)`, `var(--color-text-primary)`, `var(--color-text-secondary)`.
- Найденные font sizes:
  - Typography vars: `var(--text-xs)`, `var(--text-sm)`, `var(--text-base)`.
- Spacing pattern:
  - Primary spacing через `var(--space-1..4,8)`.
  - Micro offsets в px (`1px`, `2px`) для line/label rhythm.
- Shared UI components, которые реально используются:
  - From `src/shared/ui/`: `PageRoot`, `PageHeader`, `PageContent`, `PageGrid`, `PageSidebar`, `AsyncStateView`, `ActionPanel`, `EntityCreateSurface`, `EntityEditSurface`, `ConfirmDialog`, `EmptyStateCard`, `SectionSkeleton`, `useToast`, `EntityFormFrame`, `FormActions`.
  - From `src/components/ui/`: `Button`, `Form`, `FormControl`, `FormField`, `FormItem`, `FormLabel`, `FormMessage`, `Input`.
- Inline styles vs className:
  - Strong inline-style dominance для list items, sidebar alerts, spacing.
  - ClassName mostly delegated to shared primitives (`Button`, form primitives) and default utility strings inside those components.

### `/profile`

- Dominant colors, найденные в page/children:
  - CSS vars: `var(--color-border)`, `var(--color-brand-primary)`, `var(--color-error)`, `var(--color-error-text)`, `var(--color-surface)`, `var(--color-text-link)`, `var(--color-text-muted)`, `var(--color-text-primary)`.
- Найденные font sizes:
  - Typography vars: `var(--text-xs)`, `var(--text-sm)`, `var(--text-base)`, `var(--text-lg)`.
- Spacing pattern:
  - Использует `var(--space-2..6)`.
  - Fixed px values (`1px`, `2px`, `3px`, `10px`, `120px`) and viewport calc (`100vh`, `var(--topbar-height)`).
- Shared UI components, которые реально используются:
  - From `src/shared/ui/`: `PageRoot`, `PageHeader`, `PageContent`, `PageGrid`, `PageSidebar`, `EmptyStateCard`, `SectionSkeleton`, `useToast`, `EntityFormFrame`, `FormActions`.
  - From `src/components/ui/`: `Button`, `Card`, `CardHeader`, `CardContent`, `Form`, `FormControl`, `FormField`, `FormItem`, `FormLabel`, `FormMessage`, `Input`.
- Inline styles vs className:
  - В основном inline styles для profile cards, definition list, membership rows, spinner.
  - Limited className usage (`w-9 px-0`) and component-internal class-based styling.

### `/search`

- Dominant colors, найденные в page/children:
  - CSS vars: `var(--color-border)`, `var(--color-brand-primary)`, `var(--color-surface)`, `var(--color-surface-subtle)`, `var(--color-text-inverse)`, `var(--color-text-muted)`, `var(--color-text-primary)`.
- Найденные font sizes:
  - Typography vars: `var(--text-xs)`, `var(--text-sm)`, `var(--text-xl)`.
- Spacing pattern:
  - Использует `var(--space-2..5)`.
  - Fixed px offsets для micro-motion/chips (`-2px`, `1px`, `2px`, `8px`).
- Shared UI components, которые реально используются:
  - From `src/shared/ui/`: `PageRoot`, `PageContent`, `AsyncStateView`, `EmptyStateCard`, `SectionSkeleton`.
  - From `src/components/ui/`: `Button`, `Input`, `Label`.
- Inline styles vs className:
  - Predominantly inline styles для form grid, result cards/chips, loading/metadata rows.
  - Очень мало прямого className usage за пределами primitive component classes.

### `/audit-log`

- Dominant colors, найденные в page/children:
  - CSS vars: `var(--color-border)`, `var(--color-surface)`, `var(--color-surface-subtle)`, `var(--color-text-muted)`, `var(--color-text-primary)`.
- Найденные font sizes:
  - Typography vars: `var(--text-xs)`, `var(--text-sm)`.
- Spacing pattern:
  - Использует `var(--space-2..4)`.
  - Fixed px lines (`1px`, `2px`) and table width `%`.
- Shared UI components, которые реально используются:
  - From `src/shared/ui/`: `PageRoot`, `PageHeader`, `PageContent`, `PageGrid`, `PageSidebar`, `AsyncStateView`, `ActionPanel`, `SectionSkeleton`.
  - From `src/components/ui/`: `Button`, `Input`, `Label`.
- Inline styles vs className:
  - В основном inline styles для table, pagination, filter panel.
  - ClassName usage минимален и в основном наследуется из shared primitives.

---

## Часть 4 — Инвентарь Shared UI Components

### `src/components/ui/`

| Имя | Принимаемые props | Где сейчас используется | Визуальное описание |
|---|---|---|---|
| `Button` | Native button props + `variant` (`primary/outline/ghost/destructive`), `size` (`sm/default/lg`), `asChild`, `isLoading` | `src/features/audit/AuditLogPage.tsx`, `src/features/auth/RegisterForm.tsx`, `src/features/divisions/DivisionEditForm.tsx`, `src/features/divisions/DivisionExplorer.tsx`, `src/features/positions/PositionListPanel.tsx`, `src/features/profile/ProfileForm.tsx`, `src/features/profile/ProfilePage.tsx`, `src/features/roles/RolesPage.tsx`, `src/features/search/SearchPage.tsx`, `src/shared/ui/forms/FormActions.tsx`, `src/shared/ui/surfaces/EntityCreateSurface.tsx`, `src/shared/ui/surfaces/EntityEditSurface.tsx` | Styled button primitive с variants, sizes и optional spinner label state. |
| `Input` | Native `<input>` props | `src/features/audit/AuditLogPage.tsx`, `src/features/auth/LoginForm.tsx`, `src/features/auth/RegisterForm.tsx`, `src/features/divisions/DivisionEditForm.tsx`, `src/features/positions/AssignMemberForm.tsx`, `src/features/positions/PositionCreateForm.tsx`, `src/features/profile/ProfileForm.tsx`, `src/features/roles/RoleCreateForm.tsx`, `src/features/roles/RoleEditForm.tsx`, `src/features/search/SearchPage.tsx` | Single-line input с border/focus/error/disabled states. |
| `Label` | Radix Label props (+ variant support from CVA wrapper) | `src/features/audit/AuditLogPage.tsx`, `src/features/search/SearchPage.tsx`, plus internal use in `form.tsx` | Form label text primitive. |
| `Card` | Native `<div>` props | `src/features/profile/ProfileForm.tsx` | Surface container с rounded corners и shadow. |
| `CardHeader` | Native `<div>` props | `src/features/profile/ProfileForm.tsx` | Wrapper верхней секции для card heading block. |
| `CardContent` | Native `<div>` props | `src/features/profile/ProfileForm.tsx` | Main content wrapper inside card. |
| `CardTitle` | Native heading props | Not used outside defining file | Title typography helper для card headers. |
| `CardDescription` | Native paragraph props | Not used outside defining file | Secondary descriptive text для cards. |
| `CardFooter` | Native `<div>` props | Not used outside defining file | Footer row wrapper для card actions. |
| `Form` | React Hook Form `FormProvider` props | `src/features/auth/LoginForm.tsx`, `src/features/auth/RegisterForm.tsx`, `src/features/divisions/DivisionEditForm.tsx`, `src/features/positions/AssignMemberForm.tsx`, `src/features/positions/PositionCreateForm.tsx`, `src/features/profile/ProfileForm.tsx`, `src/features/roles/RoleCreateForm.tsx`, `src/features/roles/RoleEditForm.tsx` | Context provider wrapper для form fields. |
| `FormField` | RHF `ControllerProps<TFieldValues, TName>` | Same forms as above | Controller wrapper, связывающий field name с context. |
| `FormItem` | Native `<div>` props | Same forms as above | Field block container (label/control/message grouping). |
| `FormLabel` | `Label` props + `required?: boolean` | Same forms as above | Styled form label с optional required asterisk. |
| `FormControl` | Slot props | Same forms as above | Slot, задающий id/aria-invalid/aria-describedby wiring. |
| `FormDescription` | Native paragraph props | Not used outside defining file | Auxiliary field description text. |
| `FormMessage` | Native paragraph props | Same forms as above | Validation/server message text under control. |
| `RootError` | Native `<div>` props + `message?: ReactNode` | Not used outside defining file | Inline alert box для root-level form errors. |
| `useFormField` | No external props (hook) | Internal to `form.tsx` | Helper hook, exposing generated ids и current field state. |
| `Select` | Radix Select root props | `src/features/positions/PositionCreateForm.tsx` | Select root wrapper (headless behavior). |
| `SelectTrigger` | Radix trigger props | `src/features/positions/PositionCreateForm.tsx` | Styled trigger button для select. |
| `SelectValue` | Radix value props | `src/features/positions/PositionCreateForm.tsx` | Placeholder/current value renderer in trigger. |
| `SelectContent` | Radix content props | `src/features/positions/PositionCreateForm.tsx` | Dropdown container/portal with scroll controls. |
| `SelectItem` | Radix item props | `src/features/positions/PositionCreateForm.tsx` | Select option row with check indicator. |
| `SelectGroup` | Radix group props | Not used outside defining file | Group wrapper для option sections. |
| `SelectLabel` | Radix label props | Not used outside defining file | Label row inside option groups. |
| `SelectSeparator` | Radix separator props | Not used outside defining file | Divider line in dropdown content. |
| `SelectScrollUpButton` | Radix scroll-up props | Internal to `SelectContent` | Scroll control at top of long option list. |
| `SelectScrollDownButton` | Radix scroll-down props | Internal to `SelectContent` | Scroll control at bottom of long option list. |

### `src/shared/ui/`

| Имя | Принимаемые props | Где сейчас используется | Визуальное описание |
|---|---|---|---|
| `PageRoot` | `children`, optional `data-testid`, optional `fullHeight` | `src/features/audit/AuditLogPage.tsx`, `src/features/profile/ProfilePage.tsx`, `src/features/roles/RolesPage.tsx`, `src/features/search/SearchPage.tsx` | Top-level page section с vertical gap/padding. |
| `PageHeader` | `title`, optional `description`, optional `actions`, optional `data-testid` | `src/features/audit/AuditLogPage.tsx`, `src/features/profile/ProfilePage.tsx`, `src/features/roles/RolesPage.tsx` | Standard page title row с optional subtitle/actions. |
| `PageContent` | `children`, optional `maxWidth`, optional `centered` | `src/features/audit/AuditLogPage.tsx`, `src/features/profile/ProfilePage.tsx`, `src/features/roles/RolesPage.tsx`, `src/features/search/SearchPage.tsx` | Width-constraining content wrapper. |
| `PageGrid` | `children`, optional `sidebarWidth`, optional `gap` | `src/features/audit/AuditLogPage.tsx`, `src/features/profile/ProfilePage.tsx`, `src/features/roles/RolesPage.tsx` | Two-column content + sidebar grid. |
| `PageSidebar` | `children`, optional `sticky` | `src/features/audit/AuditLogPage.tsx`, `src/features/profile/ProfilePage.tsx`, `src/features/roles/RolesPage.tsx` | Sidebar column container с vertical stack. |
| `FullScreenPage` | `children`, optional `data-testid` | `src/features/divisions/DivisionExplorer.tsx` | Full-height viewport wrapper under top bar. |
| `EntityFormFrame` | `formId`, `title`, `description`, `rootError`, `actions`, `children`, `onSubmit` | `src/features/auth/LoginForm.tsx`, `src/features/auth/RegisterForm.tsx`, `src/features/divisions/DivisionEditForm.tsx`, `src/features/positions/AssignMemberForm.tsx`, `src/features/positions/PositionCreateForm.tsx`, `src/features/profile/ProfileForm.tsx`, `src/features/roles/RoleCreateForm.tsx`, `src/features/roles/RoleEditForm.tsx` | Reusable form shell с header, content, root error, action slot. |
| `FormActions` | `submitLabel`, optional `submittingLabel`, optional `cancelLabel`, optional `onCancel`, optional `isSubmitting`, optional `disableSubmit` | Same forms as `EntityFormFrame` | Standard submit/cancel action row для forms. |
| `FormRootError` | `message?: string \| null` | Internal in `EntityFormFrame.tsx` | Uniform root-level error alert box. |
| `ActionPanel` | `title`, optional `subtitle`, optional `headerActions`, `children`, optional `width`, optional `stickyHeader` | `src/features/audit/AuditLogPage.tsx`, `src/features/divisions/DivisionExplorer.tsx`, `src/features/roles/RolesPage.tsx` | Card-like side panel с sticky header и scrollable body. |
| `EntityCreateSurface` | `title`, optional `description`, `children`, `onClose` | `src/features/divisions/DivisionExplorer.tsx`, `src/features/roles/RolesPage.tsx` | Bordered tinted container для create form workflows. |
| `EntityEditSurface` | `title`, optional `description`, `children`, `onClose` | `src/features/divisions/DivisionExplorer.tsx`, `src/features/roles/RolesPage.tsx` | Bordered tinted container для edit form workflows. |
| `AsyncStateView<TData>` | `state`, optional `data`, optional `errorMessage`, optional `onRetry`, optional `loadingView`, optional `emptyView`, optional `idleView`, `children(data)` | `src/features/audit/AuditLogPage.tsx`, `src/features/divisions/DivisionExplorer.tsx`, `src/features/positions/PositionMembersList.tsx`, `src/features/roles/RolesPage.tsx`, `src/features/search/SearchPage.tsx` | State switcher, rendering idle/loading/error/empty/success UI branches. |
| `EmptyStateCard` | Optional `icon`, `heading`, `body`, optional `cta {label,onClick}` (legacy `ctaLabel/onCta` supported), optional `className` | `src/features/divisions/DivisionExplorer.tsx`, `src/features/positions/PositionListPanel.tsx`, `src/features/profile/ProfilePage.tsx`, `src/features/roles/RolesPage.tsx`, `src/features/search/SearchResults.tsx`, internal in `AsyncStateView` | Centered empty/error placeholder с icon, copy, optional CTA button. |
| `ConfirmDialog` | `open`, `title`, `body`, `confirmLabel`, `destructive`, `onConfirm`, `onCancel` | `src/app/App.tsx` (mounted noop), `src/features/divisions/DivisionExplorer.tsx`, `src/features/roles/RolesPage.tsx` | Portal modal confirm dialog with backdrop + cancel/confirm actions. |
| `GlobalApiBanner` | `message`, optional `severity`, optional `retryLabel`, optional `onRetry`, optional `onDismiss` | `src/app/App.tsx` | Fixed top banner для global API errors/warnings/info with optional retry/dismiss. |
| `ToastProvider` | `children` | `src/app/App.tsx` | Toast context provider + fixed toast stack renderer. |
| `useToast` | No props (hook returns `{ showToast }`) | `src/features/divisions/DivisionEditForm.tsx`, `src/features/divisions/DivisionExplorer.tsx`, `src/features/profile/ProfilePage.tsx`, `src/features/roles/RoleCreateForm.tsx`, `src/features/roles/RoleEditForm.tsx`, `src/features/roles/RolesPage.tsx` | Hook для triggering transient success/error/info toasts. |
| `SkeletonBlock` | Optional `height`, optional `width`, optional `className`, optional `style` | Internal in `SectionSkeleton.tsx` | Animated shimmer rectangle placeholder. |
| `SectionSkeleton` | Optional `rows`, optional `withHeader`, optional `className` | `src/features/audit/AuditLogPage.tsx`, `src/features/divisions/DivisionExplorer.tsx`, `src/features/positions/PositionListPanel.tsx`, `src/features/positions/PositionMembersList.tsx`, `src/features/profile/ProfilePage.tsx`, `src/features/roles/RolesPage.tsx`, `src/features/search/SearchResults.tsx`, internal in `AsyncStateView.tsx` | Vertical stack of skeleton rows для section/list loading states. |

---

## Часть 5 — Визуальные несогласованности (исчерпывающе)

### Одно и то же пользовательское действие выглядит по-разному на страницах

- Submitting forms визуально несогласован:
  - Standardized `FormActions` footer используется в division/role/profile/position/assign forms.
  - Search submit использует standalone right-aligned pill button (`SearchPage.tsx`) вместо shared `FormActions`.
  - Audit filter submit — full-width primary с отдельным ghost reset (`AuditLogPage.tsx`) вместо shared form footer pattern.
- Deleting/removing entities несогласован:
  - Role deletion использует modal `ConfirmDialog` (`RolesPage.tsx`).
  - Division archive использует modal `ConfirmDialog` (`DivisionExplorer.tsx`).
  - Membership removal в profile использует inline expand/collapse confirmation inside list row (`ProfilePage.tsx`), без modal.
- “Create new entity” affordance отличается по страницам:
  - Roles: “Создать роль” primary pill button in list header (`RolesPage.tsx`).
  - Divisions: “+ Подразделение” primary small button in panel header (`DivisionExplorer.tsx`).
  - Positions: “+ Добавить должность” ghost small button below list (`PositionListPanel.tsx`).
- Retry action style отличается:
  - AsyncState-driven retry использует `EmptyStateCard` CTA button на нескольких страницах.
  - Search также показывает отдельное fixed `GlobalApiBanner` retry action (`SearchPage.tsx` + `App.tsx`).
  - Profile error retry появляется только внутри page-level empty card (`ProfilePage.tsx`).

### Один и тот же тип content отрисовывается с разной markup/spacing

- Collection/list content использует несвязанные patterns:
  - Roles и search results — card-style `<ul><li>` lists с разными paddings и hover effects.
  - Audit log использует `<table>` с compact rows.
  - Profile memberships — bordered cards в vertical list с inline action controls.
  - Positions — clickable `<button>` rows с left accent border when selected.
  - Position members используют custom class-based rows (`member-row`) unlike other list systems.
- Metadata blocks значительно отличаются:
  - Profile использует `<dl>` two-column key/value grid (`ProfilePage.tsx`).
  - Division metadata использует free-form heading/paragraph blocks (`DivisionExplorer.tsx`).
  - Role details embedded in list rows + separate editor panel.

### Стили error messages отличаются

- Inline form root errors стандартизированы во многих forms (`FormRootError`), но non-form errors — нет:
  - Division panel errors отображаются как plain `<p>` lines (muted/error colors) (`DivisionExplorer.tsx`).
  - Role delete error — custom boxed alert (`RolesPage.tsx`).
  - Profile membership remove errors отображаются как per-row plain red text (`ProfilePage.tsx`).
  - Async load errors render `EmptyStateCard` with icon and retry.
  - Search может дополнительно показывать global fixed banner errors (`GlobalApiBanner`) поверх local state.

### Loading states отличаются

- Initial page loading indicators несогласованы:
  - Profile page использует centered spinner + text (`ProfilePage.tsx`).
  - Roles/search/audit/divisions в основном используют `SectionSkeleton` через `AsyncStateView`.
- Soft-refresh indicators vary:
  - Roles/search/audit dim content (`opacity`) и показывают “Обновление…” labels в разных местах.
  - Divisions uses button label change (`Обновление…`) instead of global section badge.
  - Position list uses dimmed list + small muted refresh text.

### Empty states отличаются

- Empty layouts не унифицированы:
  - Search empty использует `EmptyStateCard` (“Ничего не найдено”).
  - Roles empty использует `EmptyStateCard` с create CTA для admins.
  - Divisions root-empty использует centered `EmptyStateCard` overlay on viewport.
  - Position list empty использует plain muted text “Должностей пока нет” (без empty card).
  - Position members empty использует plain muted text “Нет участников”.
  - Audit использует empty table row with icon (not `EmptyStateCard`).
  - Profile memberships empty использует `EmptyStateCard`.

### Кнопки с одинаковым intent отличаются labels/sizes/styles

- Confirm destructive action labels отличаются по context:
  - “Удалить” (role dialog), “Архивировать” (division dialog), “Исключить” (profile membership inline), each with different containment and emphasis.
- Cancel buttons отличаются по location:
  - `FormActions` uses outline cancel.
  - Toggle-based create/assign actions in positions switch button text to “Отмена” using ghost variant.
  - Profile inline remove cancel is outline inside row.
- Primary action button geometry несогласована:
  - Rounded-pill primary in roles/search/division header.
  - Default rounded-card primary in many form submit actions.
  - Full-width primary in audit filter panel.
- Icon-only action buttons differ in color/intention mapping:
  - Profile membership remove icon is red-tinted ghost.
  - Roles delete icon is neutral text secondary ghost.
  - Division archive action is text button with error border in outline style.
