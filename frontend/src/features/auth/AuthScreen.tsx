import { useState } from "react";

import { LoginForm } from "./LoginForm";
import { RegisterForm } from "./RegisterForm";

type AuthScreenProps = {
  onLoggedIn: () => Promise<void>;
  onRegistered: () => Promise<void>;
};

type AuthTab = "login" | "register";

export function AuthScreen({ onLoggedIn, onRegistered }: AuthScreenProps) {
  const [activeTab, setActiveTab] = useState<AuthTab>("login");

  return (
    <main
      style={{
        alignItems: "center",
        background: "var(--color-canvas)",
        display: "flex",
        justifyContent: "center",
        minHeight: "100vh",
        padding: "var(--space-4)"
      }}
    >
      <div
        style={{
          width: "100%",
          maxWidth: 440
        }}
      >
        <div
          style={{
            alignItems: "center",
            display: "flex",
            flexDirection: "column",
            gap: "var(--space-2)",
            marginBottom: "var(--space-6)"
          }}
        >
          <div
            aria-hidden
            style={{
              alignItems: "center",
              background: "var(--color-brand-primary)",
              borderRadius: "var(--radius-icon)",
              color: "var(--color-text-inverse)",
              display: "flex",
              height: 40,
              justifyContent: "center",
              width: 40
            }}
          >
            <svg aria-hidden="true" height="20" viewBox="0 0 20 20" width="20">
              <text
                fill="currentColor"
                fontFamily="Raleway, Arial, sans-serif"
                fontSize="14"
                fontWeight="700"
                textAnchor="middle"
                x="10"
                y="15"
              >
                A
              </text>
            </svg>
          </div>
          <span
            style={{
              color: "var(--color-text-primary)",
              fontFamily: "var(--font-body)",
              fontSize: "var(--text-base)",
              fontWeight: "var(--weight-bold)"
            }}
          >
            Activist Base
          </span>
        </div>

        <section
          style={{
            background: "var(--color-surface)",
            borderRadius: "var(--radius-card)",
            boxShadow: "var(--shadow-lg)",
            padding: "var(--space-8) var(--space-8) var(--space-6)"
          }}
        >
          <h1
            style={{
              color: "var(--color-text-primary)",
              fontSize: "var(--text-lg)",
              fontWeight: "var(--weight-bold)",
              marginBottom: "var(--space-4)"
            }}
          >
            Добро пожаловать
          </h1>

          <div
            aria-label="Переключение режима авторизации"
            role="tablist"
            style={{
              borderBottom: "1px solid var(--color-border)",
              display: "flex",
              gap: 0,
              marginBottom: "var(--space-5)"
            }}
          >
            <button
              aria-selected={activeTab === "login"}
              onClick={() => setActiveTab("login")}
              role="tab"
              style={{
                background: "transparent",
                border: "none",
                borderBottom: activeTab === "login" ? "2px solid var(--color-brand-primary)" : "2px solid transparent",
                color: activeTab === "login" ? "var(--color-brand-primary)" : "var(--color-text-muted)",
                cursor: "pointer",
                fontSize: "var(--text-sm)",
                fontWeight: activeTab === "login" ? "var(--weight-bold)" : "var(--weight-medium)",
                marginBottom: -1,
                padding: "var(--space-2) var(--space-4)"
              }}
              type="button"
            >
              Вход
            </button>
            <button
              aria-selected={activeTab === "register"}
              onClick={() => setActiveTab("register")}
              role="tab"
              style={{
                background: "transparent",
                border: "none",
                borderBottom: activeTab === "register" ? "2px solid var(--color-brand-primary)" : "2px solid transparent",
                color: activeTab === "register" ? "var(--color-brand-primary)" : "var(--color-text-muted)",
                cursor: "pointer",
                fontSize: "var(--text-sm)",
                fontWeight: activeTab === "register" ? "var(--weight-bold)" : "var(--weight-medium)",
                marginBottom: -1,
                padding: "var(--space-2) var(--space-4)"
              }}
              type="button"
            >
              Регистрация
            </button>
          </div>

          {activeTab === "login"
            ? <LoginForm onLoggedIn={onLoggedIn} />
            : <RegisterForm onRegistered={onRegistered} />}
        </section>
      </div>
    </main>
  );
}
