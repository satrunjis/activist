# docs

Проектная документация.

- `tz.md` — техническое задание
- `transcript.txt` — расшифровка голосового обсуждения (Whisper)
- `current-state.md` — текущая карта backend API и frontend-возможностей
- `deploy.md` — инструкция production deploy
- `opd/` — материалы курса ОПД: raw-экспорты и workbook-файлы

## Быстрый локальный доступ

1. Поднять БД и применить migrations:
   - `make db-up`
   - `make migrate-up`
2. Засидить dev-данные:
   - `make dev-seed`
3. Запустить приложение:
   - `make dev`

Dev-учетки:
- `admin` / `admin`
- `test0` / `test0`

Примечание: скрипт `dev-seed` также создает root-подразделение `seed-root`, чтобы дерево отображалось сразу.
Также скрипт добавляет demo-ветки (Operations, Comms, Legal и дочерние узлы), чтобы сразу увидеть более глубокое дерево.

Dev ACL-флаг:
- `DEV_AUTH_ASSUME_ADMIN=true`
- `DEV_AUTH_ADMIN_LOGIN=admin`

При включенном флаге пользователь `admin` получает системные права в dev-среде.
