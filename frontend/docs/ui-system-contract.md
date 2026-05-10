# UI System Contract (Code-Grounded Baseline)

Scope date: 2026-04-21.
This contract is derived from the current code in `src/` and is intended as an implementation specification for the next refactor pass.

## 1. Layout System

Current evidence:
- Shell: `src/app/AppShell.tsx` (inline shell styles, `minWidth: 1280`, fullscreen branch)
- Route pages: `DivisionExplorerPage`, `RoleManagementPage`, `ProfilePage`, `SearchPanel`, `AuditLogPage`
- Existing right-column widths: `280` (`AuditLogPage`), `320` (`ProfilePage`), `360` (`RoleManagementPage`), `380` (`DivisionExplorerPage`)

Canonical right-column width: `360px`.
Justification: it is already used by the most editor-heavy non-fullscreen page (`RoleManagementPage`) and is the best fit for the observed spread (`280/320/360/380`) without forcing narrow form layouts.

### `PageRoot`
```ts
export type PageRootProps = {
  children: React.ReactNode;
  "data-testid"?: string;
  fullHeight?: boolean; // false by default; true for viewport-bound pages
};
```
Rules:
- Non-fullscreen pages: vertical stack, gap `var(--space-6)`, outer padding `var(--space-6)`.
- Background comes from shell (`var(--color-canvas)`), page surface composition is delegated.
Maps from existing pages:
- `/roles` (`src/features/roles/RoleManagementPage.tsx`)
- `/profile` (`src/features/profile/ProfilePage.tsx`)
- `/search` (`src/features/search/SearchPanel.tsx`)
- `/audit-log` (`src/features/audit/AuditLogPage.tsx`)
Migration note:
- Replace route-level `<section>` and top-level inline wrappers with `PageRoot`.

### `PageHeader`
```ts
export type PageHeaderProps = {
  title: string;
  description?: string;
  actions?: React.ReactNode;
  "data-testid"?: string;
};
```
Rules:
- Title typography standardizes current repeated `h1` inline style blocks.
- Header spacing is fixed and not page-specific.
Maps from existing pages:
- All current top-level `h1` blocks in roles/search/audit.
Migration note:
- Move per-page heading inline styles into `PageHeader`.

### `PageContent`
```ts
export type PageContentProps = {
  children: React.ReactNode;
  maxWidth?: number | "none"; // default "none"
  centered?: boolean; // default false
};
```
Rules:
- Default fills available width.
- Use `maxWidth={960}` + `centered` for search page parity.
Maps from existing pages:
- `SearchPanel` centered `maxWidth: 960` container.
Migration note:
- Preserve search layout by setting `maxWidth={960}` instead of inline style.

### `PageGrid`
```ts
export type PageGridProps = {
  children: React.ReactNode;
  sidebarWidth?: number; // default 360
  gap?: number; // default var(--space-6)
};
```
Rules:
- Two-column template: `minmax(0,1fr) <sidebarWidth>`.
- Sidebar width defaults to `360`.
Maps from existing pages:
- Roles/profile/audit two-column grids.
Migration note:
- Migrate `280/320` right rails to `360` unless explicitly exempted.

### `PageSidebar`
```ts
export type PageSidebarProps = {
  children: React.ReactNode;
  sticky?: boolean; // default false
};
```
Rules:
- Sidebar owns vertical stacking for right-rail panels.
- Sticky behavior is opt-in.
Maps from existing pages:
- `<aside>` in roles/profile; filters form in audit.
Migration note:
- Replace direct styled `<aside>` with `PageSidebar` + surface wrapper.

### `FullScreenPage`
```ts
export type FullScreenPageProps = {
  children: React.ReactNode;
  "data-testid"?: string;
};
```
Rules:
- Height: `calc(100vh - var(--topbar-height))`.
- `display: flex`, `overflow: hidden`, min-height constrained.
Maps from existing pages:
- `/` route via `DivisionExplorerPage` full-viewport workspace.
Migration note:
- Replace repeated fullscreen wrapper objects in `DivisionExplorerPage` with one `FullScreenPage` primitive.

## 2. Surface System

Current evidence:
- Right workspace panel in `DivisionExplorerPage` is persistent and section-based.
- Role editor is a right-side `aside` in `RoleManagementPage`.
- Position and membership create flows are nested toggles in `PositionListPanel`.

### `ActionPanel`
```ts
export type ActionPanelProps = {
  title: React.ReactNode;
  subtitle?: React.ReactNode;
  headerActions?: React.ReactNode;
  children: React.ReactNode;
  width?: number; // default 360
  stickyHeader?: boolean; // default true
};
```
Composition contract:
- Wraps right-side working area chrome.
- Owns panel shell (border, background, header, scroll container).
- Delegates domain sections and forms to children.
Migrates existing components:
- `DivisionExplorerPage` right panel wrapper.
- `RoleManagementPage` editor `<aside>`.
- `AuditLogPage` filter card can be hosted in `ActionPanel` when used as right rail.

### `EntityCreateSurface`
```ts
export type EntityCreateSurfaceProps = {
  title: string;
  description?: string;
  children: React.ReactNode; // usually EntityFormFrame
  onClose: () => void;
};
```
Composition contract:
- Wraps create flow frame in panel/body context.
- Owns create title and dismissal affordance.
- Delegates validation, submit, and payload formation to form layer.
Migrates existing components:
- `RoleCreateForm` mount block in `RoleManagementPage`.
- Create section in `DivisionExplorerPage` (`DivisionEditForm` mode create).
- `PositionCreateForm` toggle block in `PositionListPanel`.
- `AssignMemberForm` nested create block in `PositionListPanel`.

### `EntityEditSurface`
```ts
export type EntityEditSurfaceProps = {
  title: string;
  description?: string;
  children: React.ReactNode; // usually EntityFormFrame
  onClose: () => void;
};
```
Composition contract:
- Wraps edit flow frame with consistent header/actions placement.
- Owns edit-mode shell and close behavior.
- Delegates entity-specific fields and submit handler.
Migrates existing components:
- `RoleEditForm` mount block in `RoleManagementPage`.
- Edit section in `DivisionExplorerPage` (`DivisionEditForm` mode edit).

## 3. Form System

Current evidence:
- Mixed validation strategies across forms (`zodResolver`, `safeParse`, ad-hoc manual checks).
- Mixed root-error rendering (`RootError` vs `FormMessage` with `errors.root`).

### `EntityFormFrame`
```ts
export type EntityFormFrameProps = {
  formId?: string;
  title?: string;
  description?: string;
  rootError?: string | null;
  actions: React.ReactNode; // FormActions
  children: React.ReactNode; // fields
  onSubmit: (event: React.FormEvent<HTMLFormElement>) => void;
};
```
Lifecycle contract:
- Submit: clear root error, run async submit, map field errors first, map root error second.
- Cancel: delegated to `FormActions` secondary action.
- Error: always rendered through `FormRootError` (never raw `<p>` for root).

### `FormActions`
```ts
export type FormActionsProps = {
  submitLabel: string;
  submittingLabel?: string;
  cancelLabel?: string;
  onCancel?: () => void;
  isSubmitting?: boolean;
  disableSubmit?: boolean;
};
```
Lifecycle contract:
- Primary action is `type="submit"`.
- Secondary action is `type="button"` and optional.
- Button loading/disabled state is bound to `isSubmitting`.

### `FormRootError`
```ts
export type FormRootErrorProps = {
  message?: string | null;
};
```
Lifecycle contract:
- Render only when message exists.
- Visual style is unified with current `RootError` semantics from `src/components/ui/form.tsx`.

Validation contract (resolver-based only):
- Rule 1: Any component using `useForm` in `*Form.tsx` must define `resolver: zodResolver(schema)`.
- Rule 2: `schema.safeParse(...)` is allowed only in non-UI parsing utilities (for unknown external inputs), not inside `*Form.tsx` submit handlers.
- Rule 3: Manual field validation branches inside submit handlers are disallowed when a schema exists.

When to use `zodResolver` vs `safeParse`:
- Use `zodResolver` for user-entered form data managed by `react-hook-form`.
- Use `safeParse` only for non-form boundary parsing (for example, utility adapters handling unknown data shapes).

Enforcement:
- CI check: fail if `safeParse(` appears in `src/features/**/*Form.tsx`.
- CI check: fail if `useForm(` in `src/features/**/*Form.tsx` appears without `resolver:`.
- Code review rule: root errors must use `FormRootError` only.

## 4. Async State System

Current evidence:
- Repeated manual ternary chains for loading/error/empty/success in roles, audit, search, positions, and division bootstrap.

### `AsyncStateView`
```ts
export type AsyncState = "idle" | "loading" | "error" | "empty" | "success";

export type AsyncStateViewProps<TData> = {
  state: AsyncState;
  data?: TData;
  errorMessage?: string | null;
  onRetry?: () => void;
  loadingView?: React.ReactNode;
  emptyView?: React.ReactNode;
  idleView?: React.ReactNode;
  children: (data: TData) => React.ReactNode; // success renderer
};
```

Mapping from current patterns:
- `RoleManagementPage`: `showInitialRolesLoading`/`error`/`roles.length===0`/list -> `loading|error|empty|success`.
- `AuditLogPage`: `showInitialLoading`/`error`/`items.length===0`/table -> `loading|error|empty|success`.
- `SearchResults`: `!hasSearched` maps to `idle`; loading/error/empty/items map to remaining states.
- `PositionMembersList`: existing `idle|loading|loaded|error` maps to canonical with `loaded+0 => empty`, `loaded+>0 => success`.
- `DivisionExplorerPage`: bootstrap loading and top-level load error become `loading|error`; tree/panel workspace is `success`.

## 5. Panel Mode State Machine

Current evidence:
- Reducer has `panelMode` but render logic is driven by local `mode` in `DivisionExplorerPage`.
- Nested create state in `PositionListPanel` is local and independent.

Canonical reducer-owned union:
```ts
export type DivisionPanelMode =
  | { kind: "empty" }
  | { kind: "division.view"; divisionId: string }
  | { kind: "division.edit"; divisionId: string }
  | { kind: "division.create"; parentId: string | null }
  | { kind: "position.create"; divisionId: string }
  | { kind: "membership.create"; divisionId: string; positionId: string };
```

Valid transitions:

| From | Event | To |
|---|---|---|
| `empty` | `select-division(id)` | `division.view(id)` |
| `empty` | `start-division-create(parentId)` | `division.create(parentId)` |
| `division.view` | `start-division-edit(id)` | `division.edit(id)` |
| `division.view` | `start-division-create(parentId)` | `division.create(parentId)` |
| `division.view` | `start-position-create(divisionId)` | `position.create(divisionId)` |
| `division.view` | `start-membership-create(divisionId, positionId)` | `membership.create(divisionId, positionId)` |
| `division.edit` | `cancel` or `save-success(id)` | `division.view(id)` |
| `division.create` | `cancel` | `empty` or `division.view(parentId)` (if parent selected) |
| `division.create` | `save-success(newId)` | `division.view(newId)` |
| `position.create` | `cancel` or `save-success(divisionId)` | `division.view(divisionId)` |
| `membership.create` | `cancel` or `save-success(divisionId)` | `division.view(divisionId)` |
| any non-empty | `clear-selection` | `empty` |

Local state variables deleted on adoption:
- `mode` in `src/features/divisions/DivisionExplorerPage.tsx`
- `showCreateForm` in `src/features/positions/PositionListPanel.tsx`
- `assigningPositionId` in `src/features/positions/PositionListPanel.tsx`

One-way data flow rule:
- Reducer is the single source of truth for panel mode.
- `DivisionExplorerPage` renders from reducer state only.
- Child panels/forms emit events; they do not own mode state.

## 6. API Error System

Current evidence:
- Shared `request()` already throws normalized `ApiRequestError`.
- Feature code duplicates adapters and keeps `instanceof Error` branches.
- `AssignMemberForm` bypasses shared client.

Canonical file: `src/shared/api/errorAdapter.ts`

Canonical shape:
```ts
import type { ApiRequestError } from "./errorPolicy";

export type UiError = {
  message: string;
  code?: string;
  status: number;
  retryable: boolean;
  fieldErrors: Record<string, string>;
  kind: "validation" | "forbidden" | "conflict" | "network" | "unknown";
};

export function adaptApiError(error: unknown, fallback: string): UiError;
export function getFieldErrors(error: UiError): Record<string, string>;
export function getRootMessage(error: UiError): string;
```

Mapping rules from `ApiRequestError` to `UiError`:
- `status===0` -> `kind: "network"`.
- `status===409` -> `kind: "conflict"`.
- `status===400|422` with `fieldErrors` -> `kind: "validation"`.
- `code` starting with `access.` or `status===403` -> `kind: "forbidden"`.
- otherwise -> `kind: "unknown"`.

Rule eliminating `instanceof Error` branches:
- All catch blocks call `adaptApiError(error, fallback)`.
- No feature-level `parseApiError`, `prettyApiError`, or `toApiError` helpers.
- No `error instanceof Error` in feature files for request failures.

Files to migrate and required changes:
- `src/features/positions/AssignMemberForm.tsx`: replace direct `fetch` + local `toUrl` with `request()`, then map errors via `adaptApiError`.
- `src/features/divisions/DivisionExplorerPage.tsx`: remove local `toApiError`; use adapter for all catch paths.
- `src/features/roles/RoleManagementPage.tsx`: remove local `parseApiError`; use adapter and centralized forbidden mapping.
- `src/features/roles/RoleCreateForm.tsx`: remove string JSON parser and `instanceof`; use adapter.
- `src/features/roles/RoleEditForm.tsx`: remove string JSON parser and `instanceof`; use adapter.
- `src/features/profile/ProfilePage.tsx`: remove local `parseApiError`; use adapter.
- `src/features/profile/ProfileForm.tsx`: replace `parseApiErrorMessage` with adapter.
- `src/features/search/SearchPanel.tsx`: remove local `toApiError`; use adapter result for banner/retry.
- `src/features/audit/AuditLogPage.tsx`: remove `prettyApiError`; use adapter.
- `src/features/positions/PositionListPanel.tsx`: remove `prettyApiError`; use adapter.
- `src/features/positions/PositionCreateForm.tsx`: remove local `parseApiError` + `instanceof`; use adapter.
- `src/features/divisions/DivisionEditForm.tsx`: replace local type guard + `instanceof` fallback with adapter + `fieldErrors` extraction.
- `src/features/auth/LoginForm.tsx`: remove `instanceof` branch; use adapter.
- `src/features/auth/RegisterForm.tsx`: remove `instanceof` branch; use adapter.
- `src/features/auth/SessionGate.tsx`: replace string-based unauthorized detection with adapter `status/code` checks.

## 7. Component Naming and File Conventions

Route-level naming rule:
- Every route component must end with `Page`.
- Non-route components must not use `Page` suffix.

Abstraction locations:
- Layout: `src/shared/ui/layout/`
- Surfaces: `src/shared/ui/surfaces/`
- Forms: `src/shared/ui/forms/`
- Async states: `src/shared/ui/states/`
- API error adapter: `src/shared/api/errorAdapter.ts`

Required renames:
- `src/features/search/SearchPanel.tsx` -> `src/features/search/SearchPage.tsx`
- `src/features/roles/RoleManagementPage.tsx` -> `src/features/roles/RolesPage.tsx`
- `src/features/divisions/DivisionTreePage.tsx` -> `src/features/divisions/DivisionsPage.tsx`
- `src/features/divisions/DivisionExplorerPage.tsx` -> `src/features/divisions/DivisionExplorer.tsx`

## 8. Deprecation and Migration Index

| Component | Current File | Target Abstraction | Migration Complexity (Low/Med/High) | Blocks |
|---|---|---|---|---|
| Local API error helpers (`parseApiError` / `prettyApiError` / `toApiError`) | Multiple feature files (see Section 6) | `shared/api/errorAdapter.ts` | Low | None |
| Direct memberships `fetch` flow | `src/features/positions/AssignMemberForm.tsx` | `request()` + `errorAdapter` | Med | Depends on error adapter |
| Root error rendering via `FormMessage` | `DivisionEditForm`, `PositionCreateForm`, `AssignMemberForm` | `FormRootError` | Low | Depends on `shared/ui/forms` |
| Role create/edit near-mirror forms | `src/features/roles/RoleCreateForm.tsx`, `src/features/roles/RoleEditForm.tsx` | `EntityFormFrame` + shared role form model | Med | Depends on form primitives |
| Division form manual validation | `src/features/divisions/DivisionEditForm.tsx` | Resolver-based `EntityFormFrame` | Med | Depends on validation rule adoption |
| Position create `safeParse` submit validation | `src/features/positions/PositionCreateForm.tsx` | Resolver-based `EntityFormFrame` | Med | Depends on validation rule adoption |
| Nested assign/create toggles in positions | `src/features/positions/PositionListPanel.tsx` | Reducer-driven `EntityCreateSurface` | High | Depends on panel state machine |
| Dual mode ownership in division explorer | `src/features/divisions/DivisionExplorerPage.tsx`, `src/features/divisions/explorer/state/explorerReducer.ts` | Canonical `DivisionPanelMode` union | High | Depends on reducer/event redesign |
| Section-by-section right panel composition | `src/features/divisions/DivisionExplorerPage.tsx` | `ActionPanel` + `Entity*Surface` composition | High | Depends on panel mode refactor |
| Inconsistent page wrappers and widths | `RoleManagementPage`, `ProfilePage`, `SearchPanel`, `AuditLogPage` | `PageRoot`/`PageHeader`/`PageContent`/`PageGrid`/`PageSidebar` | Med | Depends on layout primitives |
| Deprecated `EmptyStateCard` API usage | `src/features/positions/PositionListPanel.tsx` | `EmptyStateCard` with `cta` only | Low | None |
| Route component naming drift | Search, roles, divisions files in Section 7 | `*Page` route naming convention | Low | Depends on import updates in `src/app/App.tsx` |

## 9. What Must Not Change

- Do not change route paths or permission contracts in `src/app/routeContracts.ts` (`/`, `/roles`, `/profile`, `/search`, `/audit-log`, `can_view_audit_log`).
- Do not change critical nav test IDs used by app guards (`nav-home`, `nav-audit-log`) in `src/app/App.tsx` and `src/app/routeContracts.ts`.
- Do not change backend endpoint paths, HTTP methods, or payload/response shapes defined in `src/shared/api/types.ts`.
- Do not change CSRF behavior in `src/shared/api/client.ts` (token storage and mutating-method header injection).
- Do not change division tree visualization mechanics in `src/features/divisions/tree/*` (layout/viewport behavior).
- Do not change domain permission semantics in `src/features/roles/permissionOptions.ts`.
- Do not change user-visible business flows: role delete confirmation, division archive confirmation, membership removal confirmation, auth tab split login/register.
