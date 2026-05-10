import { useCallback, useEffect, useState } from "react";

import { AppShell } from "./AppShell";
import {
  CRITICAL_NAV_TEST_IDS,
  hasStablePrimaryNavigationContracts,
  normalizeRoute,
  type AppRoute
} from "./routeContracts";
import { SessionGate } from "../features/auth/SessionGate";
import { AuditLogPage } from "../features/audit/AuditLogPage";
import { DivisionsPage } from "../features/divisions/DivisionsPage";
import { ProfilePage } from "../features/profile/ProfilePage";
import { RolesPage } from "../features/roles/RolesPage";
import { SearchPage } from "../features/search/SearchPage";
import { ConfirmDialog, ToastProvider } from "../shared/ui/feedback";
import { GlobalApiBanner } from "../shared/ui/feedback/GlobalApiBanner";
import type { GlobalApiBannerProps } from "../shared/ui/feedback/GlobalApiBanner";

const REQUIRED_APP_NAV_TEST_IDS = ["nav-home", "nav-audit-log"] as const;

if (!hasStablePrimaryNavigationContracts()) {
  throw new Error(
    `Primary route selectors must remain stable: ${CRITICAL_NAV_TEST_IDS.join(", ")}`
  );
}

if (!REQUIRED_APP_NAV_TEST_IDS.every((testId) => CRITICAL_NAV_TEST_IDS.includes(testId))) {
  throw new Error("App requires stable nav-home and nav-audit-log selectors for e2e actions.");
}

type GlobalApiErrorState = Omit<GlobalApiBannerProps, "onDismiss"> | null;

const noop = () => undefined;

export function App() {
  const [route, setRoute]             = useState<AppRoute>(() => normalizeRoute(window.location.pathname));
  const [globalError, setGlobalError] = useState<GlobalApiErrorState>(null);

  useEffect(() => {
    function onPopState() {
      setRoute(normalizeRoute(window.location.pathname));
    }
    window.addEventListener("popstate", onPopState);
    return () => window.removeEventListener("popstate", onPopState);
  }, []);

  const navigate = useCallback((next: AppRoute) => {
    if (window.location.pathname !== next) {
      window.history.pushState({}, "", next);
    }
    setRoute(next);
  }, []);

  function dismissGlobalError() {
    setGlobalError(null);
  }

  return (
    <ToastProvider>
      <div className="app-root app-root--shell">
        <SessionGate>
          {({ user, permissions, onLogout, isLoggingOut }) => {
            const hasAuditPermission = permissions.includes("can_view_audit_log");

            let routeContent: React.ReactNode;
            switch (route) {
              case "/":
                routeContent = <DivisionsPage />;
                break;
              case "/roles":
                routeContent = <RolesPage permissions={permissions} />;
                break;
              case "/profile":
                routeContent = <ProfilePage userId={user.id} />;
                break;
              case "/search":
                routeContent = <SearchPage onGlobalError={setGlobalError} />;
                break;
              case "/audit-log":
                routeContent = hasAuditPermission
                  ? <AuditLogPage />
                  : <p data-testid="audit-log-forbidden">Операция запрещена.</p>;
                break;
              default:
                routeContent = <DivisionsPage />;
            }

            return (
              <AppShell
                bannerSlot={
                  globalError ? (
                    <GlobalApiBanner
                      message={globalError.message}
                      onDismiss={dismissGlobalError}
                      onRetry={globalError.onRetry}
                      retryLabel={globalError.retryLabel}
                      severity={globalError.severity}
                    />
                  ) : null
                }
                isLoggingOut={isLoggingOut}
                onLogout={onLogout}
                onNavigate={navigate}
                permissions={permissions}
                route={route}
                routeContent={routeContent}
                user={user}
              />
            );
          }}
        </SessionGate>
      </div>
      <ConfirmDialog
        body=""
        confirmLabel=""
        destructive={false}
        onCancel={noop}
        onConfirm={noop}
        open={false}
        title=""
      />
    </ToastProvider>
  );
}
