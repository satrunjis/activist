# Activist Frontend

## 1. Tech Stack

| Area | Technology | Version |
|---|---|---|
| Runtime | Node.js (required by tooling) | Not pinned in repo |
| Language | TypeScript | 5.9.3 |
| Framework | React | 19.2.5 |
| Renderer | React DOM | 19.2.5 |
| Build tool | Vite | 7.3.2 |
| Vite React plugin | `@vitejs/plugin-react` | 5.2.0 |
| Testing | Vitest | 4.1.4 |
| Coverage | `@vitest/coverage-v8` | 4.1.4 |
| DOM test env | jsdom | 29.0.2 |
| Testing utils | Testing Library (`react`, `user-event`, `jest-dom`) | 16.3.2 / 14.6.1 / 6.9.1 |
| Linting | ESLint | 10.2.1 |
| TS lint integration | `typescript-eslint` | 8.58.2 |
| React hooks lint rules | `eslint-plugin-react-hooks` | 7.1.1 |
| Form state | `react-hook-form` | 7.72.1 |
| Validation | Zod | 4.3.6 |
| RHF-Zod bridge | `@hookform/resolvers` | 5.2.2 |
| Graph/tree visualization | `@xyflow/react` (React Flow) | 12.10.2 |
| Graph layout engine | `elkjs` | 0.11.1 |
| Icon set | `lucide-react` | 1.8.0 |
| UI primitives | Radix UI (`label`, `select`, `slot`) | 2.1.8 / 2.2.6 / 1.2.4 |
| Variant/class utilities | `class-variance-authority`, `clsx`, `tailwind-merge` | 0.7.1 / 2.1.1 / 3.5.0 |
| CSS pipeline | PostCSS + Autoprefixer | 8.5.9 + 10.4.27 |
| UI kit basis | `shadcn/ui` structure (`components.json`) + custom CSS design system (`src/index.css`) | Config present |

Build/test scripts from `package.json`:
- `npm run dev`
- `npm run build`
- `npm run preview`
- `npm run lint`
- `npm run test:unit`
- `npm run test:unit:run`

Environment variables:
- No frontend API base URL variable is required; browser requests use relative `/api/v1` paths.

---

## 2. Project Structure

```text
activist-frontend/ — React + Vite frontend application root.
├─ .env — Local runtime env values (current API base URL).
├─ .env.example — Template env file for required frontend variables.
├─ .gitignore — Git ignore rules.
├─ components.json — shadcn/ui generator config and aliases.
├─ docs/ — Frontend design, UI audit, and navigation analysis documents.
├─ eslint.config.js — ESLint + TypeScript + React hooks rule configuration.
├─ index.html — Vite HTML entry, root mount node, Google Fonts preload/import.
├─ package-lock.json — Dependency lockfile.
├─ package.json — Scripts, dependency manifests, project metadata.
├─ postcss.config.js — PostCSS plugin setup (Autoprefixer).
├─ README.md — Existing minimal project readme.
├─ tsconfig.json — TypeScript compiler configuration for `src`.
├─ vite.config.ts — Vite config using React plugin.
├─ vitest.config.ts — Vitest config (jsdom, setup file, coverage rules).
├─ dist/ — Build output directory.
├─ node_modules/ — Installed dependencies.
├─ .tmp/ — Local temporary artifacts directory.
└─ src/ — Application source code.
   ├─ index.css — Global design system variables and all app/page/component styling.
   ├─ main.tsx — React bootstrap (`createRoot` + `App` mount).
   ├─ vite-env.d.ts — Vite TypeScript ambient types.
   ├─ app/ — App shell, route contracts, and top-level composition.
   │  ├─ App.tsx — Root route switch, global API banner state, shell composition.
   │  ├─ AppNavigationRail.tsx — Legacy navigation rail component (not wired into shell).
   │  ├─ AppShell.tsx — Sidebar + main content layout and mobile-nav orchestration.
   │  ├─ AppSidebar.tsx — Current primary navigation sidebar with user badge and icons.
   │  ├─ AppTopBar.tsx — Legacy top bar component (not wired into shell).
   │  └─ routeContracts.ts — Route definitions, test-id contracts, permission filters.
   ├─ components/ — Reusable low-level UI primitives.
   │  └─ ui/
   │     ├─ button.tsx — Variant-based button primitive.
   │     ├─ card.tsx — Card container and card subcomponents.
   │     ├─ form.tsx — React Hook Form wrappers (`FormField`, `FormMessage`, etc.).
   │     ├─ input.tsx — Text input primitive.
   │     ├─ label.tsx — Label primitive wrapping Radix Label.
   │     └─ select.tsx — Select primitive wrappers around Radix Select.
   ├─ features/ — Domain feature modules.
   │  ├─ audit/
   │  │  └─ AuditLogPage.tsx — Audit log table, filters, and pagination.
   │  ├─ auth/
   │  │  ├─ LoginForm.behavior.test.tsx — Auth form validation behavior tests.
   │  │  ├─ LoginForm.tsx — Login form with RHF + Zod and POST login call.
   │  │  ├─ RegisterForm.tsx — Registration form with dynamic social links and POST register.
   │  │  ├─ SessionGate.behavior.test.tsx — Session bootstrap/auth-state transition tests.
   │  │  └─ SessionGate.tsx — Session bootstrap gate, auth wall, logout handling.
   │  ├─ divisions/
   │  │  ├─ DivisionEditForm.tsx — Create/edit division form with media links.
   │  │  ├─ DivisionExplorerPage.tsx — Main org-tree explorer page with right-side detail panel.
   │  │  ├─ DivisionTreePage.behavior.test.tsx — Division explorer ACL and eager-load behavior tests.
   │  │  ├─ DivisionTreePage.tsx — Wrapper page exporting explorer and child-fetch helper.
   │  │  ├─ explorer/ — Explorer support modules.
   │  │  │  ├─ divisionChildrenApi.ts — API helper for loading division children.
   │  │  │  ├─ layout/
   │  │  │  │  ├─ elkLayoutAdapter.test.ts — ELK layout adapter tests.
   │  │  │  │  └─ elkLayoutAdapter.ts — ELK-based tree node layout adapter.
   │  │  │  ├─ lod/
   │  │  │  │  ├─ zoomLod.test.ts — Zoom-to-LOD mapping tests.
   │  │  │  │  └─ zoomLod.ts — Semantic zoom level mapping logic.
   │  │  │  └─ state/
   │  │  │     ├─ explorerReducer.test.ts — Explorer reducer state-transition tests.
   │  │  │     └─ explorerReducer.ts — Explorer UI state reducer and action definitions.
   │  │  └─ tree/
   │  │     ├─ divisionTreeLayout.ts — Tree layout wrapper delegating to ELK adapter.
   │  │     ├─ DivisionTreeViewport.behavior.test.tsx — Tree viewport interaction/LOD tests.
   │  │     ├─ DivisionTreeViewport.tsx — React Flow viewport renderer for division graph.
   │  │     └─ treeTypes.ts — Tree viewport data and LOD type contracts.
   │  ├─ positions/
   │  │  ├─ AssignMemberForm.tsx — Manual fetch form to assign user to position.
   │  │  ├─ PositionCreateForm.tsx — Position creation form (role selector + constraints).
   │  │  ├─ PositionListPanel.tsx — Positions list for selected division + assign/create toggles.
   │  │  ├─ positionMembersApi.ts — API helper for position member retrieval.
   │  │  └─ PositionMembersList.tsx — Position member list rendering with loading/error states.
   │  ├─ profile/
   │  │  ├─ ProfileForm.tsx — Editable profile form with field visibility derived from API payload.
   │  │  └─ ProfilePage.tsx — Profile summary + memberships + membership removal flow.
   │  ├─ roles/
   │  │  ├─ permissionOptions.ts — Role permission code list and scope options.
   │  │  ├─ RoleCreateForm.tsx — Role creation form and permission scope assignment.
   │  │  ├─ RoleEditForm.tsx — Role editing form and permission scope assignment.
   │  │  └─ RoleManagementPage.tsx — Role list page with create/edit/delete actions.
   │  └─ search/
   │     ├─ SearchPanel.tsx — Search filter form + debounced requests + global error integration.
   │     └─ SearchResults.tsx — Search result state rendering (loading/error/empty/list).
   ├─ lib/
   │  └─ utils.ts — Shared `cn()` class merge helper (`clsx` + `tailwind-merge`).
   ├─ shared/ — Cross-feature shared contracts/helpers/UI.
   │  ├─ api/
   │  │  ├─ __tests__/
   │  │  │  └─ errorPolicy.test.ts — Error normalization and retry policy tests.
   │  │  ├─ client.ts — Shared HTTP client (`fetch`, credentials, CSRF, error normalization).
   │  │  ├─ errorPolicy.ts — API error normalization, retry eligibility, field error extraction.
   │  │  └─ types.ts — API request/response TypeScript contracts.
   │  ├─ forms/
   │  │  └─ schemas.ts — Shared Zod schemas for auth/profile form payloads.
   │  └─ ui/
   │     └─ feedback/
   │        ├─ EmptyStateCard.tsx — Empty-state card with optional CTA.
   │        ├─ GlobalApiBanner.tsx — Top-level API error banner.
   │        ├─ index.ts — Feedback component barrel exports.
   │        ├─ SectionSkeleton.tsx — Section-level loading skeleton composition.
   │        └─ SkeletonBlock.tsx — Primitive shimmering skeleton block.
   └─ test/
      └─ setup.ts — Vitest setup and browser API polyfills for jsdom.
```

---

## 3. Component Inventory

### App Shell and Navigation Components

| Component | File | Responsibility |
|---|---|---|
| `App` | `src/app/App.tsx` | Root component; maps URL path to page component; manages global API banner; renders authenticated shell through `SessionGate`. |
| `AppShell` | `src/app/AppShell.tsx` | Main shell layout; controls mobile sidebar open/close lifecycle; renders route content with/without full-screen padding. |
| `AppSidebar` | `src/app/AppSidebar.tsx` | Primary navigation sidebar with permission-filtered routes and user identity footer. |
| `AppNavigationRail` | `src/app/AppNavigationRail.tsx` | Legacy nav rail/backdrop navigation implementation. |
| `AppTopBar` | `src/app/AppTopBar.tsx` | Legacy top bar with navigation toggle and route title. |

### Auth Components

| Component | File | Responsibility |
|---|---|---|
| `SessionGate` | `src/features/auth/SessionGate.tsx` | Bootstraps session (`/auth/session`), sets CSRF token, handles anonymous/authenticated branches, logout flow. |
| `LoginForm` | `src/features/auth/LoginForm.tsx` | Login form validation and login submit call. |
| `RegisterForm` | `src/features/auth/RegisterForm.tsx` | Registration form validation, optional social links array handling, register submit call. |

### Page Components

| Component | File | Responsibility |
|---|---|---|
| `DivisionTreePage` | `src/features/divisions/DivisionTreePage.tsx` | Route page wrapper for organization explorer. |
| `DivisionExplorerPage` | `src/features/divisions/DivisionExplorerPage.tsx` | Full organization explorer: eager BFS tree load, tree viewport, division actions, position/member side panel. |
| `RoleManagementPage` | `src/features/roles/RoleManagementPage.tsx` | Role list and admin-only create/edit/delete workflows. |
| `ProfilePage` | `src/features/profile/ProfilePage.tsx` | Profile details, profile editing container, memberships list, membership removal controls. |
| `SearchPanel` | `src/features/search/SearchPanel.tsx` | Search filters, debounced submit, global banner integration. |
| `AuditLogPage` | `src/features/audit/AuditLogPage.tsx` | Audit log table, filter form, and offset pagination. |

### Feature/Subfeature Components

| Component | File | Responsibility |
|---|---|---|
| `DivisionEditForm` | `src/features/divisions/DivisionEditForm.tsx` | Shared create/edit form for division metadata and media links. |
| `DivisionTreeViewport` | `src/features/divisions/tree/DivisionTreeViewport.tsx` | React Flow viewport for division nodes/edges, zoom LOD switching, minimap/controls. |
| `MicroNode` | `src/features/divisions/tree/DivisionTreeViewport.tsx` | Low-zoom minimal node renderer (initials pill). |
| `CompactNode` | `src/features/divisions/tree/DivisionTreeViewport.tsx` | Compact node renderer for medium-low zoom. |
| `StandardNode` | `src/features/divisions/tree/DivisionTreeViewport.tsx` | Default node renderer with title/archived badge/children count. |
| `DetailNode` | `src/features/divisions/tree/DivisionTreeViewport.tsx` | High-zoom detailed node renderer (full name + description preview). |
| `DivisionFlowNode` | `src/features/divisions/tree/DivisionTreeViewport.tsx` | Node dispatcher selecting renderer by LOD and wiring React Flow handles. |
| `PositionListPanel` | `src/features/positions/PositionListPanel.tsx` | Loads/display division positions and roles; toggles create/assign forms. |
| `PositionCreateForm` | `src/features/positions/PositionCreateForm.tsx` | Position creation form with role loading and validation. |
| `AssignMemberForm` | `src/features/positions/AssignMemberForm.tsx` | Assigns user to selected position with explicit CSRF header handling. |
| `PositionMembersList` | `src/features/positions/PositionMembersList.tsx` | Renders selected position members across idle/loading/error/empty states. |
| `ProfileForm` | `src/features/profile/ProfileForm.tsx` | Edits backend-exposed profile fields only; dynamic visible-field payload construction. |
| `RoleCreateForm` | `src/features/roles/RoleCreateForm.tsx` | Creates role and selected permission scopes. |
| `RoleEditForm` | `src/features/roles/RoleEditForm.tsx` | Edits role name and permission scopes. |
| `SearchResults` | `src/features/search/SearchResults.tsx` | Search output renderer for loading/error/empty/result list states. |

### Shared Feedback Components

| Component | File | Responsibility |
|---|---|---|
| `SkeletonBlock` | `src/shared/ui/feedback/SkeletonBlock.tsx` | Base shimmering skeleton placeholder block. |
| `SectionSkeleton` | `src/shared/ui/feedback/SectionSkeleton.tsx` | Prebuilt skeleton layout for sections/lists. |
| `EmptyStateCard` | `src/shared/ui/feedback/EmptyStateCard.tsx` | Standard empty state UI with optional CTA. |
| `GlobalApiBanner` | `src/shared/ui/feedback/GlobalApiBanner.tsx` | Dismissible top-level API error/warning/info banner with optional retry button. |

### UI Primitive Components

| Component | File | Responsibility |
|---|---|---|
| `Button` | `src/components/ui/button.tsx` | Variant/size button primitive. |
| `Card` | `src/components/ui/card.tsx` | Card container primitive. |
| `CardHeader` | `src/components/ui/card.tsx` | Card header wrapper. |
| `CardTitle` | `src/components/ui/card.tsx` | Card title heading. |
| `CardDescription` | `src/components/ui/card.tsx` | Card subtitle/description text. |
| `CardContent` | `src/components/ui/card.tsx` | Card body wrapper. |
| `CardFooter` | `src/components/ui/card.tsx` | Card footer wrapper. |
| `Form` | `src/components/ui/form.tsx` | Form provider wrapper (`react-hook-form` `FormProvider`). |
| `FormField` | `src/components/ui/form.tsx` | RHF `Controller` wrapper with field context. |
| `FormItem` | `src/components/ui/form.tsx` | Form item wrapper with generated ids. |
| `FormLabel` | `src/components/ui/form.tsx` | Label bound to form field state/ids. |
| `FormControl` | `src/components/ui/form.tsx` | Control slot with aria wiring and invalid state. |
| `FormDescription` | `src/components/ui/form.tsx` | Optional field description text. |
| `FormMessage` | `src/components/ui/form.tsx` | Error/message renderer for field/root messages. |
| `Input` | `src/components/ui/input.tsx` | Input primitive. |
| `Label` | `src/components/ui/label.tsx` | Label primitive. |
| `Select` | `src/components/ui/select.tsx` | Select root primitive. |
| `SelectGroup` | `src/components/ui/select.tsx` | Select group primitive. |
| `SelectValue` | `src/components/ui/select.tsx` | Select selected-value primitive. |
| `SelectTrigger` | `src/components/ui/select.tsx` | Select trigger/button primitive. |
| `SelectContent` | `src/components/ui/select.tsx` | Select popover content primitive. |
| `SelectLabel` | `src/components/ui/select.tsx` | Select label primitive. |
| `SelectItem` | `src/components/ui/select.tsx` | Select option primitive. |
| `SelectSeparator` | `src/components/ui/select.tsx` | Select separator primitive. |
| `SelectScrollUpButton` | `src/components/ui/select.tsx` | Select scroll up control. |
| `SelectScrollDownButton` | `src/components/ui/select.tsx` | Select scroll down control. |

---

## 4. API Contract

### Auth Mechanism and Request Behavior

- Base URL source: relative same-origin `/api/v1` paths via `src/shared/api/client.ts`.
- Shared client: `src/shared/api/client.ts` (`request()`).
- Transport: `fetch`.
- Credentials: always `credentials: "include"` (cookie-based session).
- Content type: `application/json` for requests with body.
- CSRF:
  - `csrf_token` is received from `GET /api/v1/auth/session`.
  - Stored in-memory via `setCsrfToken`.
  - Shared client auto-adds `X-CSRF-Token` for `POST/PUT/PATCH/DELETE` when token is set.
  - `AssignMemberForm` does manual `fetch` but still adds `X-CSRF-Token` from `getCsrfToken()`.
- Authorization header: no `Authorization: Bearer ...` usage in codebase.

### Endpoints Consumed

| Endpoint | Method | Consumed In | Request Shape | Response Shape |
|---|---|---|---|---|
| `/api/v1/auth/session` | `GET` | `SessionGate` | No body | `AuthSessionResponse`: `{ user: UserProfile, session: { expires_at, idle_expires_at }, csrf_token: string, permissions?: string[] }` |
| `/api/v1/auth/login` | `POST` | `LoginForm` | `LoginRequest`: `{ login: string, password: string }` | Response body not used by caller (session bootstrap is performed after call). |
| `/api/v1/auth/register` | `POST` | `RegisterForm` | `RegisterRequest`: `{ login, password, first_name, gradebook_number, group_number, institute, birth_date, last_name?, middle_name?, phone?, social_links?: SocialLink[], about? }` | Response body not used by caller (session bootstrap is performed after call). |
| `/api/v1/auth/logout` | `POST` | `SessionGate` | No body | Response body not used; frontend clears local CSRF/session state regardless. |
| `/api/v1/eventlog?limit=1&offset=0` | `GET` | `SessionGate` | Query: `limit=1`, `offset=0` | `EventLogListResponse` used as permission probe fallback (if request succeeds, frontend adds `can_view_audit_log`). |
| `/api/v1/eventlog` | `GET` | `AuditLogPage` | Query: `event_type?`, `subject_type?`, `subject_id?`, `limit`, `offset` | `EventLogListResponse`: `{ items: EventLogItem[], total, limit, offset }` |
| `/api/v1/search/users` | `GET` | `SearchPanel` | Query from `UserSearchRequest`: `first_name?`, `last_name?`, `middle_name?`, `login?`, `group_number?`, `institute?`, `about?`, `position_title?`, `role_name?`, `include_archived?`, `limit`, `offset` | `UserSearchResponse`: `{ items: SearchUserItem[], total }` |
| `/api/v1/divisions?parent_id=root` | `GET` | `fetchDivisionChildren` via `DivisionExplorerPage` | Query: `parent_id=root` for top-level divisions | `DivisionChildrenResponse`: `{ items: DivisionChildItem[] }` |
| `/api/v1/divisions?parent_id={divisionId}` | `GET` | `fetchDivisionChildren` via `DivisionExplorerPage` | Query: `parent_id` for child division loading | `DivisionChildrenResponse`: `{ items: DivisionChildItem[] }` |
| `/api/v1/divisions` | `POST` | `DivisionExplorerPage` (`handleCreate`) | `DivisionCreateRequest`: `{ parent_id?, short_name, full_name, description, regulation_url, media_links: DivisionMediaLink[] }` | `DivisionEntity`: `{ id, parent_id?, short_name, full_name, description, regulation_url?, media_links?, is_archived }` |
| `/api/v1/divisions/{divisionId}` | `PATCH` | `DivisionExplorerPage` (`handleEdit`) | `DivisionPatchRequest`: `{ parent_id?, short_name?, full_name?, description?, regulation_url?, media_links? }` | `DivisionEntity` |
| `/api/v1/divisions/{divisionId}/archive` | `POST` | `DivisionExplorerPage` | No body | Response body not used by caller. |
| `/api/v1/divisions/{divisionId}/positions` | `GET` | `PositionListPanel` | Path param `divisionId` | `{ items: PositionItem[] }` |
| `/api/v1/divisions/{divisionId}/positions` | `POST` | `PositionCreateForm` | `CreatePositionRequest`: `{ title: string, role_id: string, max_count?: number }` | Response body not used by caller. |
| `/api/v1/roles` | `GET` | `RoleManagementPage`, `PositionListPanel`, `PositionCreateForm` | No body | `{ items: RoleItem[] }` |
| `/api/v1/roles` | `POST` | `RoleCreateForm` | `CreateRoleRequest`: `{ name: string, permissions: RolePermission[] }` | Response body not used by caller. |
| `/api/v1/roles/{roleId}` | `PATCH` | `RoleEditForm` | `EditRoleRequest`: `{ name?: string, permissions?: RolePermission[] }` | Response body not used by caller. |
| `/api/v1/roles/{roleId}` | `DELETE` | `RoleManagementPage` | No body | Response body not used by caller. |
| `/api/v1/memberships` | `POST` | `AssignMemberForm` (manual `fetch`) | `AssignMemberRequest`: `{ user_id: string, position_id: string }` | Response body not consumed; caller checks status codes (`409`, `403`, non-OK). |
| `/api/v1/positions/{positionId}/members` | `GET` | `getPositionMembers` / `DivisionExplorerPage` | Path param `positionId` | `PositionMembersResponse`: `{ items: PositionMemberItem[] }` |
| `/api/v1/positions/{positionId}/members/{userId}` | `DELETE` | `ProfilePage` | Path params `positionId`, `userId` | Response body not used by caller. |
| `/api/v1/users/{userId}` | `GET` | `ProfilePage` | Path param `userId` | `UserProfile` |
| `/api/v1/users/{userId}` | `PATCH` | `ProfileForm` | Subset of profile fields (`ProfilePatch`-compatible shape, filtered by visible backend-provided fields) | `UserProfile` |
| `/api/v1/users/{userId}/memberships` | `GET` | `ProfilePage` | Path param `userId` | Local `MembershipsResponse`: `{ items: Array<{ position_id, position_name, division_id, division_name, role_name }> }` |

### Shared API Type Shapes

Core API type contracts are declared in `src/shared/api/types.ts`:
- Auth/session: `AuthSessionResponse`, `AuthSession`
- Error envelope: `ApiErrorEnvelope`, `ApiErrorDetail`
- User/profile: `UserProfile`, `ProfilePatch`, `RegisterRequest`, `LoginRequest`, `SocialLink`
- Divisions: `DivisionTreeNode`, `DivisionChildItem`, `DivisionChildrenResponse`, `DivisionCreateRequest`, `DivisionPatchRequest`, `DivisionEntity`, `DivisionMediaLink`
- Roles: `RoleItem`, `RoleWithPermissions`, `CreateRoleRequest`, `EditRoleRequest`, `RolePermission`, `PermissionScope`, `RoleKind`
- Positions/memberships: `PositionItem`, `PositionMemberItem`, `PositionMembersResponse`, `CreatePositionRequest`, `AssignMemberRequest`
- Search: `UserSearchRequest`, `SearchUserItem`, `UserSearchResponse`
- Audit: `EventLogItem`, `EventLogListResponse`

---

## 5. State Management

- Global state library: none (`redux`, `zustand`, etc. are not used).
- Routing state: `App` tracks current path in local `useState` and synchronizes via `popstate`.
- Session/auth state: `SessionGate` manages bootstrap/loading/anonymous/authenticated states and CSRF token side effects.
- Feature state pattern:
  - `useState` for local UI/data/loading/error in feature pages (`SearchPanel`, `AuditLogPage`, `ProfilePage`, `RoleManagementPage`, `PositionListPanel`).
  - `useReducer` only in division explorer (`explorerReducer`) for selected node/position/panel/member-load state.
  - `useMemo` for derived maps, filtered payloads, and computed tree viewport nodes.
  - `useEffect` for data fetching and lifecycle actions.
- Form state:
  - `react-hook-form` for form state and validation lifecycle.
  - `zod` schemas for auth/profile/role/position validation.
  - `useFieldArray` for dynamic `social_links` and `media_links`.
- Data flow:
  - API requests happen inside feature/page components or thin API helpers.
  - Parent-to-child props drive UI state.
  - Child callbacks bubble events upward (e.g., position selection, create/edit success).
  - No centralized cache/query client (`react-query`/SWR not used).

---

## 6. Known Issues

- Dead/unused shell components:
  - `src/app/AppNavigationRail.tsx` is exported but not imported by active shell.
  - `src/app/AppTopBar.tsx` is exported but not imported by active shell.
  - `src/index.css` forcibly hides legacy classes `.app-top-bar`, `.app-navigation-rail`, `.app-shell-backdrop`.
- Lint failure in current state:
  - `src/features/search/SearchPanel.tsx` has `react-hooks/exhaustive-deps` warning (`useEffect` missing `executeSearch` dependency), and `npm run lint` fails because `--max-warnings 0`.
- API client usage inconsistency:
  - `src/features/positions/AssignMemberForm.tsx` uses raw `fetch` with local `toUrl` and local error parsing instead of shared `request()` from `src/shared/api/client.ts`.
- Error parsing pattern inconsistency:
  - Multiple files implement local parse helpers (`RoleCreateForm`, `RoleEditForm`, `RoleManagementPage`, `PositionCreateForm`, `ProfilePage`, `AuditLogPage`) instead of reusing one shared parser.
- Role form error handling mismatch:
  - `RoleCreateForm` and `RoleEditForm` branch on `error instanceof Error`; shared API client throws plain normalized error objects, so code-paths expecting `Error` instance are not taken.
- Mixed styling patterns:
  - UI primitives use utility-class strings (shadcn-style), while feature pages use custom CSS classes and extensive inline styles.
- Mixed UI copy language:
  - UI text contains both Russian and English strings across components.
- Unused route metadata:
  - `routeContracts.ts` defines `regions` for routes, but no renderer logic consumes route regions beyond `fullScreen`.
- TODO/FIXME markers:
  - No `TODO`, `FIXME`, `HACK`, or `XXX` comments were found in `src`.

---

## 7. Running Locally

```bash
npm install
npm run dev
```

Local backend expectation:
- Frontend API requests use same-origin `/api/v1/...` paths. In production, nginx proxies `/api/` to the backend.

---

## 8. Environment Variables

No frontend API base URL variable is required. API calls use relative `/api/v1/...` paths and rely on the serving proxy to route `/api/` to the backend.

---

## 9. Docker Build

Build image:

```bash
docker build -t frontend:latest .
```

Run container:

```bash
docker run --rm -p 8081:80 frontend:latest
```

### Production image (tag used by deploy compose)

The browser uses relative `/api/v1/...` paths, and the outer nginx handles routing to the backend.

```bash
docker build -t activist-frontend:prod .
docker save activist-frontend:prod | gzip > activist-frontend-prod.tar.gz
```

Transfer to server and load:

```bash
scp -P 64971 activist-frontend-prod.tar.gz deploy@77.221.139.39:/opt/activist/
ssh -p 64971 deploy@77.221.139.39 'docker load < /opt/activist/activist-frontend-prod.tar.gz'
```

> **Note on native bindings:** `@tailwindcss/oxide` is a native Rust module. If `package-lock.json` was generated on Windows, `npm ci` inside a Linux container will fail with "Cannot find native binding". The Dockerfile uses `npm install` (no lockfile) to avoid this. Do not change it back to `npm ci`.

---

## 10. Production nginx

`nginx.conf` in repo root is used by the production container and does the following:
- Serves static Vite build output from `/usr/share/nginx/html` on port `80`.
- Enables SPA fallback routing: `try_files $uri $uri/ /index.html`.
- Proxies `/api/` to `http://backend:8080`.
- Enables gzip for `text/html`, `text/css`, `application/javascript`, and `application/json`.
- Applies long-lived cache headers (`max-age=31536000`) to `.js`, `.css`, `.woff2`, `.png`, `.svg`, and `.ico`.
- Does not configure SSL/TLS; TLS termination is expected at an outer reverse proxy.

Proxy note:
- Frontend requests stay on the current origin and rely on `/api/` proxy routing.
