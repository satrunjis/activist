const DEFAULT_API_BASE_URL = "http://localhost:8080/api/v1";
const DEFAULT_FRONTEND_URL = "http://localhost:5173";
const DEFAULT_DATABASE_URL = "postgresql://postgres:postgres@127.0.0.1:5432/activist_base";

function requiredValue(value: string | undefined, fallback: string): string {
  const normalized = value?.trim();
  return normalized && normalized.length > 0 ? normalized : fallback;
}

export const env = {
  apiBaseUrl: requiredValue(process.env.E2E_API_BASE_URL, DEFAULT_API_BASE_URL),
  frontendUrl: requiredValue(process.env.E2E_FRONTEND_URL, DEFAULT_FRONTEND_URL),
  databaseUrl: requiredValue(process.env.E2E_DATABASE_URL, DEFAULT_DATABASE_URL)
};
