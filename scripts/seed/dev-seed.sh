#!/usr/bin/env sh
set -eu

SEED_MODE=${SEED_MODE:-local}
SEED_PG_CONTAINER=${SEED_PG_CONTAINER:-activist-base-pg}
SEED_PG_USER=${SEED_PG_USER:-postgres}
SEED_PG_DATABASE=${SEED_PG_DATABASE:-activist_base}

echo "SEED_MODE=$SEED_MODE"

SQL=$(cat <<'SQL'
INSERT INTO users (id, login, password_hash, first_name, gradebook_number, group_number, institute, birth_date)
VALUES
  ('seed-admin', 'admin', $ph_admin$$argon2id$v=19$m=65536,t=3,p=2$AjrRLbbUcdKfNdY3qu7mDg$FtWDu4fNnu7/K2Y+gR/bkh+q8pKExAT69xCooT/aCOU$ph_admin$, 'Admin', 'ADM-0001', 'DEV', 'Dev Institute', '2000-01-01')
ON CONFLICT (login) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    first_name = EXCLUDED.first_name;

INSERT INTO users (id, login, password_hash, first_name, gradebook_number, group_number, institute, birth_date)
VALUES
  ('seed-test0', 'test0', $ph_test0$$argon2id$v=19$m=65536,t=3,p=2$J6q2Z8PgZVNQXii6UTX+dg$vDWeU75WiRTnMiDHvL3K98rvT6kZC4c1bOjBFhWVCYA$ph_test0$, 'Test0', 'TST-0001', 'DEV', 'Dev Institute', '2000-01-01')
ON CONFLICT (login) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    first_name = EXCLUDED.first_name;

INSERT INTO users (id, login, password_hash, first_name, gradebook_number, group_number, institute, birth_date)
VALUES
  ('seed-target', 'seed_target', $ph_target$$argon2id$v=19$m=65536,t=3,p=2$J6q2Z8PgZVNQXii6UTX+dg$vDWeU75WiRTnMiDHvL3K98rvT6kZC4c1bOjBFhWVCYA$ph_target$, 'Target', 'TST-0002', 'DEV', 'Dev Institute', '2000-01-01')
ON CONFLICT (login) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    first_name = EXCLUDED.first_name;

INSERT INTO users (id, login, password_hash, first_name, gradebook_number, group_number, institute, birth_date)
VALUES
  ('seed-role-manager', 'role_manager', $ph_target$$argon2id$v=19$m=65536,t=3,p=2$J6q2Z8PgZVNQXii6UTX+dg$vDWeU75WiRTnMiDHvL3K98rvT6kZC4c1bOjBFhWVCYA$ph_target$, 'RoleManager', 'TST-0003', 'DEV', 'Dev Institute', '2000-01-01')
ON CONFLICT (login) DO UPDATE
SET password_hash = EXCLUDED.password_hash,
    first_name = EXCLUDED.first_name;

INSERT INTO divisions (id, parent_id, short_name, full_name, description, regulation_url, media_links, is_archived)
VALUES ('seed-root', NULL, 'Root', 'Seed Root Division', 'Development root division', '', '[]'::jsonb, false)
ON CONFLICT (id) DO NOTHING;

INSERT INTO divisions (id, parent_id, short_name, full_name, description, regulation_url, media_links, is_archived)
VALUES
  ('seed-ops', 'seed-root', 'Operations', 'Operations Division', 'Operations branch', '', '[]'::jsonb, false),
  ('seed-comms', 'seed-root', 'Comms', 'Communications Division', 'Communications branch', '', '[]'::jsonb, false),
  ('seed-legal', 'seed-root', 'Legal', 'Legal Division', 'Legal branch', '', '[]'::jsonb, false),
  ('seed-ops-field', 'seed-ops', 'Field Ops', 'Field Operations', 'Field operations team', '', '[]'::jsonb, false),
  ('seed-ops-analytics', 'seed-ops', 'Analytics', 'Operations Analytics', 'Analytics team', '', '[]'::jsonb, false),
  ('seed-comms-social', 'seed-comms', 'Social', 'Social Media Team', 'Social team', '', '[]'::jsonb, false),
  ('seed-comms-content', 'seed-comms', 'Content', 'Content Team', 'Content team', '', '[]'::jsonb, false),
  ('seed-legal-policy', 'seed-legal', 'Policy', 'Policy Team', 'Policy team', '', '[]'::jsonb, false)
ON CONFLICT (id) DO NOTHING;

INSERT INTO roles (id, name, kind, permissions, created_at, updated_at)
VALUES
  ('seed-role-standard', 'Seed Standard', 'standard', decode('010802', 'hex'), now(), now()),
  ('seed-role-manager', 'Seed Manager (Current Division)', 'standard', decode('0103020502', 'hex'), now(), now()),
  ('seed-role-can-manage-roles', 'Seed CanManageRoles', 'standard', decode('010A03', 'hex'), now(), now()),
  ('seed-role-system-admin', 'Seed System Admin', 'standard', decode('01FF03', 'hex'), now(), now()),
  ('seed-role-leader', 'Seed Leader', 'leader', ''::bytea, now(), now()),
  ('seed-role-deputy', 'Seed Deputy', 'deputy', ''::bytea, now(), now())
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name,
    kind = EXCLUDED.kind,
    permissions = EXCLUDED.permissions,
    updated_at = now();

INSERT INTO positions (id, title, role_id, division_id, max_count, is_archived)
VALUES
  ('seed-pos-manager', 'Ops Coordinator', 'seed-role-manager', 'seed-ops', NULL, false),
  ('seed-pos-system-admin', 'Root System Admin Seat', 'seed-role-system-admin', 'seed-root', NULL, false),
  ('seed-pos-role-manager', 'Role Manager Seat', 'seed-role-can-manage-roles', 'seed-root', NULL, false),
  ('seed-pos-leader', 'Ops Leader Seat', 'seed-role-leader', 'seed-ops', NULL, false),
  ('seed-pos-field', 'Field Coordinator', 'seed-role-standard', 'seed-ops-field', NULL, false)
ON CONFLICT (id) DO UPDATE
SET title = EXCLUDED.title,
    role_id = EXCLUDED.role_id,
    division_id = EXCLUDED.division_id,
    max_count = EXCLUDED.max_count,
    is_archived = EXCLUDED.is_archived;

DELETE FROM memberships
WHERE user_id IN (
  SELECT id
  FROM users
  WHERE login IN ('admin', 'test0', 'seed_target', 'role_manager')
);

INSERT INTO memberships (user_id, position_id)
VALUES
  ((SELECT id FROM users WHERE login = 'admin' LIMIT 1), 'seed-pos-system-admin'),
  ((SELECT id FROM users WHERE login = 'test0' LIMIT 1), 'seed-pos-manager'),
  ((SELECT id FROM users WHERE login = 'role_manager' LIMIT 1), 'seed-pos-role-manager')
ON CONFLICT (user_id, position_id) DO NOTHING;
SQL
)

case "$SEED_MODE" in
  local|"")
    docker exec -i "$SEED_PG_CONTAINER" psql -U "$SEED_PG_USER" -d "$SEED_PG_DATABASE" -v ON_ERROR_STOP=1 -c "$SQL"
    ;;
  ci)
    if [ -z "${DATABASE_URL:-}" ]; then
      echo "[dev-seed] DATABASE_URL is required when SEED_MODE=ci" >&2
      exit 1
    fi
    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -c "$SQL"
    ;;
  *)
    echo "[dev-seed] unsupported SEED_MODE: $SEED_MODE" >&2
    exit 1
    ;;
esac

echo "[dev-seed] seeded users:"
echo "  login=admin password=admin"
echo "  login=test0 password=test0"
echo "  login=seed_target password=test0"
echo "[dev-seed] ensured root division id=seed-root"
echo "[dev-seed] ensured additional demo divisions for tree preview"
echo "[dev-seed] ensured role fixtures via direct DB write (standard/manager/leader/deputy)"
echo "[dev-seed] ensured phase-4 acceptance fixtures (seed-pos-manager, seed-pos-leader, seed-pos-field)"
echo "[dev-seed] assigned admin to system-admin role at root division"
echo "[dev-seed] ensured seed-role-manager user account (login=role_manager password=test0) - role/position/membership prepared in DB seed"
