import { z } from "zod";

import type { LoginRequest, ProfilePatch, RegisterRequest, SocialLink } from "../api/types";

const loginRegex = /^[A-Za-z0-9_]{3,32}$/;

const requiredTrimmedRu = (message: string) =>
  z
    .string()
    .trim()
    .min(1, message);

const optionalTrimmed = z
  .string()
  .trim()
  .min(1)
  .optional()
  .or(z.literal(""));

const socialLinkSchema = z.object({
  platform: requiredTrimmedRu("Укажите платформу."),
  value: requiredTrimmedRu("Укажите ссылку.")
});

const socialLinkSchemaRu = z.object({
  platform: requiredTrimmedRu("Укажите платформу."),
  value: requiredTrimmedRu("Укажите ссылку.")
});

export const registerSchema = z.object({
  login: z
    .string()
    .trim()
    .regex(loginRegex, "Логин должен соответствовать шаблону ^[A-Za-z0-9_]{3,32}$."),
  password: requiredTrimmedRu("Пароль обязателен."),
  first_name: requiredTrimmedRu("Имя обязательно."),
  gradebook_number: requiredTrimmedRu("Зачётная книжка обязательна."),
  group_number: requiredTrimmedRu("Номер группы обязателен."),
  institute: requiredTrimmedRu("Институт обязателен."),
  birth_date: requiredTrimmedRu("Дата рождения обязательна."),
  last_name: optionalTrimmed,
  middle_name: optionalTrimmed,
  phone: optionalTrimmed,
  social_links: z.array(socialLinkSchemaRu).optional(),
  about: optionalTrimmed
});

export const loginSchema = z.object({
  login: z
    .string()
    .trim()
    .regex(loginRegex, "Логин должен соответствовать шаблону ^[A-Za-z0-9_]{3,32}$."),
  password: requiredTrimmedRu("Пароль обязателен.")
});

export const profileSchema = z.object({
  first_name: requiredTrimmedRu("Имя обязательно."),
  gradebook_number: requiredTrimmedRu("Зачётная книжка обязательна."),
  group_number: requiredTrimmedRu("Номер группы обязателен."),
  institute: requiredTrimmedRu("Институт обязателен."),
  birth_date: requiredTrimmedRu("Дата рождения обязательна."),
  last_name: optionalTrimmed,
  middle_name: optionalTrimmed,
  phone: optionalTrimmed,
  social_links: z.array(socialLinkSchema).optional(),
  about: optionalTrimmed
});

export type RegisterSchemaInput = z.input<typeof registerSchema>;
export type RegisterSchemaValues = RegisterRequest & {
  social_links?: SocialLink[];
};
export type LoginSchemaValues = LoginRequest;
export type ProfileSchemaValues = ProfilePatch & {
  social_links?: SocialLink[];
};
