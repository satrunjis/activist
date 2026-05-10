import { zodResolver } from "@hookform/resolvers/zod";
import { Plus, Trash2 } from "lucide-react";
import { useFieldArray, useForm } from "react-hook-form";

import { Button } from "../../components/ui/button";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "../../components/ui/form";
import { Input } from "../../components/ui/input";
import { request } from "../../shared/api/client";
import { adaptApiError } from "../../shared/api/errorAdapter";
import { EntityFormFrame, FormActions } from "../../shared/ui/forms";
import type { RegisterRequest } from "../../shared/api/types";
import { registerSchema, type RegisterSchemaInput } from "../../shared/forms/schemas";

type RegisterFormProps = {
  onRegistered: () => Promise<void>;
};

function optionalValue(value: string | undefined): string | undefined {
  const trimmed = value?.trim();
  return trimmed ? trimmed : undefined;
}

export function RegisterForm({ onRegistered }: RegisterFormProps) {
  const form = useForm<RegisterSchemaInput>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      login: "",
      password: "",
      first_name: "",
      gradebook_number: "",
      group_number: "",
      institute: "",
      birth_date: "",
      last_name: "",
      middle_name: "",
      phone: "",
      social_links: [],
      about: ""
    }
  });
  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "social_links"
  });

  async function onSubmit(values: RegisterSchemaInput) {
    form.clearErrors("root");
    const payload: RegisterRequest = {
      login: values.login.trim(),
      password: values.password.trim(),
      first_name: values.first_name.trim(),
      gradebook_number: values.gradebook_number.trim(),
      group_number: values.group_number.trim(),
      institute: values.institute.trim(),
      birth_date: values.birth_date.trim(),
      last_name: optionalValue(values.last_name),
      middle_name: optionalValue(values.middle_name),
      phone: optionalValue(values.phone),
      about: optionalValue(values.about),
      social_links: values.social_links
        ?.map((entry) => ({
          platform: entry.platform.trim(),
          value: entry.value.trim()
        }))
        .filter((entry) => entry.platform.length > 0 && entry.value.length > 0)
    };
    if (!payload.social_links?.length) {
      delete payload.social_links;
    }

    try {
      await request("/api/v1/auth/register", { method: "POST", body: payload });
      form.reset();
      await onRegistered();
    } catch (error) {
      form.setError("root", {
        type: "server",
        message: adaptApiError(error, "Не удалось завершить регистрацию.").message
      });
    }
  }

  return (
    <Form {...form}>
      <EntityFormFrame
        actions={(
          <FormActions
            isSubmitting={form.formState.isSubmitting}
            submitLabel="Зарегистрироваться"
            submittingLabel="Регистрация…"
          />
        )}
        onSubmit={form.handleSubmit(onSubmit)}
        rootError={form.formState.errors.root?.message}
      >
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            gap: "var(--space-4)"
          }}
        >
          <div
            style={{
              columnGap: "var(--space-4)",
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              rowGap: "var(--space-4)"
            }}
          >
            <FormField
              control={form.control}
              name="login"
              render={({ field }) => (
                <FormItem>
                  <FormLabel required>Логин</FormLabel>
                  <FormControl>
                    <Input autoComplete="username" type="text" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="password"
              render={({ field }) => (
                <FormItem>
                  <FormLabel required>Пароль</FormLabel>
                  <FormControl>
                    <Input autoComplete="new-password" type="password" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="first_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel required>Имя</FormLabel>
                  <FormControl>
                    <Input type="text" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="last_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Фамилия</FormLabel>
                  <FormControl>
                    <Input type="text" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="middle_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Отчество</FormLabel>
                  <FormControl>
                    <Input type="text" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
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
            <FormField
              control={form.control}
              name="gradebook_number"
              render={({ field }) => (
                <FormItem>
                  <FormLabel required>Зачётная книжка</FormLabel>
                  <FormControl>
                    <Input type="text" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="group_number"
              render={({ field }) => (
                <FormItem>
                  <FormLabel required>Номер группы</FormLabel>
                  <FormControl>
                    <Input type="text" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="institute"
              render={({ field }) => (
                <FormItem style={{ gridColumn: "1 / -1" }}>
                  <FormLabel required>Институт</FormLabel>
                  <FormControl>
                    <Input type="text" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="phone"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Телефон</FormLabel>
                  <FormControl>
                    <Input autoComplete="tel" type="tel" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="about"
              render={({ field }) => (
                <FormItem style={{ gridColumn: "1 / -1" }}>
                  <FormLabel>О себе</FormLabel>
                  <FormControl>
                    <textarea
                      {...field}
                      className="w-full min-h-20 px-3 py-2 text-sm font-body bg-surface border border-border rounded-input text-text-primary placeholder:text-text-muted transition-colors duration-120 focus:outline-none focus:border-border-focus focus:ring-2 focus:ring-brand-primary/20 disabled:bg-surface-subtle disabled:text-text-disabled disabled:cursor-not-allowed disabled:opacity-65"
                      rows={3}
                      style={{
                        resize: "vertical"
                      }}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
          </div>

          <section>
            <div
              style={{
                alignItems: "center",
                display: "flex",
                justifyContent: "space-between",
                marginBottom: "var(--space-2)"
              }}
            >
              <p
                style={{
                  color: "var(--color-text-primary)",
                  fontSize: "var(--text-sm)",
                  fontWeight: "var(--weight-semibold)"
                }}
              >
                Социальные сети
              </p>
              <Button
                onClick={() => append({ platform: "", value: "" })}
                size="sm"
                type="button"
                variant="ghost"
              >
                <Plus aria-hidden size={14} />
                <span>+ Добавить ссылку</span>
              </Button>
            </div>

            <div
              style={{
                display: "flex",
                flexDirection: "column",
                gap: "var(--space-2)"
              }}
            >
              {fields.map((item, index) => (
                <div
                  key={item.id}
                  style={{
                    alignItems: "center",
                    display: "flex",
                    gap: "var(--space-2)"
                  }}
                >
                  <FormField
                    control={form.control}
                    name={`social_links.${index}.platform`}
                    render={({ field }) => (
                      <FormItem style={{ width: 120 }}>
                        <FormControl>
                          <Input aria-label="Платформа" placeholder="Платформа" type="text" {...field} />
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
                          <Input aria-label="Ссылка" placeholder="https://..." type="url" {...field} />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <Button
                    aria-label="Удалить ссылку"
                    className="w-9 px-0"
                    onClick={() => remove(index)}
                    style={{ alignSelf: "flex-start", marginTop: 1 }}
                    type="button"
                    variant="ghost"
                  >
                    <Trash2 size={14} />
                  </Button>
                </div>
              ))}
            </div>
          </section>
        </div>
      </EntityFormFrame>
    </Form>
  );
}
