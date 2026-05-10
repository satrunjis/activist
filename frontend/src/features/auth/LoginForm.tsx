import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";

import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "../../components/ui/form";
import { Input } from "../../components/ui/input";
import { request } from "../../shared/api/client";
import { adaptApiError } from "../../shared/api/errorAdapter";
import { EntityFormFrame, FormActions } from "../../shared/ui/forms";
import { loginSchema, type LoginSchemaValues } from "../../shared/forms/schemas";
import { useToast } from "../../shared/ui/feedback";

type LoginFormProps = {
  onLoggedIn: () => Promise<void>;
};

export function LoginForm({ onLoggedIn }: LoginFormProps) {
  const { showToast } = useToast();
  const form = useForm<LoginSchemaValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      login: "",
      password: ""
    }
  });

  async function onSubmit(values: LoginSchemaValues) {
    form.clearErrors("root");
    try {
      await request("/api/v1/auth/login", {
        method: "POST",
        skipUnauthorizedRedirect: true,
        body: {
          login: values.login.trim(),
          password: values.password.trim()
        }
      });
      await onLoggedIn();
    } catch (error) {
      showToast("error", "Не удалось выполнить вход", adaptApiError(error, "Не удалось выполнить вход.").message);
    }
  }

  return (
    <Form {...form}>
      <EntityFormFrame
        actions={<FormActions isSubmitting={form.formState.isSubmitting} submitLabel="Войти" submittingLabel="Вход…" />}
        onSubmit={form.handleSubmit(onSubmit)}
      >
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            gap: "var(--space-4)"
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
                  <Input autoComplete="current-password" type="password" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>
      </EntityFormFrame>
    </Form>
  );
}
