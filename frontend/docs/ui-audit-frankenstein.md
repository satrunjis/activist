# Аудит фронтенда: диагноз "франкенштейна" UI/UX и интерфейсной архитектуры

Дата: 2026-04-21

## 1. Где именно в коде проявляется "франкенштейн"

- В дереве оргструктуры одновременно живут две модели состояния панели: `panelMode` в reducer и локальный `mode` в странице. Рендер панели фактически завязан на `mode`, а не на `state.panelMode`.
  - `src/features/divisions/DivisionExplorerPage.tsx`
  - `src/features/divisions/explorer/state/explorerReducer.ts`
- Правая панель дерева реализована как растущий набор условных секций (`Метаданные`, `Действия`, `Редактирование`, `Новое подразделение`, `Должности`, `Участники`) без единого композиционного контракта.
  - `src/features/divisions/DivisionExplorerPage.tsx`
- Есть дрейф от единых UI-примитивов: в shared feedback используются raw `<button>`, а в фичах `Button`-primitive.
  - `src/shared/ui/feedback/EmptyStateCard.tsx`
  - `src/shared/ui/feedback/ConfirmDialog.tsx`
  - `src/shared/ui/feedback/GlobalApiBanner.tsx`
  - `src/components/ui/button.tsx`
- Используются className-паттерны (`auth-form`, `form-grid`, `muted-text`, `position-row`, `badge`, `member-row`), но соответствующих CSS-определений в `src/index.css` нет.
  - `src/features/positions/PositionCreateForm.tsx`
  - `src/features/positions/AssignMemberForm.tsx`
  - `src/features/positions/PositionListPanel.tsx`
  - `src/features/positions/PositionMembersList.tsx`
  - `src/index.css`
- Документация частично не соответствует фактическому коду (описаны компоненты shell, которых физически нет в `src/app`).
  - `README.md`
  - `src/app/`

## 2. Все найденные create-flow паттерны

### Паттерн A: Side Editor (страница ролей)

- Открытие создания: кнопка в листинге ролей, форма в правом `aside`.
- Создание: `POST /api/v1/roles`.
- Валидация: `react-hook-form + zodResolver`.
- Ошибки: `RootError`.
- Файлы:
  - `src/features/roles/RoleManagementPage.tsx`
  - `src/features/roles/RoleCreateForm.tsx`
  - `src/features/roles/RoleEditForm.tsx`

### Паттерн B: Persistent Right Action Panel (оргдерево)

- Открытие создания: CTA в хедере панели и CTA в empty state.
- Форма встроена как секция правой панели, не модалка.
- Создание оркестрируется page-компонентом (`handleCreate`), форма только собирает payload.
- Валидация: ручная в форме (без resolver).
- Файлы:
  - `src/features/divisions/DivisionExplorerPage.tsx`
  - `src/features/divisions/DivisionEditForm.tsx`

### Паттерн C: Embedded Toggle Create (должности)

- Открытие создания: локальный toggle `showCreateForm` внутри блока должностей.
- Создание: `POST /api/v1/divisions/{id}/positions`.
- Валидация: `safeParse` вручную, не resolver.
- Ошибки: root через `FormMessage`.
- Файлы:
  - `src/features/positions/PositionListPanel.tsx`
  - `src/features/positions/PositionCreateForm.tsx`

### Паттерн D: Nested Inline Create (назначение участника)

- Открытие: toggle формы внутри выбранной строки должности.
- Создание: raw `fetch` + локальный `toUrl` + ручной CSRF.
- Ошибки: root через `FormMessage`.
- Файлы:
  - `src/features/positions/PositionListPanel.tsx`
  - `src/features/positions/AssignMemberForm.tsx`

### Паттерн E: Tabbed Create (регистрация)

- Открытие: переключение таба `Вход/Регистрация`.
- Создание: `POST /api/v1/auth/register`.
- Валидация: resolver + shared schema.
- Файлы:
  - `src/features/auth/AuthScreen.tsx`
  - `src/features/auth/RegisterForm.tsx`

## 3. Подробный разбор правой панели дерева

- Панель всегда присутствует, но ее контент собирается последовательностью условных рендеров.
- Поведение панели основано на комбинации:
  - выбранного узла,
  - локального `mode` (`create/edit/null`),
  - выбранной должности,
  - состояний загрузки участников.
- Внутри панели есть второй уровень ad-hoc-композиции:
  - `PositionListPanel` добавляет собственные локальные create/assign формы через toggle.
- Результат:
  - режимы пересекаются,
  - отсутствует единая модель surface-state,
  - панель масштабируется через "добавление блока", а не через типизированные сценарии.

Ключевые файлы:

- `src/features/divisions/DivisionExplorerPage.tsx`
- `src/features/divisions/explorer/state/explorerReducer.ts`
- `src/features/positions/PositionListPanel.tsx`

## 4. Основные причины фрагментации

- Нет общего набора abstractions уровня create/edit/panel surface.
- Feature-специфичная оркестрация разрослась в page-компонентах:
  - `DivisionExplorerPage` совмещает data orchestration, CRUD, panel rendering, async flow.
- Дублирование CRUD-паттернов:
  - разные стратегии открытия форм,
  - разные стратегии submit/error/cancel.
- Разные API-интеграции в похожих сценариях:
  - `request()` vs raw `fetch`.
- Разнобой в обработке ошибок:
  - много локальных `parseApiError`,
  - ветки на `instanceof Error` при том, что shared client бросает `ApiRequestError`-object.
- Разнобой в presentation-слое:
  - `RootError` vs `FormMessage` для root,
  - `Button` vs raw `<button>`,
  - inline-style доминирует над reusable layout-компонентами.
- Отсутствует "живой" системный UI-контракт:
  - `docs/design-spec.md` и `README.md` частично расходятся с текущей реализацией.

## 5. Что можно унифицировать

- Единый `EntityCreateSurface` для ролей, подразделений, должностей, назначений.
- Единый `EntityFormFrame`:
  - заголовок,
  - body,
  - action bar,
  - root error,
  - submitting state.
- Единый `ActionPanel` для правых/боковых рабочих областей.
- Единый `AsyncStateView` (`idle/loading/empty/error/success`).
- Единый адаптер API-ошибок:
  - `code -> user message`,
  - field errors,
  - retry policy.
- Единый контракт confirm-flow:
  - либо modal confirm, либо inline confirm, но один стиль и API.
- Единые layout primitives:
  - top-level page scaffold,
  - split workspace,
  - side editor area.

## 6. Целевая архитектура UI

Предлагаемая структура:

- `src/shared/ui/layout/`
  - `PageScaffold.tsx`
  - `SplitWorkspace.tsx`
  - `SideEditorLayout.tsx`
- `src/shared/ui/surfaces/`
  - `ActionPanel.tsx`
  - `EntityCreateSurface.tsx`
  - `EntityEditSurface.tsx`
- `src/shared/ui/forms/`
  - `EntityFormFrame.tsx`
  - `FormActions.tsx`
  - `FormRootError.tsx`
- `src/shared/ui/states/`
  - `AsyncStateView.tsx`
- `src/shared/api/`
  - `errorAdapter.ts`

Для дерева оргструктуры:

- Перевести правую панель на единую state machine-модель режима:
  - `empty`
  - `division.view`
  - `division.edit`
  - `division.create`
  - `position.create`
  - `membership.create`
- Удалить дублирующий локальный `mode` и использовать один источник истины.

## 7. Пошаговый план рефакторинга

1. Быстрый low-risk этап:
   - унифицировать обработку ошибок,
   - убрать ветки, завязанные на `instanceof Error` в create/edit формах.
2. Вынести общий `EntityFormFrame` и применить в:
   - `RoleCreateForm`,
   - `RoleEditForm`,
   - `DivisionEditForm`,
   - `PositionCreateForm`,
   - `AssignMemberForm`.
3. Нормализовать правую панель дерева:
   - единый `panelMode`,
   - вынести panel renderer в отдельный компонент.
4. Миграция create-flow по порядку:
   - роли -> подразделения -> должности -> назначения.
5. Унифицировать confirm-паттерны:
   - удаление роли,
   - архивирование подразделения,
   - исключение из членства.
6. Внедрить layout primitives на top-level страницах:
   - `/roles`, `/profile`, `/search`, `/audit-log`.
7. Синхронизировать документацию:
   - `README.md`,
   - `docs/design-spec.md`,
   - зафиксировать системный UI-контракт.

Наиболее рискованные зоны для регрессий:

- `DivisionExplorerPage` и вложенные формы панели.
- `PositionListPanel` (пересечение create/assign/selection/load states).
- API error mapping и forbidden-flow.

## 8. Критичные файлы/компоненты для приоритета

1. `src/features/divisions/DivisionExplorerPage.tsx`
2. `src/features/divisions/explorer/state/explorerReducer.ts`
3. `src/features/positions/PositionListPanel.tsx`
4. `src/features/positions/PositionCreateForm.tsx`
5. `src/features/positions/AssignMemberForm.tsx`
6. `src/features/roles/RoleManagementPage.tsx`
7. `src/features/roles/RoleCreateForm.tsx`
8. `src/features/roles/RoleEditForm.tsx`
9. `src/shared/api/client.ts`
10. `src/components/ui/form.tsx`

---

Вывод: проблема действительно не локальная. Это не только "грязный код", а отсутствие принудительного системного UI-контракта для create/edit/panel/state паттернов. Без фикса этого уровня фрагментация будет воспроизводиться.
