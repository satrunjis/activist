import { randomUUID } from "node:crypto";

export type RegisterUserInput = {
  login: string;
  password: string;
  firstName: string;
  gradebookNumber: string;
  groupNumber: string;
  institute: string;
  birthDate: string;
};

function suffix(): string {
  return randomUUID().replace(/-/g, "").slice(0, 12);
}

export function buildRegisterUserInput(): RegisterUserInput {
  const token = suffix();
  return {
    login: `e2e_${token}`,
    password: "P@ssw0rd123!",
    firstName: "E2E",
    gradebookNumber: `GB-${token}`,
    groupNumber: `G-${token.slice(0, 5)}`,
    institute: "Test Institute",
    birthDate: "2000-01-01"
  };
}
