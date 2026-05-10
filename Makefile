ifneq (,$(wildcard .env))
include .env
export
endif

GOOSE ?= goose
COMPOSE ?= docker compose
GOOSE_MODULE ?= github.com/pressly/goose/v3/cmd/goose@v3.27.0
BACKEND_DIR ?= backend
FRONTEND_DIR ?= frontend
PHASE8_ACCEPTANCE_JSON ?= e2e\\test-results\\phase8-acceptance-report.json
PHASE9_ACCEPTANCE_JSON ?= e2e\\test-results\\phase9-acceptance-report.json
PHASE10_ACCEPTANCE_JSON ?= e2e\\test-results\\phase10-acceptance-report.json

.PHONY: help env setup tools dev dev-seed db-up db-down db-logs migrate-up migrate-down backend-run frontend-install frontend-run test test-e2e-phase4 phase4-acceptance test-e2e-phase7 phase7-acceptance test-e2e-phase8 phase8-acceptance-smoke phase8-acceptance phase9-acceptance-smoke phase9-acceptance phase10-acceptance-smoke phase10-acceptance build clean

help:
	@echo "Available targets:"
	@echo "  make env            - create .env and frontend/.env from examples if missing"
	@echo "  make setup          - create env files and install frontend deps"
	@echo "  make tools          - optional: install goose binary"
	@echo "  make dev            - one command for local dev (db + migrations + backend + frontend)"
	@echo "  make dev-seed       - seed dev users (admin/admin, test0/test0) and root division"
	@echo "  make db-up          - start PostgreSQL via docker compose"
	@echo "  make db-down        - stop docker compose services"
	@echo "  make migrate-up     - apply backend migrations"
	@echo "  make backend-run    - run Go backend on :8080"
	@echo "  make frontend-run   - run Vite frontend on :5173"
	@echo "  make test           - run backend tests"
	@echo "  make test-e2e-phase4 - run phase-4 access-control acceptance e2e"
	@echo "  make phase4-acceptance - seed DB fixtures and run phase-4 access-control acceptance e2e"
	@echo "  make test-e2e-phase7 - run phase-7 audit-and-archiving acceptance e2e"
	@echo "  make phase7-acceptance - seed DB fixtures and run phase-7 audit-and-archiving acceptance e2e"
	@echo "  make phase8-acceptance-smoke - run fast SC-7 smoke subset with strict Playwright flags"
	@echo "  make phase8-acceptance - run full strict SC-7 Playwright gate and enforce failed=0, skipped=0"
	@echo "  make phase9-acceptance-smoke - run focused Phase 9 smoke scenarios with strict Playwright flags"
	@echo "  make phase9-acceptance - run full strict Phase 9 Playwright gate and enforce failed=0, skipped=0"
	@echo "  make phase10-acceptance-smoke - run focused Phase 10 redesign scenarios with strict Playwright flags"
	@echo "  make phase10-acceptance - run full strict Phase 10 Playwright gate and enforce failed=0, skipped=0"
	@echo "  make build          - build frontend"

env:
	@if not exist .env copy .env.example .env >NUL
	@if not exist frontend\\.env copy frontend\\.env.example frontend\\.env >NUL

setup: env frontend-install

tools:
	cd backend && go install github.com/pressly/goose/v3/cmd/goose@latest

db-up:
	$(COMPOSE) up -d db
	@echo "PostgreSQL is running on localhost:5432"

db-down:
	$(COMPOSE) down

db-logs:
	$(COMPOSE) logs -f db

migrate-up:
	cd backend && go run $(GOOSE_MODULE) -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	cd backend && go run $(GOOSE_MODULE) -dir migrations postgres "$(DATABASE_URL)" down

backend-run:
	cd backend && go run ./cmd/server

frontend-install:
	cd frontend && npm install

frontend-run:
	cd frontend && npm run dev -- --host 0.0.0.0 --port 5173 --force

dev: env db-up migrate-up
	powershell -NoProfile -ExecutionPolicy Bypass -File .\\scripts\\dev.ps1

dev-seed:
	bash ./scripts/seed/dev-seed.sh

test:
	cd backend && go test ./...

test-e2e-phase4:
	cd frontend && npm run test:e2e -- access-control-enforcement.spec.ts

phase4-acceptance: dev-seed
	cd frontend && npm run test:e2e -- access-control-enforcement.spec.ts

test-e2e-phase7:
	cd e2e && npm test -- audit-archiving.scenario.spec.ts

phase7-acceptance: dev-seed
	cd e2e && npm test -- audit-archiving.scenario.spec.ts

test-e2e-phase8:
	cd e2e && npm test

phase8-acceptance-smoke:
	cd e2e && npx playwright test scenarios/positions-memberships-acl.scenario.spec.ts scenarios/role-management.scenario.spec.ts --forbid-only --fail-on-flaky-tests --reporter=line

phase8-acceptance:
	@if exist $(PHASE8_ACCEPTANCE_JSON) del /Q $(PHASE8_ACCEPTANCE_JSON)
	cd e2e && set PLAYWRIGHT_JSON_OUTPUT_NAME=phase8-acceptance-report.json&& set PLAYWRIGHT_JSON_OUTPUT_DIR=test-results&& npx playwright test --forbid-only --fail-on-flaky-tests --reporter=line,json
	powershell -NoProfile -Command "$$report = Get-Content -Raw '$(PHASE8_ACCEPTANCE_JSON)' | ConvertFrom-Json; $$failed = [int]$$report.stats.unexpected; $$skipped = [int]$$report.stats.skipped; if ($$failed -ne 0) { Write-Error ('phase8-acceptance failed counter is non-zero: failed=' + $$failed); exit 1 }; if ($$skipped -ne 0) { Write-Error ('phase8-acceptance skipped counter is non-zero: skipped=' + $$skipped); exit 1 }; Write-Host ('phase8-acceptance counters: failed=' + $$failed + ' skipped=' + $$skipped)"

phase9-acceptance-smoke:
	cd e2e && npx playwright test scenarios/divisions.scenario.spec.ts scenarios/profile.scenario.spec.ts scenarios/search.scenario.spec.ts --forbid-only --fail-on-flaky-tests --reporter=line

phase9-acceptance:
	@if exist $(PHASE9_ACCEPTANCE_JSON) del /Q $(PHASE9_ACCEPTANCE_JSON)
	cd e2e && set PLAYWRIGHT_JSON_OUTPUT_NAME=phase9-acceptance-report.json&& set PLAYWRIGHT_JSON_OUTPUT_DIR=test-results&& npx playwright test --forbid-only --fail-on-flaky-tests --reporter=line,json
	powershell -NoProfile -Command "$$report = Get-Content -Raw '$(PHASE9_ACCEPTANCE_JSON)' | ConvertFrom-Json; $$failed = [int]$$report.stats.unexpected; $$skipped = [int]$$report.stats.skipped; if ($$failed -ne 0) { Write-Error ('phase9-acceptance failed counter is non-zero: failed=' + $$failed); exit 1 }; if ($$skipped -ne 0) { Write-Error ('phase9-acceptance skipped counter is non-zero: skipped=' + $$skipped); exit 1 }; Write-Host ('phase9-acceptance counters: failed=' + $$failed + ' skipped=' + $$skipped)"

phase10-acceptance-smoke:
	cd e2e && npx playwright test scenarios/divisions-redesign.scenario.spec.ts scenarios/profile.scenario.spec.ts scenarios/search.scenario.spec.ts scenarios/role-management.scenario.spec.ts scenarios/a11y-redesign.scenario.spec.ts --forbid-only --fail-on-flaky-tests --reporter=line

phase10-acceptance:
	@if exist $(PHASE10_ACCEPTANCE_JSON) del /Q $(PHASE10_ACCEPTANCE_JSON)
	cd e2e && set PLAYWRIGHT_JSON_OUTPUT_NAME=phase10-acceptance-report.json&& set PLAYWRIGHT_JSON_OUTPUT_DIR=test-results&& npx playwright test --forbid-only --fail-on-flaky-tests --reporter=line,json
	powershell -NoProfile -Command "$$report = Get-Content -Raw '$(PHASE10_ACCEPTANCE_JSON)' | ConvertFrom-Json; $$failed = [int]$$report.stats.unexpected; $$skipped = [int]$$report.stats.skipped; if ($$failed -ne 0) { Write-Error ('phase10-acceptance failed counter is non-zero: failed=' + $$failed); exit 1 }; if ($$skipped -ne 0) { Write-Error ('phase10-acceptance skipped counter is non-zero: skipped=' + $$skipped); exit 1 }; Write-Host ('phase10-acceptance counters: failed=' + $$failed + ' skipped=' + $$skipped)"

build:
	cd frontend && npm run build

clean:
	@if exist frontend\\dist powershell -NoProfile -Command "Remove-Item -Recurse -Force 'frontend/dist'"
