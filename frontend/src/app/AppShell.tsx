import type { ReactNode } from "react";

import { TopNavBar } from "./TopNavBar";
import { getRouteByPath, type AppRoute } from "./routeContracts";
import type { UserProfile } from "../shared/api/types";

type AppShellProps = {
  route: AppRoute;
  permissions: readonly string[];
  user: UserProfile;
  onNavigate: (next: AppRoute) => void;
  onLogout: () => Promise<void>;
  isLoggingOut?: boolean;
  bannerSlot?: ReactNode;
  routeContent: ReactNode;
};

export function AppShell({
  route,
  permissions,
  user,
  onNavigate,
  onLogout,
  isLoggingOut = false,
  bannerSlot,
  routeContent
}: AppShellProps) {
  const activeRoute = getRouteByPath(route);
  const isFullScreen = Boolean(activeRoute.fullScreen);

  return (
    <div
      className="app-shell"
      style={{
        background: "var(--color-canvas)",
        minHeight: "100vh",
        minWidth: 1280
      }}
    >
      <TopNavBar
        activeRoute={route}
        isLoggingOut={isLoggingOut}
        onLogout={onLogout}
        onNavigate={onNavigate}
        permissions={permissions}
        user={user}
      />

      <div
        className="app-shell__body"
        style={{
          background: "var(--color-canvas)",
          minHeight: "100vh",
          paddingTop: "var(--topbar-height)"
        }}
      >
        <main
          data-testid="app-shell-content"
          style={isFullScreen
            ? {
                display: "flex",
                flexDirection: "column",
                height: "calc(100vh - var(--topbar-height))",
                overflow: "hidden"
              }
            : {
                display: "flex",
                flexDirection: "column",
                gap: "var(--space-6)",
                padding: "var(--space-6)"
              }}
        >
          {bannerSlot ? <div>{bannerSlot}</div> : null}

          {isFullScreen ? (
            <div style={{ flex: 1, minHeight: 0, overflow: "hidden" }}>{routeContent}</div>
          ) : (
            <section>{routeContent}</section>
          )}
        </main>
      </div>
    </div>
  );
}
