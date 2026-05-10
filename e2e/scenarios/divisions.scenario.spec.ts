import { expect, type Page, type Route } from "@playwright/test";
import {
  createDivision,
  editDivision,
  reparentDivision,
  viewDivisionTree,
  type DivisionMediaLinkInput
} from "../actions/divisions.actions";
import { test } from "../fixtures/base.fixture";

type DivisionNode = {
  id: string;
  parent_id?: string;
  short_name: string;
  full_name?: string;
  description?: string;
  regulation_url?: string;
  media_links?: Array<{ platform: string; value: string }>;
  is_archived: boolean;
  has_children: boolean;
  children_count: number;
  children?: DivisionNode[];
};

type DivisionCreatePayload = {
  parent_id?: string;
  short_name: string;
  full_name: string;
  description: string;
  regulation_url: string;
  media_links?: Array<{ platform: string; value: string }>;
};

type DivisionPatchPayload = {
  parent_id?: string;
  short_name?: string;
  full_name?: string;
  description?: string;
  regulation_url?: string;
  media_links?: Array<{ platform: string; value: string }>;
};

type InstallMockOptions = {
  forbidCreate?: boolean;
  forbidEdit?: boolean;
};

function cloneTree<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}

function findNode(root: DivisionNode, id: string): DivisionNode | null {
  if (root.id === id) {
    return root;
  }
  for (const child of root.children ?? []) {
    const found = findNode(child, id);
    if (found) {
      return found;
    }
  }
  return null;
}

function findParent(root: DivisionNode, childId: string): DivisionNode | null {
  for (const child of root.children ?? []) {
    if (child.id === childId) {
      return root;
    }
    const found = findParent(child, childId);
    if (found) {
      return found;
    }
  }
  return null;
}

function recount(node: DivisionNode): void {
  const children = node.children ?? [];
  node.children_count = children.length;
  node.has_children = children.length > 0;
  for (const child of children) {
    recount(child);
  }
}

function normalizeLinks(links: DivisionMediaLinkInput[]): DivisionMediaLinkInput[] {
  return links
    .map((item) => ({
      platform: item.platform.trim(),
      value: item.value.trim()
    }))
    .filter((item) => item.platform.length > 0 && item.value.length > 0);
}

async function installDivisionTreeApiMocks(page: Page, options: InstallMockOptions = {}) {
  const tree: DivisionNode = {
    id: "seed-root",
    short_name: "Central Committee",
    full_name: "Central Committee",
    description: "Main division",
    is_archived: false,
    has_children: true,
    children_count: 1,
    children: [
      {
        id: "ops",
        parent_id: "seed-root",
        short_name: "Operations",
        full_name: "Operations Division",
        description: "Operations",
        is_archived: false,
        has_children: true,
        children_count: 1,
        children: [
          {
            id: "ops-hr",
            parent_id: "ops",
            short_name: "HR",
            full_name: "Operations HR",
            description: "People operations",
            is_archived: false,
            has_children: false,
            children_count: 0,
            children: []
          }
        ]
      }
    ]
  };

  const createdPayloads: DivisionCreatePayload[] = [];
  const patchedPayloads: DivisionPatchPayload[] = [];

  await page.route("**/api/v1/auth/session", async (route: Route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        user: {
          id: "u-1",
          login: "admin",
          first_name: "E2E",
          memberships: [
            {
              membership_id: "m-1",
              division_id: "seed-root",
              division_name: "Central Committee",
              position_id: "p-1",
              position_title: "Chair",
              role_name: "system_admin"
            }
          ]
        },
        permissions: [
          "can_create_subdivision",
          "can_edit_division",
          "can_archive_division",
          "can_manage_positions",
          "can_add_member",
          "can_remove_member",
          "can_view_audit_log"
        ],
        session: {
          expires_at: "2030-01-01T00:00:00Z",
          idle_expires_at: "2030-01-01T00:00:00Z"
        },
        csrf_token: "csrf-e2e"
      })
    });
  });

  await page.route("**/api/v1/users/u-1", async (route: Route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({
        id: "u-1",
        first_name: "E2E",
        last_name: "User",
        memberships: []
      })
    });
  });

  await page.route("**/api/v1/users/u-1/memberships", async (route: Route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ items: [] })
    });
  });

  await page.route("**/api/v1/positions**", async (route: Route) => {
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ items: [] })
    });
  });

  await page.route("**/api/v1/divisions/tree**", async (route: Route) => {
    recount(tree);
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(cloneTree(tree))
    });
  });

  // Mock for lazy-load child-list endpoint: GET /api/v1/divisions?parent_id=
  await page.route("**/api/v1/divisions?**", async (route: Route) => {
    if (route.request().method() !== "GET") {
      await route.fallback();
      return;
    }
    const url = new URL(route.request().url());
    const parentId = url.searchParams.get("parent_id");
    recount(tree);
    const source = !parentId || parentId === "seed-root"
      ? (tree.children ?? [])
      : (findNode(tree, parentId)?.children ?? []);
    const items = source.map((node) => ({
      id: node.id,
      parent_id: node.parent_id,
      short_name: node.short_name,
      full_name: node.full_name,
      description: node.description,
      regulation_url: node.regulation_url,
      media_links: node.media_links,
      is_archived: node.is_archived,
      has_children: node.has_children,
      children_count: node.children_count
    }));
    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify({ items })
    });
  });

  await page.route("**/api/v1/divisions", async (route: Route) => {
    if (route.request().method() !== "POST") {
      await route.fallback();
      return;
    }

    if (options.forbidCreate) {
      await route.fulfill({
        status: 403,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            code: "access.scope",
            message: "forbidden"
          }
        })
      });
      return;
    }

    const payload = route.request().postDataJSON() as DivisionCreatePayload;
    createdPayloads.push(payload);

    const parent = payload.parent_id ? findNode(tree, payload.parent_id) : tree;
    if (!parent) {
      await route.fulfill({
        status: 404,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            code: "not_found.division",
            message: "division not found"
          }
        })
      });
      return;
    }

    const slug = payload.short_name.trim().toLowerCase().replace(/[^a-z0-9]+/g, "-");
    const id = slug || `node-${Date.now()}`;
    const nextNode: DivisionNode = {
      id,
      parent_id: parent.id,
      short_name: payload.short_name,
      full_name: payload.full_name,
      description: payload.description,
      regulation_url: payload.regulation_url,
      media_links: payload.media_links ?? [],
      is_archived: false,
      has_children: false,
      children_count: 0,
      children: []
    };

    parent.children = parent.children ?? [];
    parent.children.push(nextNode);
    recount(tree);

    await route.fulfill({
      status: 201,
      contentType: "application/json",
      body: JSON.stringify(nextNode)
    });
  });

  await page.route("**/api/v1/divisions/*", async (route: Route) => {
    if (route.request().method() !== "PATCH") {
      await route.fallback();
      return;
    }

    if (options.forbidEdit) {
      await route.fulfill({
        status: 403,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            code: "access.scope",
            message: "forbidden"
          }
        })
      });
      return;
    }

    const url = new URL(route.request().url());
    const pathParts = url.pathname.split("/");
    const divisionId = decodeURIComponent(pathParts[pathParts.length - 1] ?? "");
    const payload = route.request().postDataJSON() as DivisionPatchPayload;
    patchedPayloads.push(payload);

    const node = findNode(tree, divisionId);
    if (!node) {
      await route.fulfill({
        status: 404,
        contentType: "application/json",
        body: JSON.stringify({
          error: {
            code: "not_found.division",
            message: "division not found"
          }
        })
      });
      return;
    }

    if (payload.parent_id && payload.parent_id !== node.parent_id) {
      const currentParent = findParent(tree, node.id);
      const nextParent = findNode(tree, payload.parent_id);
      if (!nextParent) {
        await route.fulfill({
          status: 404,
          contentType: "application/json",
          body: JSON.stringify({
            error: {
              code: "not_found.division",
              message: "new parent not found"
            }
          })
        });
        return;
      }

      if (currentParent) {
        currentParent.children = (currentParent.children ?? []).filter((child) => child.id !== node.id);
      }
      nextParent.children = nextParent.children ?? [];
      nextParent.children.push(node);
      node.parent_id = nextParent.id;
    }

    if (payload.short_name !== undefined) {
      node.short_name = payload.short_name;
    }
    if (payload.full_name !== undefined) {
      node.full_name = payload.full_name;
    }
    if (payload.description !== undefined) {
      node.description = payload.description;
    }
    if (payload.regulation_url !== undefined) {
      node.regulation_url = payload.regulation_url;
    }
    if (payload.media_links !== undefined) {
      node.media_links = payload.media_links;
    }
    recount(tree);

    await route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(node)
    });
  });

  return {
    getNode: (id: string) => findNode(tree, id),
    createdPayloads,
    patchedPayloads
  };
}

test("division-tree view renders nested structure", async ({ page, entities }) => {
  const input = {
    path: "/",
    visibleNodeIds: ["ops", "ops-hr"]
  };

  await installDivisionTreeApiMocks(page);
  await viewDivisionTree(page, entities, input);

  await expect(page.getByTestId("division-node-ops-hr")).toBeVisible();
});

test("create-division flow adds a child to the tree", async ({ page, entities }) => {
  const input = {
    parentId: "seed-root",
    shortName: "Legal",
    fullName: "Legal Division",
    description: "Legal support",
    regulationUrl: "https://example.org/legal"
  };
  const expected = {
    createdId: "legal",
    selectedShortName: "Legal"
  };

  await installDivisionTreeApiMocks(page);
  await viewDivisionTree(page, entities, { path: "/", visibleNodeIds: ["ops"] });
  const created = await createDivision(page, entities, input);

  expect(created).not.toBeNull();
  if (!created) {
    return;
  }
  expect(created.id).toBe(expected.createdId);
  expect(created.shortName).toBe(expected.selectedShortName);
  await expect(page.getByTestId(`division-node-${expected.createdId}`)).toBeVisible();
});

test("edit-division flow updates selected division details", async ({ page, entities }) => {
  const input = {
    targetDivisionId: "ops",
    shortName: "Operations HQ",
    fullName: "Operations Headquarters",
    description: "Updated description"
  };
  const expected = {
    nodeId: "ops",
    shortName: "Operations HQ"
  };

  await installDivisionTreeApiMocks(page);
  await viewDivisionTree(page, entities, { path: "/", visibleNodeIds: ["ops"] });
  const updated = await editDivision(page, entities, input);

  expect(updated).not.toBeNull();
  if (!updated) {
    return;
  }
  expect(updated.id).toBe(expected.nodeId);
  expect(updated.shortName).toBe(expected.shortName);
});

test("reparent division moves under a new parent", async ({ page, entities }) => {
  const input = {
    create: {
      parentId: "seed-root",
      shortName: "Legal",
      fullName: "Legal Division",
      description: "Legal support"
    },
    reparentTo: "ops"
  };
  const expected = {
    newParentId: "ops"
  };

  const mock = await installDivisionTreeApiMocks(page);
  await viewDivisionTree(page, entities, { path: "/", visibleNodeIds: ["ops"] });
  const created = await createDivision(page, entities, input.create);
  expect(created).not.toBeNull();
  if (!created) {
    return;
  }

  const moved = await reparentDivision(page, entities, {
    targetDivisionId: created.id,
    newParentId: input.reparentTo
  });
  const movedNode = mock.getNode(moved.id);
  expect(movedNode?.parent_id).toBe(expected.newParentId);
});

test("media links are sent on create and edit division flows", async ({ page, entities }) => {
  const input = {
    create: {
      parentId: "seed-root",
      shortName: "Media Team",
      fullName: "Media Team",
      mediaLinks: [
        { platform: " telegram ", value: " https://t.me/activist " },
        { platform: "telegram", value: "https://t.me/activist" },
        { platform: "youtube", value: "https://youtube.com/@activist" }
      ]
    },
    edit: {
      mediaLinks: [
        { platform: "vk", value: "https://vk.com/activist" }
      ]
    }
  };
  const expected = {
    createdMediaLinks: normalizeLinks(input.create.mediaLinks),
    editedMediaLinks: normalizeLinks(input.edit.mediaLinks)
  };

  const mock = await installDivisionTreeApiMocks(page);
  await viewDivisionTree(page, entities, { path: "/", visibleNodeIds: ["ops"] });
  const created = await createDivision(page, entities, input.create);

  expect(created).not.toBeNull();
  if (!created) {
    return;
  }

  await editDivision(page, entities, {
    targetDivisionId: created.id,
    mediaLinks: input.edit.mediaLinks
  });

  const lastCreated = mock.createdPayloads[mock.createdPayloads.length - 1];
  const lastPatched = mock.patchedPayloads[mock.patchedPayloads.length - 1];
  expect(lastCreated?.media_links).toEqual(expected.createdMediaLinks);
  expect(lastPatched?.media_links).toEqual(expected.editedMediaLinks);
});

test("forbidden create locks create and edit controls", async ({ page, entities }) => {
  const input = {
    parentId: "seed-root",
    shortName: "Blocked Division",
    fullName: "Blocked Division",
    description: "Should fail",
    expectForbidden: true
  };
  const expected = {
    createdNodeTestId: "division-node-blocked-division"
  };

  await installDivisionTreeApiMocks(page, { forbidCreate: true });
  await viewDivisionTree(page, entities, { path: "/", visibleNodeIds: ["ops"] });
  const result = await createDivision(page, entities, input);

  expect(result).toBeNull();
  await expect(page.getByTestId(expected.createdNodeTestId)).toHaveCount(0);
});
