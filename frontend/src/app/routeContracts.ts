export type AppRoute = "/" | "/roles" | "/profile" | "/search" | "/audit-log";

export type PrimaryNavigationTestId =
  | "nav-home"
  | "nav-roles"
  | "nav-profile"
  | "nav-search"
  | "nav-audit-log";

export type ShellPermission = "can_view_audit_log";

export type RouteRegion = "page-header" | "page-content";

export type PrimaryRoute = {
  path: AppRoute;
  label: string;
  testId: PrimaryNavigationTestId;
  requiredPermission?: ShellPermission;
  regions: readonly RouteRegion[];
  /** When true the shell renders without padding so the page fills 100% height */
  fullScreen?: boolean;
};

export const PRIMARY_ROUTES: readonly PrimaryRoute[] = [
  {
    path: "/",
    label: "Оргструктура",
    testId: "nav-home",
    regions: ["page-content"],
    fullScreen: true
  },
  {
    path: "/roles",
    label: "Роли",
    testId: "nav-roles",
    regions: ["page-header", "page-content"]
  },
  {
    path: "/profile",
    label: "Профиль",
    testId: "nav-profile",
    regions: ["page-header", "page-content"]
  },
  {
    path: "/search",
    label: "Поиск",
    testId: "nav-search",
    regions: ["page-header", "page-content"]
  },
  {
    path: "/audit-log",
    label: "Журнал",
    testId: "nav-audit-log",
    requiredPermission: "can_view_audit_log",
    regions: ["page-header", "page-content"]
  }
] as const;

export const CRITICAL_NAV_TEST_IDS: readonly PrimaryNavigationTestId[] = [
  "nav-home",
  "nav-audit-log"
];

const ROUTE_BY_PATH: Record<AppRoute, PrimaryRoute> = {
  "/":          PRIMARY_ROUTES[0],
  "/roles":     PRIMARY_ROUTES[1],
  "/profile":   PRIMARY_ROUTES[2],
  "/search":    PRIMARY_ROUTES[3],
  "/audit-log": PRIMARY_ROUTES[4]
};

export function normalizeRoute(pathname: string): AppRoute {
  if (pathname === "/audit-log") return "/audit-log";
  if (pathname === "/roles")     return "/roles";
  if (pathname === "/profile")   return "/profile";
  if (pathname === "/search")    return "/search";
  return "/";
}

export function hasRoutePermission(
  route: PrimaryRoute,
  permissions: readonly string[]
): boolean {
  if (!route.requiredPermission) return true;
  return permissions.includes(route.requiredPermission);
}

export function getVisiblePrimaryRoutes(permissions: readonly string[]): PrimaryRoute[] {
  return PRIMARY_ROUTES.filter((route) => hasRoutePermission(route, permissions));
}

export function getRouteByPath(route: AppRoute): PrimaryRoute {
  return ROUTE_BY_PATH[route];
}

export function hasStablePrimaryNavigationContracts(): boolean {
  return CRITICAL_NAV_TEST_IDS.every((testId) =>
    PRIMARY_ROUTES.some((route) => route.testId === testId)
  );
}
