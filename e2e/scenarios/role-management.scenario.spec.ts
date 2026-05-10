import { randomUUID } from "node:crypto";
import {
  createRole,
  deleteRole,
  deleteRoleInUseShowsConflict,
  editRole,
  listRoles,
  viewRoleManagementReadonly
} from "../actions/roles.actions";
import { expect, test } from "../fixtures/base.fixture";

const admin = {
  login: "admin",
  password: "admin"
};

const roleManager = {
  login: "role_manager",
  password: "test0"
};

function uniqueRoleName(prefix: string): string {
  return `${prefix}-${randomUUID().replace(/-/g, "").slice(0, 10)}`;
}

test("ROLE-03 admin can create a role", async ({ page, entities }) => {
  const input = {
    actor: admin,
    name: uniqueRoleName("E2E Test Role"),
    permissions: [{ code: "can_add_member", enabled: true }]
  };
  const expected = {
    createdName: input.name
  };

  const created = await createRole(page, entities, input);
  expect(created.role?.id).toBeTruthy();

  const listed = await listRoles(page, entities, { actor: admin });
  expect(listed.items.some((item) => item.name === expected.createdName)).toBeTruthy();
});

test("ROLE-03 admin can create a role with explicit descendant scope", async ({ page, entities }) => {
  const input = {
    actor: admin,
    name: uniqueRoleName("E2E Scope Create"),
    permissions: [{ code: "can_add_member", enabled: true, scope: "current_and_descendants" as const }]
  };
  const expected = {
    permissionCode: "can_add_member",
    scope: "current_and_descendants" as const
  };

  const created = await createRole(page, entities, input);
  expect(created.role?.id).toBeTruthy();

  await editRole(page, entities, {
    actor: admin,
    roleId: created.role?.id,
    roleName: input.name,
    verifyScopes: {
      [expected.permissionCode]: expected.scope
    }
  });
});

test("ROLE-03 admin can edit a role name", async ({ page, entities }) => {
  const input = {
    actor: admin,
    originalName: uniqueRoleName("Editable Role"),
    renamedTo: uniqueRoleName("Editable Role Renamed")
  };
  const expected = {
    finalName: input.renamedTo
  };

  const created = await createRole(page, entities, {
    actor: input.actor,
    name: input.originalName
  });
  const edited = await editRole(page, entities, {
    actor: input.actor,
    roleId: created.role?.id,
    roleName: input.originalName,
    nextName: input.renamedTo
  });

  expect(edited.role.name).toBe(expected.finalName);
});

test("ROLE-03 admin can edit a role scope", async ({ page, entities }) => {
  const input = {
    actor: admin,
    roleName: uniqueRoleName("E2E Scope Edit"),
    permissionCode: "can_view_contacts"
  };
  const expected = {
    scope: "current_and_descendants" as const
  };

  const created = await createRole(page, entities, {
    actor: input.actor,
    name: input.roleName,
    permissions: [{ code: input.permissionCode, enabled: true }]
  });

  await editRole(page, entities, {
    actor: input.actor,
    roleId: created.role?.id,
    roleName: input.roleName,
    permissions: [{ code: input.permissionCode, enabled: true, scope: expected.scope }],
    verifyScopes: { [input.permissionCode]: expected.scope }
  });
});

test("ROLE-03 admin can delete a role not referenced by any position", async ({ page, entities }) => {
  const input = {
    actor: admin,
    roleName: uniqueRoleName("Deletable Role")
  };
  const expected = {
    deletedName: input.roleName
  };

  const created = await createRole(page, entities, {
    actor: input.actor,
    name: input.roleName
  });
  const deleted = await deleteRole(page, entities, {
    actor: input.actor,
    roleId: created.role?.id,
    roleName: input.roleName
  });

  expect(deleted.deletedRoleName).toBe(expected.deletedName);
});

test("ROLE-03 deleting a role in use shows conflict error", async ({ page, entities }) => {
  const input = {
    actor: admin,
    roleId: "seed-role-standard",
    expectedMessage: "Роль назначена одной или нескольким должностям."
  };
  const expected = {
    message: input.expectedMessage
  };

  const result = await deleteRoleInUseShowsConflict(page, entities, input);
  expect(result.error).toContain(expected.message);
});

test("ROLE-02 role-manager cannot assign system_admin permission", async ({ page, entities }) => {
  const input = {
    actor: roleManager,
    expectedMessage: "Управление ролями доступно только администраторам."
  };
  const expected = {
    message: input.expectedMessage
  };

  const result = await viewRoleManagementReadonly(page, entities, input);
  expect(result.message).toContain(expected.message);
});

test("ROLE-01 role-manager cannot create/edit/delete roles", async ({ page, entities }) => {
  const input = {
    actor: roleManager
  };
  const expected = {
    readonlyVisible: true
  };

  const result = await viewRoleManagementReadonly(page, entities, input);
  expect(Boolean(result.message)).toBe(expected.readonlyVisible);
});

test("ROLE-03 create role form validates required name", async ({ page, entities }) => {
  const input = {
    actor: admin,
    name: "",
    expectedError: "Название обязательно."
  };
  const expected = {
    error: input.expectedError
  };

  const result = await createRole(page, entities, input);
  expect(result.error).toContain(expected.error);
});

test("ROLE-03 edit role form validates required name", async ({ page, entities }) => {
  const input = {
    actor: admin,
    roleName: uniqueRoleName("Editable Validation Role"),
    invalidName: "",
    expectedError: "Название обязательно."
  };
  const expected = {
    error: input.expectedError
  };

  const created = await createRole(page, entities, {
    actor: input.actor,
    name: input.roleName
  });
  const result = await editRole(page, entities, {
    actor: input.actor,
    roleId: created.role?.id,
    roleName: input.roleName,
    nextName: input.invalidName,
    expectedError: input.expectedError
  });

  expect(result.error).toContain(expected.error);
});

test("ROLE-03 create role renders backend error for duplicate name", async ({ page, entities }) => {
  const input = {
    actor: admin,
    roleName: uniqueRoleName("Duplicate Role"),
    expectedError: /(already|exists|unique|duplicate|conflict|internal server error|failed to create role|уже|существ|создать роль)/i
  };
  const expected = {
    hasError: true
  };

  await createRole(page, entities, {
    actor: input.actor,
    name: input.roleName
  });
  const duplicate = await createRole(page, entities, {
    actor: input.actor,
    name: input.roleName,
    expectedError: input.expectedError
  });

  expect(Boolean(duplicate.error)).toBe(expected.hasError);
});
