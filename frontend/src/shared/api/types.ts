export type ApiErrorDetail = {
  code: string;
  message: string;
  field?: string;
};

export type ApiErrorEnvelope = {
  error: ApiErrorDetail;
};

export type SocialLink = {
  platform: string;
  value: string;
};

export type RegisterRequest = {
  login: string;
  password: string;
  first_name: string;
  gradebook_number: string;
  group_number: string;
  institute: string;
  birth_date: string;
  last_name?: string;
  middle_name?: string;
  phone?: string;
  social_links?: SocialLink[];
  about?: string;
};

export type LoginRequest = {
  login: string;
  password: string;
};

export type MembershipSummary = {
  position_id: string;
  position_name: string;
  division_id: string;
  division_name: string;
  role_name: string;
};

export type UserProfile = {
  id: string;
  login?: string;
  first_name: string;
  last_name?: string;
  middle_name?: string;
  gradebook_number?: string;
  group_number?: string;
  institute?: string;
  birth_date?: string;
  phone?: string;
  social_links?: SocialLink[];
  about?: string;
  memberships: MembershipSummary[];
};

export type ProfilePatch = {
  first_name: string;
  gradebook_number: string;
  group_number: string;
  institute: string;
  birth_date: string;
  last_name?: string;
  middle_name?: string;
  phone?: string;
  social_links?: SocialLink[];
  about?: string;
};

export type AuthSession = {
  expires_at: string;
  idle_expires_at: string;
};

export type AuthSessionResponse = {
  user: UserProfile;
  session: AuthSession;
  csrf_token: string;
  permissions?: string[];
};

export type EventLogItem = {
  id: string;
  event_type: string;
  actor_id: string;
  subject_type: string;
  subject_id: string;
  payload: Record<string, string>;
  timestamp: string;
};

export type EventLogListResponse = {
  items: EventLogItem[];
  total: number;
  limit: number;
  offset: number;
};

export type DivisionMediaLink = {
  platform: string;
  value: string;
};

export type DivisionTreeNode = {
  id: string;
  parent_id?: string;
  short_name: string;
  full_name?: string;
  description?: string;
  regulation_url?: string;
  media_links?: DivisionMediaLink[];
  is_archived: boolean;
  has_children: boolean;
  children_count: number;
  positions_count?: number;
  members_count?: number;
  children?: DivisionTreeNode[];
};

export type DivisionChildItem = {
  id: string;
  parent_id?: string;
  short_name: string;
  full_name?: string;
  description?: string;
  regulation_url?: string;
  media_links?: DivisionMediaLink[];
  is_archived: boolean;
  has_children: boolean;
  children_count: number;
};

export type DivisionChildrenResponse = {
  items: DivisionChildItem[];
};

export type DivisionCreateRequest = {
  parent_id?: string;
  short_name: string;
  full_name: string;
  description: string;
  regulation_url: string;
  media_links: DivisionMediaLink[];
};

export type DivisionPatchRequest = {
  parent_id?: string;
  short_name?: string;
  full_name?: string;
  description?: string;
  regulation_url?: string;
  media_links?: DivisionMediaLink[];
};

export type DivisionEntity = {
  id: string;
  parent_id?: string;
  short_name: string;
  full_name: string;
  description: string;
  regulation_url?: string;
  media_links?: DivisionMediaLink[];
  is_archived: boolean;
};

export type PermissionScope = "self" | "current_division" | "current_and_descendants";

export type RolePermission = {
  code: string;
  scope: PermissionScope;
};

export type RoleItem = {
  id: string;
  name: string;
  permissions?: RolePermission[];
};

export type RoleListResponse = {
  items: RoleItem[];
  total: number;
  limit: number;
  offset: number;
};

export type CreateRoleRequest = {
  name: string;
  permissions: RolePermission[];
};

export type EditRoleRequest = {
  name?: string;
  permissions?: RolePermission[];
};

export type RoleWithPermissions = {
  id: string;
  name: string;
  permissions?: RolePermission[];
};

export type PositionItem = {
  id: string;
  title: string;
  role_id: string;
  division_id: string;
  max_count?: number;
  is_archived: boolean;
};

export type PositionMemberItem = {
  user_id: string;
  position_id: string;
  position_name: string;
  role_id: string;
  role_name: string;
  division_id: string;
  division_name: string;
};

export type PositionMembersResponse = {
  items: PositionMemberItem[];
};

export type CreatePositionRequest = {
  title: string;
  role_id: string;
  max_count?: number;
};

export type AssignMemberRequest = {
  user_id: string;
  position_id: string;
};

export type UserSearchRequest = {
  first_name?: string;
  last_name?: string;
  middle_name?: string;
  login?: string;
  group_number?: string;
  institute?: string;
  about?: string;
  position_title?: string;
  role_name?: string;
  include_archived?: boolean;
  limit?: number;
  offset?: number;
};

export type SearchUserItem = {
  id: string;
  first_name: string;
  last_name: string;
  middle_name: string;
  login?: string;
  gradebook_number?: string;
  group_number?: string;
  institute?: string;
  birth_date?: string;
  phone?: string;
  social_links?: SocialLink[];
  about?: string;
  memberships?: MembershipSummary[];
};

export type UserSearchResponse = {
  items: SearchUserItem[];
  total: number;
};
