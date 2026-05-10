import { useEffect, useState } from "react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "../../components/ui/form";
import { Input } from "../../components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../../components/ui/select";
import { request } from "../../shared/api/client";
import { adaptApiError } from "../../shared/api/errorAdapter";
import { EntityFormFrame, FormActions } from "../../shared/ui/forms";
import type { CreatePositionRequest, RoleItem } from "../../shared/api/types";

const schema = z.object({
  title: z.string().trim().min(1, "Укажите название."),
  role_id: z.string().trim().min(1, "Выберите роль."),
  max_count: z
    .string()
    .trim()
    .optional()
    .refine((value) => !value || (/^\d+$/.test(value) && Number(value) >= 1), {
      message: "Значение должно быть не меньше 1."
    })
});

type FormValues = {
  title: string;
  role_id: string;
  max_count?: string;
};

type PositionCreateFormProps = {
  divisionId: string;
  onSuccess: () => void;
  onCancel: () => void;
  onForbidden?: (message: string) => void;
};

const FORBIDDEN_SCOPE_MESSAGE = "Операция недоступна в текущем контексте.";

function toPayload(values: FormValues): CreatePositionRequest {
  const maxCount = values.max_count?.trim();
  return {
    title: values.title.trim(),
    role_id: values.role_id,
    ...(maxCount ? { max_count: Number(maxCount) } : {})
  };
}

export function PositionCreateForm({ divisionId, onSuccess, onCancel, onForbidden }: PositionCreateFormProps) {
  const [roles, setRoles] = useState<RoleItem[]>([]);
  const [rolesLoading, setRolesLoading] = useState(true);

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      title: "",
      role_id: "",
      max_count: ""
    }
  });

  useEffect(() => {
    let mounted = true;
    const run = async () => {
      setRolesLoading(true);
      try {
        const response = await request<{ items: RoleItem[] }>("/api/v1/roles?limit=100&offset=0");
        if (mounted) {
          setRoles(response.items);
        }
      } catch (error) {
        if (mounted) {
          const apiError = adaptApiError(error, "Не удалось загрузить роли.");
          form.setError("root", {
            type: "server",
            message: apiError.message
          });
        }
      } finally {
        if (mounted) {
          setRolesLoading(false);
        }
      }
    };
    void run();
    return () => {
      mounted = false;
    };
  }, [form]);

  async function handleSubmit(values: FormValues) {
    form.clearErrors("root");

    try {
      await request(`/api/v1/divisions/${encodeURIComponent(divisionId)}/positions`, {
        method: "POST",
        body: toPayload(values)
      });
      onSuccess();
    } catch (error) {
      const apiError = adaptApiError(error, "Не удалось создать должность.");
      if (apiError.kind === "forbidden" || apiError.code?.startsWith("access.")) {
        form.setError("root", { type: "server", message: FORBIDDEN_SCOPE_MESSAGE });
        onForbidden?.(FORBIDDEN_SCOPE_MESSAGE);
        return;
      }
      form.setError("root", {
        type: "server",
        message: apiError.message
      });
    }
  }

  return (
    <Form {...form}>
      <EntityFormFrame
        actions={(
          <FormActions
            cancelLabel="Отмена"
            isSubmitting={form.formState.isSubmitting}
            onCancel={onCancel}
            submitLabel="Создать должность"
            submittingLabel="Сохранение…"
          />
        )}
        formId="create-position-form"
        onSubmit={form.handleSubmit(handleSubmit)}
        rootError={form.formState.errors.root?.message}
      >
        <div className="form-grid">
          <FormField
            control={form.control}
            name="title"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Название</FormLabel>
                <FormControl>
                  <Input data-testid="position-title" {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="role_id"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Роль</FormLabel>
                <FormControl>
                  {rolesLoading ? (
                    <p className="muted-text" style={{ height: 36, display: "flex", alignItems: "center", margin: 0 }}>Загрузка ролей…</p>
                  ) : (
                    <Select name={field.name} onValueChange={field.onChange} value={field.value} modal={false}>
                      <SelectTrigger data-testid="position-role" ref={field.ref}>
                        <SelectValue placeholder="Выберите роль" />
                      </SelectTrigger>
                      <SelectContent>
                        {roles.map((role) => (
                          <SelectItem key={role.id} value={role.id}>
                            {role.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="max_count"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Максимум участников</FormLabel>
                <FormControl>
                  <Input
                    data-testid="position-max-count"
                    min={1}
                    placeholder="Без ограничений"
                    type="number"
                    {...field}
                  />
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
