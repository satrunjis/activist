import { ClipboardList, LogOut, Network, Search, Shield, User, type LucideIcon } from "lucide-react";
import { useMemo } from "react";

import { getVisiblePrimaryRoutes, type AppRoute } from "./routeContracts";
import { cn } from "../lib/utils";
import type { UserProfile } from "../shared/api/types";

type TopNavBarProps = {
  activeRoute: AppRoute;
  permissions: readonly string[];
  user: UserProfile;
  onNavigate: (next: AppRoute) => void;
  onLogout: () => Promise<void>;
  isLoggingOut?: boolean;
};

const ROUTE_ICONS: Record<AppRoute, LucideIcon> = {
  "/": Network,
  "/roles": Shield,
  "/profile": User,
  "/search": Search,
  "/audit-log": ClipboardList
};


function getUsername(user: UserProfile): string {
  const fullName = [user.first_name, user.last_name].filter(Boolean).join(" ").trim();
  return user.login?.trim() || fullName || user.id;
}

function getInitials(username: string): string {
  const normalized = username.trim();
  if (!normalized) return "U";
  return normalized.slice(0, 2).toUpperCase();
}

export function TopNavBar({
  activeRoute,
  permissions,
  user,
  onNavigate,
  onLogout,
  isLoggingOut = false
}: TopNavBarProps) {
  const routes = useMemo(() => getVisiblePrimaryRoutes(permissions), [permissions]);

  const username = getUsername(user);
  const initials = getInitials(username);

  return (
    <header
      className="fixed left-0 right-0 top-0 z-[100] grid h-14 grid-cols-[auto_1fr_auto] items-center border-b border-border bg-nav-bg shadow-sm"
      style={{ paddingLeft: 12, paddingRight: 12 }}
    >
      <div className="flex items-center gap-2 justify-self-start">
        <div aria-hidden className="flex h-8 w-8 items-center justify-center rounded-icon bg-brand-primary text-text-inverse">
          <svg aria-hidden="true" height="16" viewBox="0 0 16 16" width="16">
            <text
              fill="currentColor"
              fontFamily="Raleway, Arial, sans-serif"
              fontSize="11"
              fontWeight="700"
              textAnchor="middle"
              x="8"
              y="12"
            >
              A
            </text>
          </svg>
        </div>
        <span className="font-body text-base font-bold text-nav-text-active">
          Activist Base
        </span>
      </div>

      <nav aria-label="Основная навигация" className="flex min-w-0 justify-center gap-2">
        {routes.map((route) => {
          const Icon = ROUTE_ICONS[route.path];
          const isActive = route.path === activeRoute;

          return (
            <button
              aria-current={isActive ? "page" : undefined}
              data-testid={route.testId}
              key={route.path}
              onClick={() => onNavigate(route.path)}
              className={cn(
                "group relative flex h-9 items-center gap-2 rounded-card px-[14px] font-body text-sm font-medium leading-tight transition-colors duration-120",
                isActive
                  ? "text-nav-text-active"
                  : "text-nav-text hover:text-nav-text-active"
              )}
              type="button"
            >
              <Icon className={cn("h-4 w-4 transition-opacity duration-120", isActive ? "opacity-100" : "opacity-65 group-hover:opacity-100")} />
              <span>{route.label}</span>
              {isActive ? (
                <span
                  aria-hidden
                  className="absolute -bottom-px left-[14px] right-[14px] h-0.5 rounded-sm bg-nav-active-indicator"
                />
              ) : null}
            </button>
          );
        })}
      </nav>

      <div className="flex items-center gap-3 justify-self-end">
        <div aria-hidden className="flex h-8 w-8 items-center justify-center rounded-full bg-brand-primary text-xs font-bold text-text-inverse">
          {initials}
        </div>
        <span className="text-sm font-medium text-nav-text-active">
          {username}
        </span>
        <button
          aria-label="Выйти"
          disabled={isLoggingOut}
          onClick={() => {
            void onLogout();
          }}
          className={cn(
            "flex items-center rounded-card p-2 transition-colors duration-120",
            isLoggingOut
              ? "cursor-default text-nav-text"
              : "text-nav-text hover:text-nav-text-active"
          )}
          type="button"
        >
          <LogOut className="h-4 w-4" />
        </button>
      </div>
    </header>
  );
}
