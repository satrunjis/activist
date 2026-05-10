import { Plus, Trash2 } from "lucide-react";
import { useEffect, useMemo } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useFieldArray, useForm } from "react-hook-form";
import { z } from "zod";

import { Button } from "../../components/ui/button";
import { Card, CardContent, CardHeader } from "../../components/ui/card";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "../../components/ui/form";
import { Input } from "../../components/ui/input";
import { request } from "../../shared/api/client";
import { adaptApiError } from "../../shared/api/errorAdapter";
import { EntityFormFrame, FormActions } from "../../shared/ui/forms";
import type { SocialLink, UserProfile } from "../../shared/api/types";

type ProfileFormProps = {
  userId: string;
  profile: UserProfile;
  onSaved: (profile: UserProfile) => void;
  onCancel: () => void;
};

type ProfileFormValues = {
  first_name: string;
  last_name: string;
  middle_name: string;
  gradebook_number: string;
  group_number: string;
  institute: string;
  birth_date: string;
  phone: string;
  about: string;
  social_links: SocialLink[];
};

const profileSocialLinkSchema = z
  .object({
    platform: z.string(),
    value: z.string()
  })
  .superRefine((value, context) => {
    const platform = value.platform.trim();
    const linkValue = value.value.trim();
    if (!platform && !linkValue) {
      return;
    }
    if (!platform) {
      context.addIssue({ code: z.ZodIssueCode.custom, message: "Укажите платформу.", path: ["platform"] });
    }
    if (!linkValue) {
      context.addIssue({ code: z.ZodIssueCode.custom, message: "Укажите ссылку.", path: ["value"] });
    }
  });

function buildProfileResolverSchema(visibleFields: Record<keyof ProfileFormValues, boolean>) {
  const requiredText = (enabled: boolean, message: string) => (enabled ? z.string().trim().min(1, message) : z.string());

  return z.object({
    first_name: requiredText(visibleFields.first_name, "Имя обязательно."),
    last_name: z.string(),
    middle_name: z.string(),
    gradebook_number: requiredText(visibleFields.gradebook_number, "Зачётная книжка обязательна."),
    group_number: requiredText(visibleFields.group_number, "Номер группы обязателен."),
    institute: requiredText(visibleFields.institute, "Институт обязателен."),
    birth_date: requiredText(visibleFields.birth_date, "Дата рождения обязательна."),
    phone: z.string(),
    about: z.string(),
    social_links: z.array(visibleFields.social_links ? profileSocialLinkSchema : z.object({ platform: z.string(), value: z.string() }))
  });
}

function hasField(profile: UserProfile, key: keyof ProfileFormValues): boolean {
  if (key === "social_links") {
    return profile.social_links !== undefined;
  }
  return profile[key as keyof UserProfile] !== undefined;
}

function toDefaults(profile: UserProfile): ProfileFormValues {
  return {
    first_name: profile.first_name ?? "",
    last_name: profile.last_name ?? "",
    middle_name: profile.middle_name ?? "",
    gradebook_number: profile.gradebook_number ?? "",
    group_number: profile.group_number ?? "",
    institute: profile.institute ?? "",
    birth_date: profile.birth_date ?? "",
    phone: profile.phone ?? "",
    about: profile.about ?? "",
    social_links: profile.social_links?.map((item) => ({ ...item })) ?? []
  };
}

export function ProfileForm({ userId, profile, onSaved, onCancel }: ProfileFormProps) {
  const visibleFields = useMemo(
    () => ({
      first_name: hasField(profile, "first_name"),
      last_name: hasField(profile, "last_name"),
      middle_name: hasField(profile, "middle_name"),
      gradebook_number: hasField(profile, "gradebook_number"),
      group_number: hasField(profile, "group_number"),
      institute: hasField(profile, "institute"),
      birth_date: hasField(profile, "birth_date"),
      phone: hasField(profile, "phone"),
      about: hasField(profile, "about"),
      social_links: hasField(profile, "social_links")
    }),
    [profile]
  );

  const defaultValues = useMemo(() => toDefaults(profile), [profile]);
  const resolverSchema = useMemo(() => buildProfileResolverSchema(visibleFields), [visibleFields]);
  const form = useForm<ProfileFormValues>({
    resolver: zodResolver(resolverSchema),
    defaultValues
  });
  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "social_links"
  });

  useEffect(() => {
    form.reset(defaultValues);
  }, [defaultValues, form]);

  async function onSubmit(values: ProfileFormValues) {
    form.clearErrors("root");

    const payload: Partial<ProfileFormValues> = {};
    const editableKeys = Object.entries(visibleFields).filter(([, visible]) => visible);

    (Object.keys(visibleFields) as Array<keyof typeof visibleFields>).forEach((key) => {
      if (!visibleFields[key]) {
        return;
      }

      if (key === "social_links") {
        payload.social_links = (values.social_links ?? [])
          .map((entry) => ({
            platform: entry.platform.trim(),
            value: entry.value.trim()
          }))
          .filter((entry) => entry.platform.length > 0 || entry.value.length > 0);
        return;
      }

      const normalized = (values[key] ?? "").trim();
      payload[key] = normalized;
    });

    if (!editableKeys.length) {
      form.setError("root", { type: "manual", message: "Нет доступных полей для редактирования." });
      return;
    }
    if (Array.isArray(payload.social_links) && payload.social_links.length === 0) {
      delete payload.social_links;
    }

    try {
      const updated = await request<UserProfile>(`/api/v1/users/${encodeURIComponent(userId)}`, {
        method: "PATCH",
        body: payload
      });
      onSaved(updated);
    } catch (error) {
      form.setError("root", {
        type: "server",
        message: adaptApiError(error, "Не удалось обновить профиль.").message
      });
    }
  }

  return (
    <Card>
      <CardHeader style={{ padding: "var(--space-6) var(--space-6) var(--space-4)" }}>
        <h2 style={{ margin: 0, fontSize: "var(--text-lg)", fontWeight: "var(--weight-bold)", color: "var(--color-text-primary)" }}>
          Редактирование профиля
        </h2>
      </CardHeader>
      <CardContent style={{ padding: "0 var(--space-6) var(--space-6)" }}>
        <Form {...form}>
          <EntityFormFrame
            actions={(
              <FormActions
                cancelLabel="Отмена"
                isSubmitting={form.formState.isSubmitting}
                onCancel={onCancel}
                submitLabel="Сохранить"
                submittingLabel="Сохранение…"
              />
            )}
            onSubmit={form.handleSubmit(onSubmit)}
            rootError={form.formState.errors.root?.message}
          >
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "var(--space-4)" }}>
              {visibleFields.first_name ? (
                <FormField
                  control={form.control}
                  name="first_name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel required>Имя</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              ) : null}

              {visibleFields.last_name ? (
                <FormField
                  control={form.control}
                  name="last_name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Фамилия</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              ) : null}

              {visibleFields.middle_name ? (
                <FormField
                  control={form.control}
                  name="middle_name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Отчество</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              ) : null}

              {visibleFields.birth_date ? (
                <FormField
                  control={form.control}
                  name="birth_date"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel required>Дата рождения</FormLabel>
                      <FormControl>
                        <Input type="date" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              ) : null}

              {visibleFields.gradebook_number ? (
                <FormField
                  control={form.control}
                  name="gradebook_number"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel required>Зачётная книжка</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              ) : null}

              {visibleFields.group_number ? (
                <FormField
                  control={form.control}
                  name="group_number"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel required>Группа</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              ) : null}

              {visibleFields.institute ? (
                <FormField
                  control={form.control}
                  name="institute"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel required>Институт</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              ) : null}

              {visibleFields.phone ? (
                <FormField
                  control={form.control}
                  name="phone"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Телефон</FormLabel>
                      <FormControl>
                        <Input autoComplete="tel" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              ) : null}

              {visibleFields.about ? (
                <FormField
                  control={form.control}
                  name="about"
                  render={({ field }) => (
                    <FormItem style={{ gridColumn: "1 / -1" }}>
                      <FormLabel>О себе</FormLabel>
                      <FormControl>
                        <textarea
                          name={field.name}
                          onBlur={field.onBlur}
                          onChange={field.onChange}
                          ref={field.ref}
                          rows={3}
                          style={{
                            width: "100%",
                            minHeight: 80,
                            border: "1px solid var(--color-border)",
                            borderRadius: "var(--radius-input)",
                            background: "var(--color-surface)",
                            padding: "var(--space-2) var(--space-3)",
                            fontFamily: "var(--font-body)",
                            fontSize: "var(--text-sm)",
                            color: "var(--color-text-primary)",
                            resize: "vertical"
                          }}
                          value={field.value}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              ) : null}
            </div>

            {visibleFields.social_links ? (
              <section style={{ marginTop: "var(--space-4)" }}>
                <div
                  style={{
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "space-between",
                    marginBottom: "var(--space-2)"
                  }}
                >
                  <p style={{ margin: 0, fontSize: "var(--text-sm)", fontWeight: "var(--weight-semibold)", color: "var(--color-text-primary)" }}>
                    Социальные сети
                  </p>
                  <Button onClick={() => append({ platform: "", value: "" })} size="sm" type="button" variant="ghost">
                    <Plus size={14} />
                    <span>Добавить ссылку</span>
                  </Button>
                </div>

                <div style={{ display: "flex", flexDirection: "column", gap: "var(--space-2)" }}>
                  {fields.map((item, index) => (
                    <div key={item.id} style={{ display: "flex", alignItems: "flex-start", gap: "var(--space-2)" }}>
                      <FormField
                        control={form.control}
                        name={`social_links.${index}.platform`}
                        render={({ field }) => (
                          <FormItem style={{ width: 120 }}>
                            <FormControl>
                              <Input placeholder="vk" {...field} />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                      <FormField
                        control={form.control}
                        name={`social_links.${index}.value`}
                        render={({ field }) => (
                          <FormItem style={{ flex: 1 }}>
                            <FormControl>
                              <Input placeholder="https://…" {...field} />
                            </FormControl>
                            <FormMessage />
                          </FormItem>
                        )}
                      />
                      <Button
                        aria-label="Удалить ссылку"
                        className="w-9 px-0"
                        onClick={() => remove(index)}
                        style={{ marginTop: 1 }}
                        type="button"
                        variant="ghost"
                      >
                        <Trash2 size={14} />
                      </Button>
                    </div>
                  ))}
                </div>
              </section>
            ) : null}
          </EntityFormFrame>
        </Form>
      </CardContent>
    </Card>
  );
}
