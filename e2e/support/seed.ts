import type { APIRequestContext, APIResponse } from "@playwright/test";
import { env } from "./env";
import type { RegisterUserInput } from "./test-data";
import { trackCreated, type CreatedEntities } from "./teardown";

type LoginInput = {
  login: string;
  password: string;
};

type ApiErrorEnvelope = {
  error?: {
    code?: string;
    message?: string;
  };
};

type AuthUserResult = {
  user: {
    id: string;
    login: string;
    first_name?: string;
    gradebook_number?: string;
  };
};

type AuthSessionResult = {
  user?: {
    id?: string;
    login?: string;
  };
  csrf_token?: string;
};

type DivisionTreeNode = {
  id: string;
  children?: DivisionTreeNode[];
};

type RoleItem = {
  id: string;
  name: string;
};

type RoleListResponse = {
  items?: RoleItem[];
  total?: number;
};

export type CreateDivisionSeedInput = {
  shortName: string;
  parentId?: string;
  fullName?: string;
  description?: string;
  regulationUrl?: string;
  mediaLinks?: Array<{ platform: string; value: string }>;
};

export type CreatePositionSeedInput = {
  divisionId: string;
  title: string;
  roleId?: string;
  roleName?: string;
  maxCount?: number;
};

export type AssignMembershipSeedInput = {
  userId: string;
  positionId: string;
};

export type CreateRoleSeedInput = {
  name: string;
  permissions: string[];
};

export type ArchiveDivisionSeedInput = {
  divisionId: string;
};

export type ArchivePositionSeedInput = {
  positionId: string;
};

export type EventLogSeedFilter = {
  eventType?: string;
  subjectType?: string;
  subjectId?: string;
  limit?: number;
  offset?: number;
};

type EventLogSeedItem = {
  id: string;
  event_type: string;
  actor_id: string;
  subject_type: string;
  subject_id: string;
  payload: Record<string, string>;
  timestamp: string;
};

type EventLogSeedListResult = {
  items: EventLogSeedItem[];
  total: number;
  limit: number;
  offset: number;
};

export type SeedHelpers = {
  loginUser(input: LoginInput): Promise<void>;
  registerUser(input: RegisterUserInput): Promise<{ body: AuthUserResult }>;
  createDivision(input: CreateDivisionSeedInput): Promise<{ id: string; shortName: string }>;
  createPosition(input: CreatePositionSeedInput): Promise<{ id: string; title: string; divisionId: string }>;
  assignMembership(input: AssignMembershipSeedInput): Promise<{ userId: string; positionId: string; membershipKey: string }>;
  createRole(input: CreateRoleSeedInput): Promise<{ id: string; name: string }>;
  archiveDivision(input: ArchiveDivisionSeedInput): Promise<{ id: string; isArchived: boolean }>;
  archivePosition(input: ArchivePositionSeedInput): Promise<{ id: string; isArchived: boolean }>;
  listEventLog(filter?: EventLogSeedFilter): Promise<EventLogSeedListResult>;
};

export function createSeedHelpers(request: APIRequestContext, entities: CreatedEntities): SeedHelpers {
  const normalizedApiBaseUrl = env.apiBaseUrl.endsWith("/") ? env.apiBaseUrl.slice(0, -1) : env.apiBaseUrl;
  let csrfToken: string | null = null;

  function apiUrl(pathname: string): string {
    const withLeadingSlash = pathname.startsWith("/") ? pathname : `/${pathname}`;
    const normalizedPath = withLeadingSlash.startsWith("/api/v1/")
      ? withLeadingSlash.slice("/api/v1".length)
      : withLeadingSlash;
    return `${normalizedApiBaseUrl}${normalizedPath}`;
  }

  async function tryLoadCsrfToken(): Promise<void> {
    if (csrfToken) {
      return;
    }
    const sessionResponse = await request.get(apiUrl("/api/v1/auth/session"));
    if (sessionResponse.status() === 401) {
      return;
    }
    await assertOk(sessionResponse, "tryLoadCsrfToken.session", 200);
    const sessionBody = await readJson<AuthSessionResult>(sessionResponse, "tryLoadCsrfToken.session");
    const token = sessionBody.csrf_token?.trim();
    if (token) {
      csrfToken = token;
    }
  }

  async function requestMutating(
    method: "POST" | "PATCH" | "DELETE",
    pathname: string,
    data?: unknown
  ): Promise<APIResponse> {
    await tryLoadCsrfToken();
    return request.fetch(apiUrl(pathname), {
      method,
      ...(data === undefined ? {} : { data }),
      ...(csrfToken ? { headers: { "X-CSRF-Token": csrfToken } } : {})
    });
  }

  async function requestPost(pathname: string, data: unknown): Promise<APIResponse> {
    return requestMutating("POST", pathname, data);
  }

  async function readJson<T>(response: APIResponse, context: string): Promise<T> {
    const raw = await response.text();
    if (!raw.trim()) {
      throw new Error(`${context} failed: empty response body`);
    }
    try {
      return JSON.parse(raw) as T;
    } catch {
      throw new Error(`${context} failed: response is not valid JSON: ${raw}`);
    }
  }

  async function assertOk(response: APIResponse, context: string, expectedStatus?: number): Promise<void> {
    const status = response.status();
    const statusMatches = expectedStatus === undefined ? response.ok() : status === expectedStatus;
    if (statusMatches) {
      return;
    }

    let details = "";
    try {
      const parsed = await readJson<ApiErrorEnvelope>(response, context);
      details = parsed.error?.message?.trim() || "";
    } catch {
      const raw = (await response.text()).trim();
      details = raw;
    }

    const suffix = details ? `: ${details}` : "";
    throw new Error(`${context} failed: expected status ${expectedStatus ?? "2xx"}, got ${status}${suffix}`);
  }

  function flattenTree(root: DivisionTreeNode): DivisionTreeNode[] {
    const result: DivisionTreeNode[] = [];
    const stack: DivisionTreeNode[] = [root];
    while (stack.length > 0) {
      const node = stack.pop();
      if (!node) {
        continue;
      }
      result.push(node);
      const children = node.children ?? [];
      for (let i = children.length - 1; i >= 0; i -= 1) {
        stack.push(children[i]);
      }
    }
    return result;
  }

  async function resolveDefaultParentDivisionId(): Promise<string> {
    const treeResponse = await request.get(apiUrl("/api/v1/divisions/tree"));
    await assertOk(treeResponse, "resolveDefaultParentDivisionId", 200);
    const tree = await readJson<DivisionTreeNode>(treeResponse, "resolveDefaultParentDivisionId");
    const nodes = flattenTree(tree);
    const preferred = nodes.find((node) => node.id === "seed-ops");
    if (preferred) {
      return preferred.id;
    }
    if (!tree.id) {
      throw new Error("resolveDefaultParentDivisionId failed: tree root id is missing");
    }
    return tree.id;
  }

  async function resolveRoleId(input: CreatePositionSeedInput): Promise<string> {
    if (input.roleId) {
      return input.roleId;
    }
    if (!input.roleName) {
      throw new Error("createPosition failed: either roleId or roleName must be provided");
    }

    const limit = 100;
    for (let offset = 0; offset < 1000; offset += limit) {
      const rolesResponse = await request.get(apiUrl(`/api/v1/roles?limit=${limit}&offset=${offset}`));
      await assertOk(rolesResponse, "createPosition.resolveRoleId", 200);
      const parsed = await readJson<RoleListResponse>(rolesResponse, "createPosition.resolveRoleId");
      const role = (parsed.items ?? []).find((item) => item.name === input.roleName);
      if (role?.id) {
        return role.id;
      }
      if ((parsed.items ?? []).length < limit || (typeof parsed.total === "number" && offset + limit >= parsed.total)) {
        break;
      }
    }
    throw new Error(`createPosition failed: role "${input.roleName}" was not found`);
  }

  async function loginUser(input: LoginInput): Promise<void> {
    const response = await requestPost(
      "/api/v1/auth/login",
      {
        login: input.login,
        password: input.password
      }
    );
    await assertOk(response, "loginUser", 200);

    const sessionResponse = await request.get(apiUrl("/api/v1/auth/session"));
    await assertOk(sessionResponse, "loginUser.session", 200);
    const sessionBody = await readJson<AuthSessionResult>(sessionResponse, "loginUser.session");
    const sessionUserId = sessionBody.user?.id?.trim();
    if (!sessionUserId) {
      throw new Error(
        "loginUser.session failed: login returned 200 but no authenticated user was found in /api/v1/auth/session; check cookie persistence and E2E_API_BASE_URL/E2E_FRONTEND_URL host alignment."
      );
    }

    const token = sessionBody.csrf_token?.trim();
    if (!token) {
      throw new Error("loginUser.session failed: csrf_token is missing in /api/v1/auth/session response");
    }
    csrfToken = token;
  }

  async function registerUser(input: RegisterUserInput): Promise<{ body: AuthUserResult }> {
    const response = await requestPost("/api/v1/auth/register", {
      login: input.login,
      password: input.password,
      first_name: input.firstName,
      gradebook_number: input.gradebookNumber,
      group_number: input.groupNumber,
      institute: input.institute,
      birth_date: input.birthDate
    });

    await assertOk(response, "registerUser", 201);
    const body = await readJson<AuthUserResult>(response, "registerUser");
    const userId = body.user?.id?.trim();
    if (!userId) {
      throw new Error("registerUser failed: response does not contain user.id");
    }
    trackCreated(entities, "users", userId);
    return { body };
  }

  async function createDivision(input: CreateDivisionSeedInput): Promise<{ id: string; shortName: string }> {
    const parentId = input.parentId ?? (await resolveDefaultParentDivisionId());
    const response = await requestPost("/api/v1/divisions", {
      parent_id: parentId,
      short_name: input.shortName,
      full_name: input.fullName ?? "",
      description: input.description ?? "",
      regulation_url: input.regulationUrl ?? "",
      media_links: input.mediaLinks ?? []
    });

    await assertOk(response, "createDivision", 201);
    const body = await readJson<{ id?: string; short_name?: string }>(response, "createDivision");
    const id = body.id?.trim();
    if (!id) {
      throw new Error("createDivision failed: response does not contain id");
    }
    trackCreated(entities, "divisions", id);
    return {
      id,
      shortName: body.short_name?.trim() || input.shortName
    };
  }

  async function createPosition(input: CreatePositionSeedInput): Promise<{ id: string; title: string; divisionId: string }> {
    const roleId = await resolveRoleId(input);
    const response = await requestPost(`/api/v1/divisions/${encodeURIComponent(input.divisionId)}/positions`, {
      title: input.title,
      role_id: roleId,
      ...(input.maxCount !== undefined ? { max_count: input.maxCount } : {})
    });

    await assertOk(response, "createPosition", 201);
    const body = await readJson<{ id?: string; title?: string; division_id?: string }>(response, "createPosition");
    const id = body.id?.trim();
    if (!id) {
      throw new Error("createPosition failed: response does not contain id");
    }
    trackCreated(entities, "positions", id);
    return {
      id,
      title: body.title?.trim() || input.title,
      divisionId: body.division_id?.trim() || input.divisionId
    };
  }

  async function assignMembership(
    input: AssignMembershipSeedInput
  ): Promise<{ userId: string; positionId: string; membershipKey: string }> {
    const response = await requestPost("/api/v1/memberships", {
      user_id: input.userId,
      position_id: input.positionId
    });

    await assertOk(response, "assignMembership", 201);
    const membershipKey = `${input.userId}::${input.positionId}`;
    trackCreated(entities, "memberships", membershipKey);
    return {
      userId: input.userId,
      positionId: input.positionId,
      membershipKey
    };
  }

  async function createRole(input: CreateRoleSeedInput): Promise<{ id: string; name: string }> {
    const response = await requestPost("/api/v1/roles", {
      name: input.name,
      permissions: input.permissions
    });

    await assertOk(response, "createRole", 201);
    const body = await readJson<{ id?: string; name?: string }>(response, "createRole");
    const id = body.id?.trim();
    if (!id) {
      throw new Error("createRole failed: response does not contain id");
    }
    trackCreated(entities, "roles", id);
    return {
      id,
      name: body.name?.trim() || input.name
    };
  }

  async function archiveDivision(input: ArchiveDivisionSeedInput): Promise<{ id: string; isArchived: boolean }> {
    const response = await requestPost(`/api/v1/divisions/${encodeURIComponent(input.divisionId)}/archive`, {});
    await assertOk(response, "archiveDivision", 200);
    const body = await readJson<{ id?: string; is_archived?: boolean }>(response, "archiveDivision");
    const id = body.id?.trim();
    if (!id) {
      throw new Error("archiveDivision failed: response does not contain id");
    }
    return {
      id,
      isArchived: body.is_archived === true
    };
  }

  async function archivePosition(input: ArchivePositionSeedInput): Promise<{ id: string; isArchived: boolean }> {
    const response = await requestPost(`/api/v1/positions/${encodeURIComponent(input.positionId)}/archive`, {});
    await assertOk(response, "archivePosition", 200);
    const body = await readJson<{ id?: string; is_archived?: boolean }>(response, "archivePosition");
    const id = body.id?.trim();
    if (!id) {
      throw new Error("archivePosition failed: response does not contain id");
    }
    return {
      id,
      isArchived: body.is_archived === true
    };
  }

  async function listEventLog(filter: EventLogSeedFilter = {}): Promise<EventLogSeedListResult> {
    const query = new URLSearchParams();
    if (filter.eventType) {
      query.set("event_type", filter.eventType);
    }
    if (filter.subjectType) {
      query.set("subject_type", filter.subjectType);
    }
    if (filter.subjectId) {
      query.set("subject_id", filter.subjectId);
    }
    if (filter.limit !== undefined) {
      query.set("limit", String(filter.limit));
    }
    if (filter.offset !== undefined) {
      query.set("offset", String(filter.offset));
    }

    const pathname = query.size > 0 ? `/api/v1/eventlog?${query.toString()}` : "/api/v1/eventlog";
    const response = await request.get(apiUrl(pathname));
    await assertOk(response, "listEventLog", 200);
    return readJson<EventLogSeedListResult>(response, "listEventLog");
  }

  return {
    loginUser,
    registerUser,
    createDivision,
    createPosition,
    assignMembership,
    createRole,
    archiveDivision,
    archivePosition,
    listEventLog
  };
}
