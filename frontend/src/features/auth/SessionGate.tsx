import { useCallback, useEffect, useState } from "react";

import {
  clearCsrfToken,
  request,
  setCsrfToken
} from "../../shared/api/client";
import { adaptApiError } from "../../shared/api/errorAdapter";
import type { AuthSessionResponse, EventLogListResponse, UserProfile } from "../../shared/api/types";
import { AuthScreen } from "./AuthScreen";

type SessionGateChildrenProps = {
  user: UserProfile;
  permissions: string[];
  onLogout: () => Promise<void>;
  isLoggingOut: boolean;
};

type SessionGateProps = {
  children: (session: SessionGateChildrenProps) => React.ReactNode;
};

type SessionState =
  | { status: "loading" }
  | { status: "anonymous" }
  | { status: "authenticated"; user: UserProfile; permissions: string[] };

function isUnauthorized(error: unknown): boolean {
  const apiError = adaptApiError(error, "Не удалось проверить сессию.");
  return apiError.status === 401 || apiError.code?.startsWith("auth.") === true;
}

export function SessionGate({ children }: SessionGateProps) {
  const [sessionState, setSessionState] = useState<SessionState>({ status: "loading" });
  const [isBootstrapping, setBootstrapping] = useState(true);
  const [isLoggingOut, setLoggingOut] = useState(false);

  const bootstrapSession = useCallback(async () => {
    setBootstrapping(true);
    try {
      const response = await request<AuthSessionResponse>("/api/v1/auth/session");
      setCsrfToken(response.csrf_token);
      const permissions = Array.isArray(response.permissions) ? [...response.permissions] : [];
      if (!Array.isArray(response.permissions)) {
        try {
          await request<EventLogListResponse>("/api/v1/eventlog?limit=1&offset=0");
          permissions.push("can_view_audit_log");
        } catch {
          // Keep the permissions list empty when backend does not expose session permissions
          // and eventlog endpoint is forbidden for current actor.
        }
      }
      setSessionState({
        status: "authenticated",
        user: response.user,
        permissions
      });
    } catch (error) {
      clearCsrfToken();
      if (!isUnauthorized(error)) {
        console.error(error);
      }
      setSessionState({ status: "anonymous" });
    } finally {
      setBootstrapping(false);
    }
  }, []);

  useEffect(() => {
    void bootstrapSession();
  }, [bootstrapSession]);

  const handleLogout = useCallback(async () => {
    setLoggingOut(true);
    try {
      await request("/api/v1/auth/logout", { method: "POST" });
    } catch (error) {
      if (!isUnauthorized(error)) {
        console.error(error);
      }
    } finally {
      clearCsrfToken();
      setSessionState({ status: "anonymous" });
      setLoggingOut(false);
    }
  }, []);

  if (isBootstrapping || sessionState.status === "loading") {
    return (
      <main
        style={{
          alignItems: "center",
          background: "var(--color-canvas)",
          display: "flex",
          justifyContent: "center",
          minHeight: "100vh"
        }}
      >
        <section
          aria-live="polite"
          style={{
            alignItems: "center",
            display: "flex",
            flexDirection: "column",
            gap: "var(--space-2)"
          }}
        >
          <span
            aria-hidden="true"
            style={{
              animation: "spin 0.8s linear infinite",
              border: "2px solid var(--color-border)",
              borderRadius: "var(--radius-full)",
              borderTopColor: "var(--color-brand-primary)",
              display: "inline-block",
              height: 24,
              width: 24
            }}
          />
          <p
            style={{
              color: "var(--color-text-muted)",
              fontSize: "var(--text-sm)"
            }}
          >
            Проверка сессии…
          </p>
        </section>
      </main>
    );
  }

  if (sessionState.status === "anonymous") {
    return <AuthScreen onLoggedIn={bootstrapSession} onRegistered={bootstrapSession} />;
  }

  return children({
    user: sessionState.user,
    permissions: sessionState.permissions,
    onLogout: handleLogout,
    isLoggingOut
  });
}
