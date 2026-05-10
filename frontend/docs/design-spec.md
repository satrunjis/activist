# DESIGN_SPEC.md — Activist Base Frontend Redesign

---

## Section 1 — Design Tokens

```css
:root {
  /* ─── COLOR: Brand ─────────────────────────────────── */
  --color-brand-primary: #46AE5B;          /* primary CTA bg */
  --color-brand-primary-hover: #00B052;    /* primary CTA hover bg */
  --color-brand-dark: #004751;             /* dark teal, nav bar bg */
  --color-brand-dark-hover: #005F6E;       /* nav item hover */
  --color-brand-dark-active: #003840;      /* nav item active */

  /* ─── COLOR: Surface ───────────────────────────────── */
  --color-canvas: #F0F2F4;                 /* page background */
  --color-surface: #FFFFFF;               /* card / panel surface */
  --color-surface-subtle: #F8F8F8;        /* input bg, row hover */
  --color-surface-overlay: rgba(0,0,0,0.40); /* modal backdrop */

  /* ─── COLOR: Navigation ────────────────────────────── */
  --color-nav-bg: #004751;                /* top bar background */
  --color-nav-text: rgba(255,255,255,0.65); /* inactive label */
  --color-nav-text-active: #FFFFFF;       /* active label */
  --color-nav-item-hover-bg: #005F6E;     /* item hover bg */
  --color-nav-item-active-bg: #003840;    /* active item bg */
  --color-nav-active-indicator: #46AE5B;  /* active underline bar */

  /* ─── COLOR: Text ──────────────────────────────────── */
  --color-text-primary: #333333;          /* headings, body */
  --color-text-secondary: #555F6D;        /* secondary/meta */
  --color-text-muted: #8A94A0;            /* tertiary, hints */
  --color-text-disabled: #C4CAD2;         /* disabled */
  --color-text-inverse: #FFFFFF;          /* on dark surfaces */
  --color-text-link: #46AE5B;             /* links */

  /* ─── COLOR: Border ────────────────────────────────── */
  --color-border: #E2E6EA;                /* default border */
  --color-border-strong: #C5CBD2;         /* hover border */
  --color-border-focus: #46AE5B;          /* focus ring */
  --color-border-error: #D94F4F;          /* error border */

  /* ─── COLOR: Semantic ──────────────────────────────── */
  --color-success: #46AE5B;
  --color-success-subtle: #E8F7EB;
  --color-success-text: #1E6E30;
  --color-error: #D94F4F;
  --color-error-subtle: #FDECEA;
  --color-error-text: #8B1A1A;
  --color-warning: #E08A00;
  --color-warning-subtle: #FFF3D6;
  --color-warning-text: #7A4A00;
  --color-info: #0070C0;
  --color-info-subtle: #E0F0FF;
  --color-info-text: #003F6E;

  /* ─── COLOR: Tree Nodes ────────────────────────────── */
  --color-node-bg: #FFFFFF;
  --color-node-border: #D0D7DE;
  --color-node-border-hover: #46AE5B;
  --color-node-border-selected: #004751;
  --color-node-shadow-selected: 0 0 0 3px rgba(70,174,91,0.30);
  --color-node-archived-bg: #F5F5F5;
  --color-node-archived-text: #9AA3AD;

  /* ─── COLOR: Edge ──────────────────────────────────── */
  --color-edge: #B0BAC4;

  /* ─── TYPOGRAPHY ───────────────────────────────────── */
  --font-body: 'Raleway', Arial, sans-serif;
  --font-mono: 'IBM Plex Mono', 'Consolas', monospace;

  --text-xs:   11px;   /* meta, badges */
  --text-sm:   13px;   /* secondary body, labels */
  --text-base: 15px;   /* default body */
  --text-md:   17px;   /* section titles */
  --text-lg:   20px;   /* card titles */
  --text-xl:   24px;   /* page headings */
  --text-2xl:  30px;   /* auth heading */

  --weight-regular:  400;
  --weight-medium:   500;
  --weight-semibold: 600;
  --weight-bold:     700;

  --leading-tight:  1.2;
  --leading-normal: 1.5;

  /* ─── SPACING (4px base) ───────────────────────────── */
  --space-1:  4px;
  --space-2:  8px;
  --space-3:  12px;
  --space-4:  16px;
  --space-5:  20px;
  --space-6:  24px;
  --space-8:  32px;
  --space-10: 40px;
  --space-12: 48px;
  --space-16: 64px;

  /* ─── RADIUS ───────────────────────────────────────── */
  --radius-input: 5px;    /* inputs, chips, badges */
  --radius-card:  10px;   /* cards, panels, modals */
  --radius-pill:  30px;   /* pill buttons */
  --radius-icon:  8px;    /* icon containers */
  --radius-full:  999px;  /* avatars */

  /* ─── SHADOW ───────────────────────────────────────── */
  --shadow-xs:         0 1px 2px rgba(0,0,0,0.06);
  --shadow-sm:         0 1px 4px rgba(0,0,0,0.09);
  --shadow-md:         0 4px 12px rgba(0,0,0,0.11);
  --shadow-lg:         0 8px 24px rgba(0,0,0,0.14);
  --shadow-card-hover: 0 0 10px rgba(0,0,0,0.20);
  --shadow-panel:      -6px 0 24px rgba(0,0,0,0.10);
  --shadow-modal:      0 16px 48px rgba(0,0,0,0.22);
  --shadow-toast:      0 4px 16px rgba(0,0,0,0.16);

  /* ─── TRANSITION ───────────────────────────────────── */
  --transition-fast: 120ms ease;
  --transition-base: 200ms ease;
  --transition-slow: 300ms ease;

  /* ─── DIMENSIONS ───────────────────────────────────── */
  --topbar-height:      56px;
  --detail-panel-width: 380px;
  --modal-width:        440px;
  --toast-width:        320px;
}
```

---

## Section 2 — Shell Layout

### Top Navigation Bar

- Height: `var(--topbar-height)` = 56px
- Background: `var(--color-nav-bg)`
- Box-shadow: `var(--shadow-sm)`
- Position: fixed, top 0, left 0, right 0, z-index 100
- Layout: `display: flex; align-items: center; padding: 0 var(--space-6);`

### Left Region — Brand

- Logo mark: 32×32px, border-radius `var(--radius-icon)`, background `var(--color-brand-primary)`, white "A" glyph SVG centered
- Wordmark: "Activist Base", `var(--font-body)`, `var(--text-base)`, `var(--weight-bold)`, `var(--color-text-inverse)`
- Gap between mark and wordmark: `var(--space-2)`
- Right margin to nav center: `var(--space-8)`

### Center Region — Navigation Items

| Route | Icon | Label |
|---|---|---|
| `/` | `Network` | Оргструктура |
| `/roles` | `Shield` | Роли |
| `/search` | `Search` | Поиск |
| `/profile` | `User` | Профиль |
| `/audit-log` | `ClipboardList` | Журнал |

- `/audit-log` rendered only when user has `can_view_audit_log` permission
- Center region: `flex: 1; display: flex; justify-content: center; gap: var(--space-1)`

**Nav item anatomy:**

- Element: `<a>` or `<button>` with class `.nav-item`
- Children: icon 16×16px + label `<span>`
- Padding: `8px 14px`
- Border-radius: `var(--radius-card)`
- Font: `var(--text-sm)` `var(--weight-medium)` `var(--font-body)`
- Color default: `var(--color-nav-text)`
- Background default: transparent
- Display: `flex; align-items: center; gap: var(--space-2)`
- **Hover:** background `var(--color-nav-item-hover-bg)`, color `var(--color-text-inverse)`, transition `var(--transition-fast)`
- **Active:** background `var(--color-nav-item-active-bg)`, color `var(--color-text-inverse)`; `position: relative` with `::after`: `height: 2px; background: var(--color-nav-active-indicator); position: absolute; bottom: -1px; left: 14px; right: 14px; border-radius: 2px`
- Icon opacity: 0.65 default, 1.0 on hover and active

### Right Region — User Identity

- Layout: `display: flex; align-items: center; gap: var(--space-3); margin-left: var(--space-8)`
- Avatar: 32×32px, border-radius `var(--radius-full)`, background `var(--color-brand-primary)`, color `var(--color-text-inverse)`, font `var(--text-xs)` `var(--weight-bold)`, content = first 2 chars of username uppercase
- Username: `var(--text-sm)` `var(--weight-medium)` `var(--color-nav-text-active)`
- Logout button: icon-only (`LogOut`, 16px), padding `var(--space-2)`, border-radius `var(--radius-card)`, color `var(--color-nav-text)`, hover color `var(--color-text-inverse)` hover background `var(--color-nav-item-hover-bg)`, `aria-label="Выйти"`

### Content Area

- `padding-top: var(--topbar-height)` on `.app-shell__body`
- Background: `var(--color-canvas)`
- Full-screen routes (`/`): content height `calc(100vh - var(--topbar-height))`, overflow hidden, zero additional padding
- Standard routes: content padding `var(--space-6)`

**No sidebar. No mobile breakpoints. Min-width: 1280px.**

---

## Section 3 — Auth Screen

### Initial Mode

- Login form shown first. Returning users are the majority; registration is a one-time action.

### Page Background and Card

- Full-page background: `var(--color-canvas)`
- Centered vertically and horizontally: `display: flex; align-items: center; justify-content: center; min-height: 100vh`
- Above card: logo mark 40×40px + wordmark, centered, `margin-bottom: var(--space-6)`
- Auth card: width 440px, background `var(--color-surface)`, border-radius `var(--radius-card)`, box-shadow `var(--shadow-lg)`, padding `var(--space-8) var(--space-8) var(--space-6)`

### Toggle Mechanic

- Two tab labels inside card, below card heading, above form fields: "Вход" | "Регистрация"
- Tab row: `display: flex; gap: 0; border-bottom: 1px solid var(--color-border); margin-bottom: var(--space-5)`
- **Active tab:** `var(--weight-bold)`, `var(--color-brand-primary)`, `border-bottom: 2px solid var(--color-brand-primary)`, margin-bottom -1px, padding `var(--space-2) var(--space-4)`
- **Inactive tab:** `var(--weight-medium)`, `var(--color-text-muted)`, no underline, padding `var(--space-2) var(--space-4)`, cursor pointer
- Click switches the form instantly; only one form is mounted at a time

### Bootstrap Loading State

- Full-page, centered: spinner 24px (`border: 2px solid var(--color-border); border-top-color: var(--color-brand-primary); border-radius: var(--radius-full); animation: spin 0.8s linear infinite`) + text "Проверка сессии…" below, `var(--text-sm)` `var(--color-text-muted)`

### Login Form

Fields (stacked, gap `var(--space-4)`):
1. Логин — text input, `autocomplete="username"`, required
2. Пароль — password input, `autocomplete="current-password"`, required

- Submit: full-width, border-radius `var(--radius-pill)`, background `var(--color-brand-primary)`, hover `var(--color-brand-primary-hover)`, label "Войти" / loading state "Вход…" + 12px spinner inline left
- Root server error: below submit, `margin-top: var(--space-2)`, background `var(--color-error-subtle)`, border `1px solid var(--color-border-error)`, border-radius `var(--radius-input)`, padding `var(--space-2) var(--space-3)`, text `var(--text-sm)` `var(--color-error-text)`

### Register Form

Fields (2-column grid `grid-template-columns: 1fr 1fr; gap: var(--space-4)`):

| # | Field | Type | Span |
|---|---|---|---|
| 1 | Логин | text | 1 |
| 2 | Пароль | password | 1 |
| 3 | Имя | text | 1 |
| 4 | Фамилия | text | 1 |
| 5 | Отчество | text | 1 |
| 6 | Дата рождения | date | 1 |
| 7 | Зачётная книжка | text | 1 |
| 8 | Номер группы | text | 1 |
| 9 | Институт | text | 2 (full width) |
| 10 | Телефон | tel | 1 |
| 11 | О себе | textarea rows=3 | 2 (full width) |

Social links block (full width, below grid):
- Label "Социальные сети", `var(--text-sm)` `var(--weight-semibold)`, `margin-bottom: var(--space-2)`
- "+ Добавить ссылку" ghost button, size sm, icon `Plus` 14px
- Each link row: `display: flex; gap: var(--space-2); align-items: center` — Платформа input (width 120px) + Ссылка input (flex 1) + trash icon button (`Trash2` 14px, color `var(--color-error)` on hover)

- Submit: full-width pill radius, label "Зарегистрироваться" / loading "Регистрация…" + spinner
- Root server error: same anatomy as login form

---

## Section 4 — Division Explorer

### 4a — Layout

- Explorer occupies `calc(100vh - var(--topbar-height))`, `overflow: hidden`
- Toolbar strip: 48px height, `background: var(--color-surface)`, `border-bottom: 1px solid var(--color-border)`, `padding: 0 var(--space-4)`, `display: flex; align-items: center; justify-content: space-between`
  - Left: "Оргструктура" `var(--text-md)` `var(--weight-bold)` `var(--font-body)`
  - Right: "Обновить" (ghost sm) + "+ Подразделение" (primary sm, border-radius `var(--radius-pill)`)
- Below toolbar: flex row `height: calc(100% - 48px)`
  - Tree canvas: `flex: 1; position: relative; overflow: hidden`
  - Detail panel: `width: var(--detail-panel-width)` = 380px, always rendered, never toggled, `border-left: 1px solid var(--color-border)`, `box-shadow: var(--shadow-panel)`, `display: flex; flex-direction: column`

### 4b — Tree Nodes (LOD Tiers)

**Child sizing rule — option (a): fixed dimensions per depth level**

| Depth | Width | Height |
|---|---|---|
| 0 (root) | 200px | 64px |
| 1 | 176px | 56px |
| 2 | 156px | 50px |
| 3+ | 140px | 46px |

Dimensions are fixed; they do not change on zoom within a tier. ELK params: `nodeSep: 24`, `rankSep: 64`.

---

**Tier: micro** (zoom < 0.35)
- Node dimensions: per depth table above
- Content: colored left accent bar only — 4px wide, full node height, `var(--color-brand-primary)`; no text
- Background: `var(--color-node-bg)`; archived: `var(--color-node-archived-bg)`
- Border default: `1px solid var(--color-node-border)`
- Border hover: `1px solid var(--color-node-border-hover)`
- Border selected: `2px solid var(--color-node-border-selected)` + `box-shadow: var(--color-node-shadow-selected)`
- Border-radius: `var(--radius-card)`

**Tier: compact** (zoom 0.35 – 0.65)
- Node dimensions: per depth table above
- Content: `short_name` only — centered, `var(--text-xs)` `var(--weight-semibold)` `var(--color-text-primary)`, single line, `overflow: hidden; text-overflow: ellipsis; white-space: nowrap`
- Padding: `var(--space-2) var(--space-3)`
- Border rules: same as micro

**Tier: standard** (zoom 0.65 – 1.10)
- Node dimensions: per depth table above
- Content (top to bottom, left-aligned):
  - `short_name`: `var(--text-xs)` `var(--weight-semibold)` `var(--color-text-secondary)`
  - `name`: `var(--text-sm)` `var(--weight-bold)` `var(--color-text-primary)`, max 2 lines, `overflow: hidden; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical`
- Padding: `var(--space-2) var(--space-3)`
- Border rules: same as micro

**Tier: detail** (zoom > 1.10)
- Node dimensions: per depth table above
- Content:
  - Top-left: `short_name` `var(--text-xs)` `var(--weight-semibold)` `var(--color-text-muted)`
  - Below: `name` `var(--text-sm)` `var(--weight-bold)` `var(--color-text-primary)`, max 2 lines
  - Bottom-right: children count badge — `var(--text-xs)`, background `var(--color-surface-subtle)`, border `1px solid var(--color-border)`, padding `2px 6px`, border-radius `var(--radius-full)`, label `"{n} подр."` (hidden if 0)
  - Bottom-left: archived dot — 6px circle `var(--color-warning)`, visible only if archived
- Padding: `var(--space-2) var(--space-3)`
- Border rules: same as micro

### 4c — Tree Edges and Viewport Controls

- Edge type: bezier (cubic)
- Color: `var(--color-edge)` = #B0BAC4
- Stroke width: 1.5px
- No arrowheads
- Edges are not interactive

**Minimap:**
- Position: bottom-left, 12px from bottom, 12px from left of canvas
- Dimensions: 140×90px
- Background: `var(--color-surface)`, border `1px solid var(--color-border)`, border-radius `var(--radius-card)`
- Visible when tree has ≥ 2 nodes; hidden when empty

**Zoom controls:**
- Position: bottom-right, 12px from bottom, 12px from right of canvas
- Style: vertical button group (32×32px each), background `var(--color-surface)`, border `1px solid var(--color-border)`, border-radius `var(--radius-card)`, shadow `var(--shadow-sm)`, divider `1px solid var(--color-border)` between buttons
- Buttons: "+" (zoom in) and "−" (zoom out), ghost style, icon-only

### 4d — Empty Tree State

- Rendered in canvas area when tree has no root children
- EmptyStateCard centered absolutely:
  - Icon: `GitFork` 40px `var(--color-text-muted)`
  - Heading: "Оргструктура пуста"
  - Body: "Создайте первое подразделение, чтобы начать строить оргструктуру."
  - CTA (if user has create permission): "Создать подразделение", primary, border-radius `var(--radius-pill)`
  - CTA click: sets `panelMode = "create"`, clears node selection, opens create form in detail panel

### 4e — Detail Panel

Always rendered at fixed right position. Never a toggle or overlay.

**Header (48px, sticky within panel):**
- Background: `var(--color-surface)`, border-bottom `1px solid var(--color-border)`, padding `0 var(--space-4)`
- Layout: `display: flex; flex-direction: column; justify-content: center`
- **When node selected:**
  - Line 1 (breadcrumb): ancestor names separated by ` › `, `var(--text-xs)` `var(--color-text-muted)`, single line, `overflow: hidden; text-overflow: ellipsis; white-space: nowrap`
  - Line 2: selected division `name`, `var(--text-sm)` `var(--weight-bold)` `var(--color-text-primary)`, single line truncated
- **When nothing selected:**
  - Single label "Подразделение", `var(--text-sm)` `var(--weight-semibold)` `var(--color-text-muted)`
- No close button (panel is permanent)

**Panel body** — `overflow-y: auto; flex: 1; padding: var(--space-4)`

Section separator pattern: `border-top: 1px solid var(--color-border); padding-top: var(--space-4); margin-top: var(--space-4)`

Section label pattern: `var(--text-xs)` `var(--weight-semibold)` `var(--color-text-muted)` uppercase letter-spacing 0.06em, `margin-bottom: var(--space-2)`

**Sections in order when node selected:**

1. **Метаданные**
   - Full name: `var(--text-lg)` `var(--weight-bold)` `var(--font-body)` `var(--color-text-primary)`
   - Short name: `@{short_name}`, `var(--text-sm)` `var(--color-text-muted)`, `margin-top: var(--space-1)`
   - Description: `var(--text-sm)` `var(--color-text-secondary)`, `margin-top: var(--space-2)`, max 4 lines `-webkit-line-clamp: 4`
   - Archived badge (if archived): pill, background `var(--color-warning-subtle)`, text `var(--color-warning-text)`, `var(--text-xs)`, label "Архив", `margin-top: var(--space-2)`

2. **Действия** — section label "Действия"
   - Button row: `display: flex; gap: var(--space-2)`
   - "Редактировать" — outline sm; disabled if node is root or already in edit mode
   - "Архивировать" — outline sm, color `var(--color-error)`, border `var(--color-error)`; disabled if root or already archived
   - "Архивировать" click → opens ConfirmDialog (does NOT execute immediately):
     - Title: "Архивировать подразделение?"
     - Body: "Подразделение «{name}» и все его дочерние подразделения будут переведены в архив. Это действие можно отменить позже."
     - Cancel: "Отмена" (outline)
     - Confirm: "Архивировать" (background `var(--color-error)`, color white)
     - On confirm: execute; success → toast "Подразделение архивировано"

3. **Редактирование** (visible only when edit mode active) — section label "Редактирование"
   - DivisionEditForm inline, mode `"edit"`

4. **Новое подразделение** (visible only when create mode active) — section label "Новое подразделение"
   - DivisionEditForm inline, mode `"create"`

5. **Должности** (visible when node is not root) — section label "Должности"
   - PositionListPanel inline

6. **Участники** — section label "Участники"
   - PositionMembersList inline
   - Prompt when no position selected: `var(--text-sm)` `var(--color-text-muted)` centered "Выберите должность выше, чтобы увидеть участников"

**Empty/prompt state (no node selected):**
- Panel body shows centered EmptyStateCard:
  - Icon: `MousePointerClick` 32px `var(--color-text-muted)`
  - Heading: "Выберите подразделение"
  - Body: "Нажмите на узел дерева, чтобы увидеть его данные и действия."
  - No CTA

**Position row selection:**
- Selected row: background `var(--color-success-subtle)`, left border `3px solid var(--color-brand-primary)`
- Not selected: background transparent, left border `3px solid transparent`
- Selection updates PositionMembersList below

### 4f — Division Create/Edit Form

Form appears inline inside the detail panel section. Not a modal.

Fields (stacked, gap `var(--space-3)`):

| # | Field | Type | Required |
|---|---|---|---|
| 1 | Полное название | text | yes |
| 2 | Краткое название | text | yes |
| 3 | Описание | textarea rows=3 | no |
| 4 | Родительское подразделение | select | yes (create); disabled if root on edit |

- Validation: per-field errors use canonical anatomy (Section 10)
- Button row: `display: flex; justify-content: flex-end; gap: var(--space-2); margin-top: var(--space-4)`
  - "Отмена" (outline, closes form) left of submit
  - Create mode submit: "Создать" / "Создание…" + spinner; success → toast "Подразделение создано"; form closes
  - Edit mode submit: "Сохранить" / "Сохранение…" + spinner; success → toast "Изменения сохранены"; form closes

---

## Section 5 — Roles Management Page

### Layout

- Page padding: `var(--space-6)`
- Page title "Управление ролями": `var(--text-xl)` `var(--weight-bold)` `var(--font-body)`, `margin-bottom: var(--space-4)`
- Two-column grid: `grid-template-columns: minmax(0, 1fr) 360px; gap: var(--space-6)`

### Role List (left column)

- Card: background `var(--color-surface)`, border-radius `var(--radius-card)`, box-shadow `var(--shadow-sm)`, padding `var(--space-4)`
- Card header row: label "Роли" `var(--text-base)` `var(--weight-bold)` + "Создать роль" primary pill button (admin only), right-aligned
- List: `display: flex; flex-direction: column; gap: var(--space-2); margin-top: var(--space-3)`

**Role list item anatomy:**
- Container: `display: flex; align-items: center; justify-content: space-between; padding: var(--space-2) var(--space-3); border-radius: var(--radius-card); border: 1px solid var(--color-border); cursor: pointer; transition: var(--transition-fast)`
- Hover (not selected): background `var(--color-surface-subtle)`
- Selected: background `var(--color-success-subtle)`, border-color `var(--color-brand-primary)`
- Left: role name `var(--text-sm)` `var(--weight-semibold)` `var(--color-text-primary)` + kind tag below `var(--text-xs)` `var(--color-text-muted)`
- Right (admin only): `Pencil` icon button (ghost sm icon-only 28×28px) + `Trash2` icon button (ghost sm icon-only 28×28px, color `var(--color-error)` on hover), gap `var(--space-1)`

- Loading: SectionSkeleton 5 rows
- Empty: EmptyStateCard, icon `ShieldOff` 36px, heading "Роли не созданы", body "Создайте первую роль для назначения должностям.", CTA "Создать роль" (admin only)
- Error: EmptyStateCard, icon `AlertCircle` 36px, heading "Ошибка загрузки ролей", body "{error}", CTA "Повторить"

### Editor Panel (right column)

- Card: same style, `align-self: start`
- Card header: "Редактор роли" `var(--text-base)` `var(--weight-bold)`
- **Idle (admin, nothing selected):** centered content: `MousePointerClick` 28px `var(--color-text-muted)` + "Выберите роль для редактирования или создайте новую." `var(--text-sm)` `var(--color-text-muted)`, `padding: var(--space-8) var(--space-4)`
- **Non-admin:** "Управление ролями доступно только администраторам." `var(--text-sm)` `var(--color-text-muted)`, `padding: var(--space-4)`
- **Create/edit mode:** form with fields below

**Permission section layout:**
- Section label "Разрешения" `var(--text-xs)` uppercase `var(--color-text-muted)`, margin-bottom `var(--space-2)`
- Each permission row: `display: flex; align-items: center; gap: var(--space-3); padding: var(--space-1) 0`
  - Checkbox 16×16, accent border `var(--color-brand-primary)` when checked
  - Permission name `var(--text-sm)` `var(--color-text-primary)`, flex 1
  - Scope select: width 96px, `var(--radius-input)`, `var(--text-xs)`
- Groups separated by category headers: `var(--text-xs)` uppercase `var(--color-text-muted)` letter-spacing 0.06em, `margin-top: var(--space-3) margin-bottom: var(--space-1)`

- Error display: canonical form anatomy (Section 10)
- Success: toast "Роль сохранена" (create/edit), toast "Роль удалена" (delete)

**Delete confirmation dialog:**
- Title: "Удалить роль?"
- Body: "Роль «{name}» будет удалена безвозвратно. Убедитесь, что она не назначена ни одной должности."
- Cancel: "Отмена" (outline)
- Confirm: "Удалить" (background `var(--color-error)`, color white)

---

## Section 6 — Profile Page

### Layout

- Page padding: `var(--space-6)`
- Two-column grid: `grid-template-columns: minmax(0, 1fr) 320px; gap: var(--space-6)`

### Profile Summary (left column)

- Card: background `var(--color-surface)`, border-radius `var(--radius-card)`, box-shadow `var(--shadow-sm)`, padding `var(--space-6)`
- Card header row: full name `var(--text-lg)` `var(--weight-bold)` + "Редактировать профиль" outline pill button right-aligned
- Definition list below, `margin-top: var(--space-4)`, `display: grid; grid-template-columns: 120px 1fr; gap: var(--space-2) var(--space-4); align-items: baseline`
  - Term: `var(--text-sm)` `var(--weight-semibold)` `var(--color-text-muted)`
  - Value: `var(--text-sm)` `var(--color-text-primary)`
- Fields shown: Логин, Имя, Фамилия, Отчество, Дата рождения, Зачётная книжка, Группа, Институт, Телефон, О себе
- Social links: chip list below definition list, `margin-top: var(--space-3)`, `display: flex; flex-wrap: wrap; gap: var(--space-2)`
  - Each chip: `var(--text-xs)`, border `1px solid var(--color-border)`, border-radius `var(--radius-pill)`, padding `2px 10px`, color `var(--color-text-link)`, underline on hover

### Edit Form (left column, replaces summary card)

- Summary card hides when editing
- Edit card same style as summary
- Card title "Редактирование профиля"
- Fields: same as register minus login and password, 2-column grid `grid-template-columns: 1fr 1fr; gap: var(--space-4)`; О себе full-width
- Submit: "Сохранить" / "Сохранение…" + spinner; success → toast "Профиль обновлён"; summary card returns
- Cancel: "Отмена" (outline); summary returns, no changes

### Memberships (right column)

- Card: same style
- Card title "Членства" `var(--text-base)` `var(--weight-bold)`
- List: `display: flex; flex-direction: column; gap: var(--space-2); margin-top: var(--space-3)`

**Membership item anatomy:**
- Container: `display: flex; align-items: flex-start; justify-content: space-between; padding: var(--space-2) var(--space-3); border-radius: var(--radius-card); border: 1px solid var(--color-border)`
- Left: position name `var(--text-sm)` `var(--weight-semibold)` + division · role `var(--text-xs)` `var(--color-text-muted)`
- Right: `Trash2` 14px icon button ghost sm, `var(--color-error)` on hover

**Remove mechanic (inline, not modal):**
- Click trash → row expands inline: "Исключить из «{position_name}»?" `var(--text-sm)` `var(--color-text-primary)` + "Исключить" destructive sm + "Отмена" outline sm
- Confirm → execute; success → toast "Участие удалено"; row removed with fade-out 150ms
- Cancel → row collapses back to normal

- Loading: SectionSkeleton 3 rows
- Empty: EmptyStateCard, icon `Users` 36px, heading "Нет членств", body "Назначения появятся после добавления в должность."

---

## Section 7 — Search Page

### Layout

- Page padding: `var(--space-6)`
- Single-column layout, max-width 960px, `margin: 0 auto`
- Page title "Поиск пользователей" `var(--text-xl)` `var(--weight-bold)`, `margin-bottom: var(--space-4)`
- Filter card on top; results list directly below (same page, no navigation)

### Filter Form

- Card: background `var(--color-surface)`, border-radius `var(--radius-card)`, box-shadow `var(--shadow-sm)`, padding `var(--space-5)`
- Field grid: `grid-template-columns: repeat(3, 1fr); gap: var(--space-4)`

| Field | Label | Type |
|---|---|---|
| first_name | Имя | text |
| last_name | Фамилия | text |
| middle_name | Отчество | text |
| login | Логин | text |
| group_number | Группа | text |
| institute | Институт | text |
| about | О себе | text |
| position_title | Должность | text |
| role_name | Роль | text |

- Below grid: checkbox "Включить архивные назначения", `var(--text-sm)`, `margin-top: var(--space-3)`
- Submit row: `display: flex; justify-content: flex-end; margin-top: var(--space-4)`; "Найти" primary pill button `min-width: 120px`

### Result List Item Anatomy

- Results count: "Найдено: {total}", `var(--text-sm)` `var(--color-text-muted)`, `margin: var(--space-4) 0 var(--space-2)`
- List: `display: flex; flex-direction: column; gap: var(--space-2)`
- Item container: `display: flex; align-items: center; gap: var(--space-3); padding: var(--space-3) var(--space-4); border-radius: var(--radius-card); border: 1px solid var(--color-border); background: var(--color-surface); transition: var(--transition-base)`
- Item hover: `transform: translateY(-2px); box-shadow: var(--shadow-card-hover)`
- Avatar: 36×36px circle, background `var(--color-brand-primary)`, color white, `var(--text-xs)` `var(--weight-bold)`, initials = first letter of first_name + first letter of last_name, uppercase
- Center: full name `var(--text-sm)` `var(--weight-semibold)` + login below `var(--text-xs)` `var(--color-text-muted)`
- Right: membership chips — each chip `var(--text-xs)`, border `1px solid var(--color-border)`, border-radius `var(--radius-pill)`, padding `2px 8px`; max 3 shown; "+N" overflow chip

### Loading / Empty / Error States

- Loading: SectionSkeleton 5 rows, `margin-top: var(--space-4)`
- Empty (submitted, 0 results): EmptyStateCard, icon `SearchX` 36px, heading "Ничего не найдено", body "Попробуйте изменить параметры поиска.", no CTA
- Not yet searched: results area not rendered
- Error: EmptyStateCard, icon `AlertCircle` 36px, heading "Ошибка поиска", body "{error}", CTA "Повторить"

---

## Section 8 — Audit Log Page

### Layout

- Page padding: `var(--space-6)`
- Page title "Журнал аудита" `var(--text-xl)` `var(--weight-bold)`, `margin-bottom: var(--space-4)`
- Two-column grid: `grid-template-columns: minmax(0, 1fr) 280px; gap: var(--space-6); align-items: start`

### Table (left column)

- Card: background `var(--color-surface)`, border-radius `var(--radius-card)`, box-shadow `var(--shadow-sm)`, padding `var(--space-4)`, `overflow-x: auto`
- Table: `width: 100%; border-collapse: collapse`
- Header row background: `var(--color-surface-subtle)`, border-bottom `2px solid var(--color-border)`

| Column | Label | Width |
|---|---|---|
| timestamp | Дата и время | 180px |
| event_type | Тип события | auto |
| subject_type | Тип объекта | 140px |
| subject_id | ID объекта | 220px |
| actor_id | Пользователь | 180px |

- `th`: `var(--text-xs)` `var(--weight-semibold)` uppercase `var(--color-text-muted)` letter-spacing 0.06em, padding `var(--space-2) var(--space-3)`, text-align left
- `td`: `var(--text-sm)` `var(--color-text-primary)`, padding `var(--space-2) var(--space-3)`, `border-bottom: 1px solid var(--color-border)`
- Row hover: background `var(--color-surface-subtle)`

### Loading / Empty / Error

- Loading: table not rendered; SectionSkeleton 5 rows in its place (no text loading row in table)
- Empty: render full table skeleton with one body row; row contains centered EmptyStateCard-inline layout: `ClipboardX` 24px + "Нет записей аудита" `var(--text-sm)` `var(--color-text-muted)`, spanning all columns
- Error: EmptyStateCard above table, icon `AlertCircle` 36px, heading "Ошибка загрузки", body "{error}", CTA "Повторить"

### Pagination

- Row below table, `margin-top: var(--space-3)`: `display: flex; align-items: center; gap: var(--space-4)`
- Left: "← Назад" outline sm; disabled if offset = 0
- Center: "Записи {offset+1}–{min(offset+limit, total)} из {total}", `var(--text-sm)` `var(--color-text-muted)`
- Right: "Вперёд →" outline sm; disabled if offset + limit ≥ total

### Filter Panel (right column)

- Card: background `var(--color-surface)`, border-radius `var(--radius-card)`, box-shadow `var(--shadow-sm)`, padding `var(--space-4)`
- Title "Фильтры" `var(--text-sm)` `var(--weight-bold)`, `margin-bottom: var(--space-3)`
- Fields stacked, gap `var(--space-3)`:
  1. Тип события — text
  2. Тип объекта — text
  3. ID объекта — text
- "Применить" primary full-width pill, `margin-top: var(--space-3)`
- "Сбросить" ghost full-width, `margin-top: var(--space-2)`; visible only when ≥ 1 filter field is non-empty

---

## Section 9 — Feedback System

### Page-Level Loading

- Pattern: spinner (no skeleton at page level)
- Spinner: 32px circle, `border: 3px solid var(--color-border); border-top-color: var(--color-brand-primary); border-radius: var(--radius-full); animation: spin 0.75s linear infinite`
- Centered in content area: `display: flex; flex-direction: column; align-items: center; justify-content: center; height: calc(100vh - var(--topbar-height))`
- Text below spinner: `var(--text-sm)` `var(--color-text-muted)`, `margin-top: var(--space-3)`, e.g., "Загрузка страницы…"

### Section-Level Loading

- Pattern: SectionSkeleton — shimmer rows
- Each row: `height: 16px; border-radius: var(--radius-input); background: linear-gradient(90deg, #E8EAED 25%, #F4F5F7 50%, #E8EAED 75%); background-size: 200% 100%; animation: shimmer 1.4s ease infinite`
- Row gap: `var(--space-2)`
- Row widths cycle: 100%, 80%, 90%, 65%, 85%
- Default 5 rows unless overridden

### Toast Notifications

- Position: fixed top-right, 16px from top, 16px from right, z-index 300
- Stack: downward, gap `var(--space-2)`, max 4 toasts visible; oldest auto-dismissed when 5th arrives
- Width: `var(--toast-width)` = 320px
- Container: background `var(--color-surface)`, border-radius `var(--radius-card)`, box-shadow `var(--shadow-toast)`, `overflow: hidden`
- Left accent bar: 4px wide full height, left-side radius only
- Inner layout: `display: flex; align-items: flex-start; gap: var(--space-2); padding: var(--space-3) var(--space-3) var(--space-3) var(--space-4)`
  - Icon 20×20px left
  - Title `var(--text-sm)` `var(--weight-semibold)`, optional body below `var(--text-xs)` `var(--color-text-muted)`
  - ×  button 20×20px top-right, ghost

| Variant | Accent bar | Icon | Icon color |
|---|---|---|---|
| success | `var(--color-success)` | `CheckCircle2` | `var(--color-success)` |
| error | `var(--color-error)` | `XCircle` | `var(--color-error)` |
| info | `var(--color-info)` | `Info` | `var(--color-info)` |

- Auto-dismiss: 4000ms
- Entry animation: `translateX(calc(100% + 16px))` → `translateX(0)`, 200ms ease
- Exit animation: `opacity: 1` → `opacity: 0` + `translateX(calc(100% + 16px))`, 150ms ease

### Global Error Banner

- Triggered: network failure (fetch throws TypeError), server 5xx
- NOT triggered: 4xx, 403, validation errors
- Position: fixed, `top: var(--topbar-height)`, left 0, right 0, z-index 90
- Background: `var(--color-error-subtle)`, border-bottom `1px solid var(--color-border-error)`, padding `var(--space-3) var(--space-6)`
- Layout: `display: flex; align-items: center; gap: var(--space-3)`
  - `AlertCircle` 16px `var(--color-error)`
  - Message text `var(--text-sm)` `var(--color-error-text)`, flex 1
  - "Повторить" ghost sm (if retryable)
  - ×  dismiss button 24×24px ghost, far right
- Does NOT auto-dismiss; user must click ×

### Inline Field Error

- Position: directly below input, `margin-top: 4px`
- Typography: `var(--text-xs)` `var(--weight-medium)` `var(--color-error-text)`
- No icon; text only

### Confirmation Dialogs

- Backdrop: fixed fullscreen, z-index 200, background `var(--color-surface-overlay)`, `animation: fadeIn 150ms ease`
- Dialog: centered, width `var(--modal-width)` = 440px, background `var(--color-surface)`, border-radius `var(--radius-card)`, box-shadow `var(--shadow-modal)`, padding `var(--space-6)`
- Title: `var(--text-md)` `var(--weight-bold)` `var(--color-text-primary)`, `margin-bottom: var(--space-2)`
- Body: `var(--text-sm)` `var(--color-text-secondary)`, `margin-bottom: var(--space-6)`
- Button row: `display: flex; justify-content: flex-end; gap: var(--space-3)`
  - Cancel: outline, label "Отмена"
  - Confirm destructive: background `var(--color-error)`, color white, border none, border-radius `var(--radius-card)`
  - Confirm non-destructive: primary (`var(--color-brand-primary)`)
- Close on: Cancel click, Escape key, backdrop click

### Empty States

- Layout: `display: flex; flex-direction: column; align-items: center; text-align: center; padding: var(--space-12) var(--space-6)`
- Icon: lucide-react, 36px, `var(--color-text-muted)`
- Heading: `var(--text-base)` `var(--weight-semibold)` `var(--color-text-primary)`, `margin-top: var(--space-3)`
- Body: `var(--text-sm)` `var(--color-text-muted)`, max-width 40ch, `margin-top: var(--space-1)`
- CTA (optional): primary pill button, `margin-top: var(--space-4)`

---

## Section 10 — Form System (canonical anatomy)

### Label

- Position: above input, `margin-bottom: var(--space-1)`
- Typography: `var(--text-sm)` `var(--weight-semibold)` `var(--color-text-primary)`
- Required indicator: ` *` suffix, `color: var(--color-error)`, `aria-hidden="true"`

### Input Default State

- Border: `1px solid var(--color-border)`
- Background: `var(--color-surface)`
- Border-radius: `var(--radius-input)` = 5px
- Padding: `var(--space-2) var(--space-3)`
- Height: 36px (single-line inputs)
- Font: `var(--text-sm)` `var(--font-body)` `var(--weight-regular)` `var(--color-text-primary)`
- Placeholder: `var(--color-text-muted)`

### Input Focus State

- `outline: none`
- `border-color: var(--color-border-focus)`
- `box-shadow: 0 0 0 3px rgba(70,174,91,0.20)`

### Input Error State

- `border-color: var(--color-border-error)`
- `box-shadow: 0 0 0 3px rgba(217,79,79,0.15)`

### Input Disabled State

- `background: var(--color-surface-subtle)`
- `color: var(--color-text-disabled)`
- `border-color: var(--color-border)`
- `cursor: not-allowed`
- `opacity: 0.65`

### Select Element

- Same border, background, radius, padding, height as text input
- Dropdown: background `var(--color-surface)`, border `1px solid var(--color-border)`, border-radius `var(--radius-card)`, box-shadow `var(--shadow-md)`
- Option hover: background `var(--color-surface-subtle)`
- Option selected: background `var(--color-success-subtle)`, color `var(--color-success-text)` `var(--weight-semibold)`

### Textarea

- Same border, background, radius, padding as text input
- `resize: vertical`
- `min-height: 80px`

### Field-Level Error Message

- Position: below input, `margin-top: 4px`
- Typography: `var(--text-xs)` `var(--weight-medium)` `var(--color-error-text)`
- No icon prefix

### Root / Server Error

- Position: below all fields, above submit button, `margin-bottom: var(--space-3)`
- Container: background `var(--color-error-subtle)`, border `1px solid var(--color-border-error)`, border-radius `var(--radius-input)`, padding `var(--space-2) var(--space-3)`
- Text: `var(--text-sm)` `var(--color-error-text)`

### Submit Button

- Position: right-aligned in button row (not full-width, unless form width < 300px)
- Loading state: disabled + 12px spinner (white, `border: 2px solid rgba(255,255,255,0.3); border-top-color: white; animation: spin 0.75s linear infinite`) prepended + loading label text
- Disabled condition: `isSubmitting === true` OR required fields are empty
- Border-radius: `var(--radius-card)` = 10px

### Success

- On successful submit: fire toast notification (Section 9)
- Create forms: clear field values after success
- Edit forms: keep field values (reflect saved state)
- Do NOT silently close or reset without a toast

---

## Section 11 — String Localization

| English | Russian |
|---|---|
| Signing in… | Вход… |
| Registering… | Регистрация… |
| Saving… | Сохранение… |
| Assigning… | Назначение… |
| Loading… | Загрузка… |
| No audit entries | Нет записей аудита |
| Apply filters | Применить |
| Hint: click to open details | Нажмите на узел, чтобы открыть панель |
| child(ren) | подр. |
| Add link | Добавить ссылку |
| Checking active session | Проверка сессии… |
| Loading account context… | Загрузка данных пользователя… |
| Edit | Редактировать |
| Archive | Архивировать |
| Create | Создать |
| Delete | Удалить |
| Cancel | Отмена |
| Confirm | Подтвердить |
| Remove | Исключить |
| Save | Сохранить |
| Search | Найти |
| Login | Вход |
| Register | Регистрация |
| Logout | Выйти |
| Logging out… | Выход… |
| Password | Пароль |
| First name | Имя |
| Last name | Фамилия |
| Middle name | Отчество |
| Birth date | Дата рождения |
| Gradebook number | Зачётная книжка |
| Group number | Номер группы |
| Institute | Институт |
| Phone | Телефон |
| About | О себе |
| Social links | Социальные сети |
| No social links added. | (no message; only the add button shown) |
| Role Management | Управление ролями |
| Create Role | Создать роль |
| Only admin can create, edit, or delete roles. | Управление ролями доступно только администраторам. |
| No roles yet | Роли не созданы |
| Role editor | Редактор роли |
| Select a role to edit or start creating a new role. | Выберите роль для редактирования или создайте новую. |
| Role list unavailable | Ошибка загрузки ролей |
| Retry | Повторить |
| Read-only mode | Только просмотр |
| Role is in use by one or more positions. | Роль назначена одной или нескольким должностям. |
| Operation is forbidden for current scope. | Операция недоступна в текущем контексте. |
| Operation is forbidden. | Операция запрещена. |
| Memberships | Членства |
| Current position assignments returned by backend. | Текущие назначения на должности. |
| No memberships assigned yet. | Нет членств. |
| Refresh | Обновить |
| Refreshing… | Обновление… |
| Confirm remove? | Исключить? |
| Removing… | Удаление… |
| Filter first, then review results below without leaving the page. | Задайте параметры фильтрации и просматривайте результаты здесь же. |
| Audit log | Журнал аудита |
| Filters | Фильтры |
| Prev | Назад |
| Next | Вперёд |
| Event type | Тип события |
| Subject type | Тип объекта |
| Subject id | ID объекта |
| timestamp | Дата и время |
| event_type | Тип события |
| subject_type | Тип объекта |
| subject_id | ID объекта |
| actor_id | Пользователь |
| Create position | Создать должность |
| Select a role | Выберите роль |
| Max count | Максимум участников |
| No limit | Без ограничений |
| Loading roles… | Загрузка ролей… |
| Title | Название |
| Role | Роль |
| Failed to load members. | Не удалось загрузить участников. |
| Authenticated as | Вы вошли как |
| Social link platform placeholder "vk" | vk |
| Social link value placeholder "https://…" | https://… |
| Use your activist account credentials. | Введите данные вашей учётной записи. |
| Create a new activist profile. | Создайте новую учётную запись. |
| No positions yet ("Должностей пока нет") | Должностей пока нет |
| No members ("Нет участников") | Нет участников |

---

## Section 12 — Component Change Log

| Component | File | Change type | What changes |
|---|---|---|---|
| AppShell | src/app/AppShell.tsx | redesign | Remove sidebar slot; add `padding-top: var(--topbar-height)`; background `var(--color-canvas)`; render TopNavBar |
| AppSidebar | src/app/AppSidebar.tsx | delete | Entirely replaced by TopNavBar; file removed |
| TopNavBar | src/app/TopNavBar.tsx (new) | redesign | New component: 56px fixed top bar, brand left, nav center, user+logout right |
| routeContracts | src/app/routeContracts.ts | fix | Remove sidebar-specific dimension contracts; add `topbar-height` constant |
| App | src/app/App.tsx | fix | Wrap root with ToastProvider; render ConfirmDialog portal at root level |
| SessionGate | src/features/auth/SessionGate.behavior.test.tsx | fix + localize | Show login first; toggle tabs instead of side-by-side; localize all strings; update tests |
| LoginForm | src/features/auth/LoginForm.tsx | redesign + localize | Single-column fields; pill submit; brand primary color; canonical form anatomy; Russian strings |
| RegisterForm | src/features/auth/RegisterForm.tsx | redesign + localize | 2-column grid; social links block with add/remove; canonical form anatomy; Russian strings |
| DivisionExplorerPage | src/features/divisions/DivisionExplorerPage.tsx | redesign | Permanent split layout; toolbar anatomy; wire CTA to empty tree create; archive → ConfirmDialog |
| DivisionExplorerPage | src/features/divisions/DivisionExplorerPage.tsx | fix | Detail panel header shows selected name + breadcrumb; empty prompt state when nothing selected |
| DivisionTreeViewport | src/features/divisions/tree/DivisionTreeViewport.tsx | redesign | 4 LOD tiers with fixed dimensions; bezier edges #B0BAC4 1.5px; minimap bottom-left; zoom controls bottom-right |
| divisionTreeLayout | src/features/divisions/tree/divisionTreeLayout.ts | fix | ELK nodeSep=24, rankSep=64; node width/height per depth level |
| treeTypes | src/features/divisions/tree/treeTypes.ts | fix | Add `lod: 'micro' \| 'compact' \| 'standard' \| 'detail'` and `depth: number` to node data type |
| DivisionEditForm | src/features/divisions/DivisionEditForm.tsx | redesign + localize | Stacked fields; canonical form anatomy; Russian labels; success toast on create/edit |
| PositionListPanel | src/features/positions/PositionListPanel.tsx | fix + localize | Selected row highlighted (green subtle bg + left border); triggers member list update; Russian strings |
| PositionMembersList | src/features/positions/PositionMembersList.tsx | fix + localize | Prompt when no position selected; localize all strings |
| PositionCreateForm | src/features/positions/PositionCreateForm.tsx | redesign + localize | Canonical form anatomy; Russian labels; success toast |
| AssignMemberForm | src/features/positions/AssignMemberForm.tsx | localize | Russian labels and button loading states |
| RoleManagementPage | src/features/roles/RoleManagementPage.tsx | redesign + localize | Two-column card layout; role list item anatomy; editor idle/non-admin states; Russian strings |
| RoleCreateForm | src/features/roles/RoleCreateForm.tsx | redesign + localize | Canonical form anatomy; permission checkbox+scope layout; success toast |
| RoleEditForm | src/features/roles/RoleEditForm.tsx | redesign + localize | Same as RoleCreateForm; pre-filled; delete confirmation via ConfirmDialog |
| ProfilePage | src/features/profile/ProfilePage.tsx | redesign + localize | Two-column grid; definition-list summary; edit form toggle; inline membership remove confirm |
| ProfileForm | src/features/profile/ProfileForm.tsx | redesign + localize | Canonical form anatomy; 2-column grid; success toast |
| SearchPanel | src/features/search/SearchPanel.tsx | redesign + localize | 3-column filter grid; localize info text |
| SearchResults | src/features/search/SearchResults.tsx | redesign + localize | Avatar+name+chips list items; hover lift; loading/empty/error states |
| AuditLogPage | src/features/audit/AuditLogPage.tsx | redesign + localize | Two-column grid; structured table headers; skeleton loading (not text row); localize all strings |
| EmptyStateCard | src/shared/ui/feedback/EmptyStateCard.tsx | fix | Add `icon` prop (ReactNode); enforce `var(--space-12)` vertical padding; update default strings to Russian |
| SectionSkeleton | src/shared/ui/feedback/SectionSkeleton.tsx | fix | Shimmer animation; alternating widths; configurable `rows` prop |
| GlobalApiBanner | src/shared/ui/feedback/GlobalApiBanner.tsx | fix | Position `top: var(--topbar-height)` (not 0); update severity colors to design token variables |
| ToastProvider | src/shared/ui/feedback/ToastProvider.tsx (new) | redesign | New: top-right stack, 4-toast max, 4s auto-dismiss, slide-in/fade-out animations, success/error/info variants |
| ConfirmDialog | src/shared/ui/feedback/ConfirmDialog.tsx (new) | redesign | New: modal overlay portal, title/body/cancel/confirm, destructive vs non-destructive confirm variant |
| button | src/components/ui/button.tsx | migrate | Replace all hardcoded hex values with design token CSS variables; update primary to `var(--color-brand-primary)` |
| form | src/components/ui/form.tsx | migrate | FormMessage uses `var(--color-error-text)`; FormLabel uses `var(--color-text-primary)` |
| index.css | src/index.css | redesign | Replace all tokens: indigo → green/teal brand palette; update font to Raleway; add all new token categories; add Raleway @import |
