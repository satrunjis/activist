# activist

activist — web-платформа для ведения реестра участников организации: хранит профили пользователей, иерархию подразделений, должности, memberships, role-based permissions, поиск пользователей и append-only event log для изменений в организации.

## Технологический стек

| Область | Технологии |
| --- | --- |
| Backend | Go 1.25.0, chi v5.2.3, pgx v5.7.6, sqlc v1.29.0, goose v3 |
| Frontend | React 19, TypeScript 5.9.3, Vite 7, Tailwind CSS 4 |
| E2E | Playwright 1.59.1 |
| Infra | Docker, Docker Compose, PostgreSQL 16 |

## Структура репозитория

| Директория | Содержимое |
| --- | --- |
| `backend/` | Go API server, SQL migrations, generated query layer и backend Dockerfiles. |
| `frontend/` | React/Vite frontend-приложение, UI components, feature modules и frontend Dockerfile. |
| `e2e/` | Playwright scenarios, fixtures и end-to-end test helpers. |
| `scripts/` | Скрипты local development и seed scripts, используемые Make targets. |
| `docs/` | Product requirements и документация текущей реализации. |
| `deploy/` | Production Docker Compose configuration. |
| `perf/` | k6 performance scenarios и reporting helpers. |
| `.github/` | Конфигурация GitHub repository. |

## Требования

| Инструмент | Требуемая версия |
| --- | --- |
| Go | 1.25.0 |
| Node.js | `^20.19.0 || >=22.12.0` |
| Docker | Версия не зафиксирована; требуется Docker Compose v2. |
| make | Версия не зафиксирована; требуется GNU Make-compatible `make`. |

## Быстрый старт

Создайте локальные environment-файлы:

```sh
make env
```

Заполните обязательные значения в `.env` и `frontend/.env`. Для стандартной локальной базы данных из `docker-compose.yml` `DATABASE_URL` должен указывать на PostgreSQL на `localhost:5432`.

Установите frontend dependencies:

```sh
make frontend-install
```

Запустите PostgreSQL:

```sh
make db-up
```

Примените backend migrations:

```sh
make migrate-up
```

Запустите локальные backend и frontend:

```sh
make dev
```

Backend по умолчанию слушает `:8080`. Frontend dev server работает на `http://localhost:5173`.

## Environment Variables

| Переменная | Обязательность | Описание | Пример значения |
| --- | --- | --- | --- |
| `DATABASE_URL` | Обязательная | PostgreSQL connection string для API server и migration runner. | `postgres://USER:PASSWORD@localhost:5432/activist_base?sslmode=disable` |
| `APP_ENV` | Опциональная | Метка application environment. | `development` |
| `HTTP_ADDR` | Обязательная | Backend HTTP listen address. | `:8080` |
| `PUBLIC_BASE_URL` | Обязательная | Public base URL, настроенный для backend. | `http://localhost:8080` |
| `SESSION_COOKIE_NAME` | Обязательная | Имя authentication session cookie. | `__Host-session` |
| `SESSION_IDLE_TTL` | Обязательная | Idle session timeout как Go duration. | `30m` |
| `SESSION_ABSOLUTE_TTL` | Обязательная | Максимальное время жизни session как Go duration. | `8h` |
| `LOG_LEVEL` | Обязательная | Structured log level. | `info` |
| `SEED_SUPERUSER_LOGIN` | Обязательная | Login, используемый seed tooling для начального superuser. | `superadmin` |
| `SEED_SUPERUSER_PASSWORD` | Обязательная | Password, используемый seed tooling для начального superuser. | `change-me` |
| `DEV_AUTH_ASSUME_ADMIN` | Опциональная | Включает development-only admin assumption в auth middleware. | `false` |
| `DEV_AUTH_ADMIN_LOGIN` | Опциональная | Login, используемый при включенном development admin assumption. | `admin` |
| `E2E_FRONTEND_URL` | Опциональная | Frontend URL, используемый Playwright tests. | `http://localhost:5173` |
| `E2E_API_BASE_URL` | Опциональная | API base URL, используемый Playwright tests. | `http://localhost:8080/api/v1` |
| `E2E_DATABASE_URL` | Опциональная | Database URL, используемый Playwright fixtures. | `postgres://USER:PASSWORD@localhost:5432/activist_base?sslmode=disable` |
| `DEMO_API_BASE_URL` | Опциональная | API base URL, используемый demo seed scripts. | `http://localhost:8080/api/v1` |
| `DEMO_PG_CONTAINER` | Опциональная | Имя PostgreSQL container, используемое demo seed scripts. | `activist-base-pg` |
| `DEMO_PG_USER` | Опциональная | PostgreSQL user, используемый demo seed scripts. | `postgres` |
| `DEMO_PG_DATABASE` | Опциональная | PostgreSQL database, используемая demo seed scripts. | `activist_base` |
| `PERF_BASE_URL` | Опциональная | Base URL, используемый performance scenarios. | `http://127.0.0.1:8080` |
| `PERF_BACKEND_PORT` | Опциональная | Backend port, используемый для performance resource sampling. | `8080` |
| `PERF_BACKEND_CONTAINER` | Опциональная | Имя backend container, используемое для performance resource sampling. | `activist-backend` |
| `PERF_LOGIN` | Опциональная | Login, используемый performance scenarios с authentication. | `admin` |
| `PERF_PASSWORD` | Опциональная | Password, используемый performance scenarios с authentication. | `change-me` |

## Запуск тестов

Запустите backend unit tests:

```sh
make test
```

Эквивалентная backend-команда:

```sh
cd backend && go test ./...
```

Запустите frontend unit tests:

```sh
cd frontend && npm run test:unit:run
```

Запустите end-to-end tests:

```sh
cd e2e && npm test
```

## Deployment

Container images собираются через GitHub Actions и публикуются в GHCR. Production запускается с `deploy/docker-compose.prod.yml`.

## Участие в разработке

Issues и pull requests приветствуются. Держите изменения сфокусированными, добавляйте tests для изменений поведения и обновляйте документацию, когда меняется setup или runtime behavior. Не коммитьте локальные environment-файлы, generated secrets или machine-specific configuration.
