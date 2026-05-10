# Анализ страниц верхней навигации

Дата: 2026-04-21
Проект: `activist-frontend`

## 1) Страницы из верхнего горизонтального меню

Маршруты из `src/app/routeContracts.ts` и `src/app/App.tsx`:

- `/` — Оргструктура (`DivisionTreePage` -> `DivisionExplorerPage`)
- `/roles` — Роли (`RoleManagementPage`)
- `/profile` — Профиль (`ProfilePage`)
- `/search` — Поиск (`SearchPanel`)
- `/audit-log` — Журнал (`AuditLogPage`, доступ по `can_view_audit_log`)

## 2) Краткий анализ каждой страницы

### `/` Оргструктура

- Full-screen layout: дерево + правая панель.
- Сложная локальная оркестрация: `useReducer` + много `useState`.
- Сильная концентрация логики в `DivisionExplorerPage`.
- Используется `@xyflow/react` для визуализации дерева.

Основные файлы:
- `src/features/divisions/DivisionTreePage.tsx`
- `src/features/divisions/DivisionExplorerPage.tsx`
- `src/features/divisions/tree/DivisionTreeViewport.tsx`
- `src/features/divisions/explorer/state/explorerReducer.ts`

### `/roles` Роли

- Двухколоночный layout (список ролей + редактор).
- CRUD-операции с ручным `loadRoles`.
- Формы вынесены отдельно (`RoleCreateForm`, `RoleEditForm`), но сильно дублируются.

Основные файлы:
- `src/features/roles/RoleManagementPage.tsx`
- `src/features/roles/RoleCreateForm.tsx`
- `src/features/roles/RoleEditForm.tsx`

### `/profile` Профиль

- Двухколоночный layout (профиль + членства).
- Параллельная загрузка профиля и членств через `Promise.all`.
- Inline-подтверждение удаления членства и локальная анимация fade-out.

Основные файлы:
- `src/features/profile/ProfilePage.tsx`
- `src/features/profile/ProfileForm.tsx`

### `/search` Поиск

- Центрированный контейнер, form-first подход.
- Ручное управление фильтрами и состояниями загрузки/ошибки.
- Результаты вынесены в `SearchResults`.

Основные файлы:
- `src/features/search/SearchPanel.tsx`
- `src/features/search/SearchResults.tsx`

### `/audit-log` Журнал

- Двухколоночный layout (таблица + фильтры).
- Draft/applied filters, пагинация на `offset/limit`.
- Ручной `load()` + обработка загрузки/ошибок.

Основной файл:
- `src/features/audit/AuditLogPage.tsx`

## 3) Проблемы и расхождения

- Непоследовательный нейминг:
  - Смешение `*Page` и `*Panel` для роутовых экранов.
  - `/` идёт через `DivisionTreePage` -> `DivisionExplorerPage`.
- Разные layout-паттерны:
  - full-screen, max-width container, разные правые колонки (`280/320/360/380`).
  - В `AppShell` зафиксирован `minWidth: 1280`.
- Стилизация преимущественно inline-style на уровне страниц, мало общих layout-компонентов.
- Повторяется парсинг ошибок API (`parseApiError`, `prettyApiError`, `toApiError`) в нескольких фичах.
- Непоследовательный доступ к API:
  - в основном `request(...)` из `shared/api/client.ts`,
  - но есть прямой `fetch` + локальный `toUrl` в `AssignMemberForm`.
- Дублирование форм ролей:
  - `RoleCreateForm` и `RoleEditForm` почти зеркальны.
- `regions` в `routeContracts` сейчас не влияет на реальный рендер shell.
- `EmptyStateCard` используется по-разному (новый `cta` и deprecated `ctaLabel/onCta`).

## 4) Что можно унифицировать

### Общий layout

Вынести в `src/shared/ui/layout/`:

- `PageRoot`
- `PageHeader`
- `PageContent`
- `PageGrid`
- `PageSidebar`
- `FullScreenPage` (для оргструктуры)

### Общие UI-блоки

- `PanelCard` (единая карточка панели)
- `SectionTitle`
- `InlineAlert`
- Единая обвязка для пустых/ошибочных/скелетон-состояний на странице

### Единые паттерны data fetching/state

- `useApiQuery` и `useApiMutation` (поверх `request`)
- Общий `shared/api/errors.ts`:
  - `parseApiError`
  - `mapAccessError`
  - `toFieldErrors`/`toRootError`
- Запрет прямого `fetch` в feature-слое без причины

## 5) Целевая структура страницы

Пример шаблона:

```tsx
<PageRoot data-testid="roles-page">
  <PageHeader title="Управление ролями" />
  <PageContent>
    <PageGrid sidebarWidth={360}>
      <PanelCard>
        {/* primary content */}
      </PanelCard>
      <PageSidebar>
        <PanelCard>
          {/* side content */}
        </PanelCard>
      </PageSidebar>
    </PageGrid>
  </PageContent>
</PageRoot>
```

## 6) Рекомендованный план рефакторинга

1. Ввести `shared/ui/layout` и мигрировать `/search` и `/audit-log` (минимальный риск).
2. Вынести общий `shared/api/errors.ts`, заменить локальные `parseApiError`.
3. Унифицировать API-вызовы: убрать прямой `fetch` из `AssignMemberForm`.
4. Объединить `RoleCreateForm` + `RoleEditForm` в `RoleForm` (`mode: create | edit`).
5. Привести нейминг роутовых компонентов к `*Page`:
   - `SearchPanel` -> `SearchPage`
   - рассмотреть `DivisionExplorerPage` как `DivisionsPage`.
6. Привести все top-nav страницы к единому page-contract для новых экранов.

## 7) Пример рефакторинга (до/после)

До (`SearchPanel`):

```tsx
<section data-testid="search-panel" style={{ maxWidth: 960, margin: "0 auto" }}>
  <h1>Поиск пользователей</h1>
  ...
</section>
```

После:

```tsx
<PageRoot data-testid="search-page">
  <PageHeader title="Поиск пользователей" />
  <PageContent maxWidth="md">
    <SearchFiltersForm />
    <SearchResults />
  </PageContent>
</PageRoot>
```

