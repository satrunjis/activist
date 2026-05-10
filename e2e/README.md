# E2E Scaffold (Phase 2)

Last updated: 2026-04-18  
Owner area: `activist_base/e2e`

## Purpose
This directory contains the Playwright E2E scaffold for API-driven acceptance tests.

Principles in this scaffold:
- Runtime setup only: each test creates all required data through backend API helpers.
- No shared baseline accounts in tests.
- Per-test teardown is unconditional and runs even when a test fails.
- Teardown order is FK-safe and fixed.

Out of scope in this phase:
- Positions archive endpoint with no UI coverage.

## Directory Layout
- `playwright.config.ts` - Playwright config.
- `fixtures/db.ts` - direct SQL client helpers for teardown.
- `fixtures/base.fixture.ts` - base test fixture with hidden setup/teardown.
- `actions/*.actions.ts` - reusable action helpers consumed by scenario suites.
- `scenarios/*.spec.ts` - declarative E2E scenarios.
- `support/seed.ts` - API setup helpers.
- `support/teardown.ts` - ID tracker + FK-safe cleanup.
- `support/test-data.ts` - data builders.
- `support/env.ts` - environment settings.

## Quick Start
1. Start DB from `activist_base`:
   `docker compose up -d db`
2. Start backend:
   from `activist-backend`, run `go run ./cmd/server`
3. Install dependencies:
   from `activist_base/e2e`, run `npm install`
4. Run E2E:
   `npm test`

## Environment
- `E2E_API_BASE_URL` (default: `http://127.0.0.1:8080/api/v1`)
- `E2E_DATABASE_URL` (default: `postgresql://postgres:postgres@127.0.0.1:5432/activist_base`)

## Known-Good Pipeline (Local)
1. Start DB from `activist_base`: `docker compose up -d db`
2. Start backend with local env and `DEV_AUTH_ASSUME_ADMIN=false`
3. From `activist_base/e2e`, run `npm test`

## Phase 7 Acceptance (D-10..D-14)
Run from `activist_base` root:
1. `make db-up`
2. Start backend in a separate terminal: `make backend-run`
3. Seed deterministic fixtures (`admin/admin`, `test0/test0`, `Seed Standard` role): `make dev-seed`
4. Execute acceptance gate: `make phase7-acceptance`

Direct scenario rerun (without reseeding): `make test-e2e-phase7`

## SC-7 Action Coverage Linkage Check
Use this low-latency check before smoke/full runs to confirm scenarios still import the shared action layer.

Run from `activist_base` root:
1. `cd e2e && rg -n 'from "../actions/' scenarios`

Expected result:
- All scenario suites print at least one `from "../actions/...` import line.
- Every action file under `e2e/actions/` remains represented in scenario imports.

## Fast Smoke Subset (Pre-Gate)
Run this targeted subset before strict full-suite gating.

Run from `activist_base` root:
1. `cd e2e && npx playwright test scenarios/positions-memberships-acl.scenario.spec.ts scenarios/role-management.scenario.spec.ts --forbid-only --fail-on-flaky-tests --reporter=line`

This matches the strict smoke gate in `make phase8-acceptance-smoke`.

## Known-Good Pipeline (General Playwright)
1. Ensure backend is reachable from `E2E_API_BASE_URL`.
2. Ensure DB is reachable from `E2E_DATABASE_URL`.
3. Run:
   `cd activist_base/e2e && npm test`

## Stable Baselines (Current)
- Teardown uses direct SQL deletes in strict FK-safe order:
  memberships -> positions -> divisions -> roles -> users
- Test setup uses API helpers from `support/seed.ts`.
- Scenario files keep declarative style by defining input and expected result at test top.

## Common Failure Patterns

### 1) Backend not reachable
- Symptom: requests fail with connection errors.
- Fix: verify backend is listening on the `E2E_API_BASE_URL` host/port.

### 2) DB not reachable for teardown
- Symptom: test passes but teardown throws DB connection errors.
- Fix: verify Postgres container and `E2E_DATABASE_URL`.

### 3) Duplicate unique fields during setup
- Symptom: `auth` or validation conflict for login/gradebook data.
- Fix: use generated unique test data from `support/test-data.ts`.

## Debugging
- Headed mode: `npm run test:headed`
- Inspector mode: `npm run test:debug`
- UI mode: `npm run test:ui`

## Update Protocol (Mandatory)
When any E2E startup/runtime issue appears:
1. Add a new failure item under `Common Failure Patterns` or extend an existing one.
2. Record exact reproduce and fix commands.
3. Update `Last updated` date.
4. If fix changes code/config, add file paths and commit hash.
