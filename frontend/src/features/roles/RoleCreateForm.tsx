import { zodResolver } from "@hookform/resolvers/zod";
import { useMemo, useState, type CSSProperties } from "react";
import { useForm } from "react-hook-form";

import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "../../components/ui/form";
import { Input } from "../../components/ui/input";
import { request } from "../../shared/api/client";
import { adaptApiError } from "../../shared/api/errorAdapter";
import { EntityFormFrame, FormActions } from "../../shared/ui/forms";
import type { CreateRoleRequest, PermissionScope, RolePermission } from "../../shared/api/types";
import { useToast } from "../../shared/ui/feedback";
import { PERMISSION_OPTIONS, SCOPE_OPTIONS, type PermissionOption } from "./permissionOptions";
import { roleFormSchema, type RoleFormValues } from "./roleFormSchema";

const HIDDEN_PERMISSION_CODE = "can_manage_roles";

type RoleCreateFormProps = {
  onSuccess: () => void;
  onCancel: () => void;
};

function toPayload(values: RoleFormValues, scopeByCode: Record<string, PermissionScope>): CreateRoleRequest {
  const permissions: RolePermission[] = values.permissions.map((code) => ({
    code,
    scope: scopeByCode[code] ?? "current_division"
  }));
  return {
    name: values.name.trim(),
    permissions
  };
}

export function RoleCreateForm({ onSuccess, onCancel }: RoleCreateFormProps) {
  const { showToast } = useToast();
  // Scope state is kept for unchecked permissions so the selection is restored
  // if the user re-checks the same permission within the same form session.
  const [scopeByCode, setScopeByCode] = useState<Record<string, PermissionScope>>(() => getDefaultScopeMap());
  const form = useForm<RoleFormValues>({
    resolver: zodResolver(roleFormSchema),
    defaultValues: {
      name: "",
      permissions: []
    }
  });

  const selectedPermissions = form.watch("permissions");
  const roleName = form.watch("name");
  const selectedSet = useMemo(() => new Set(selectedPermissions ?? []), [selectedPermissions]);
  const groupedPermissions = useMemo(() => {
    const grouped = new Map<string, PermissionOption[]>();
    for (const option of PERMISSION_OPTIONS) {
      if (option.value === HIDDEN_PERMISSION_CODE) {
        continue;
      }
      const items = grouped.get(option.category) ?? [];
      items.push(option);
      grouped.set(option.category, items);
    }
    return Array.from(grouped.entries());
  }, []);

  async function handleSubmit(values: RoleFormValues) {
    form.clearErrors("root");

    try {
      await request("/api/v1/roles", { method: "POST", body: toPayload(values, scopeByCode) });
      form.reset({ name: "", permissions: [] });
      setScopeByCode(getDefaultScopeMap());
      showToast("success", "Роль сохранена");
      onSuccess();
    } catch (error) {
      const apiError = adaptApiError(error, "Не удалось создать роль.");
      const mapped = mapRoleError(apiError.code, apiError.message, apiError.kind);
      form.setError("root", { type: "server", message: mapped });
    }
  }

  return (
    <Form {...form}>
      <EntityFormFrame
        actions={(
          <FormActions
            cancelLabel="Отмена"
            disableSubmit={form.formState.isSubmitting || roleName.trim().length === 0}
            isSubmitting={form.formState.isSubmitting}
            onCancel={onCancel}
            submitLabel="Сохранить"
            submittingLabel="Сохранение…"
          />
        )}
        formId="create-role-form"
        onSubmit={form.handleSubmit(handleSubmit)}
        rootError={form.formState.errors.root?.message}
      >
        <FormField
          control={form.control}
          name="name"
          render={({ field }) => (
            <FormItem>
              <FormLabel required>Название</FormLabel>
              <FormControl>
                <Input data-testid="role-name" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />

        <section style={{ marginTop: "var(--space-4)" }}>
          <p
            style={{
              margin: "0 0 var(--space-2)",
              fontSize: "var(--text-xs)",
              fontWeight: "var(--weight-semibold)",
              color: "var(--color-text-muted)",
              textTransform: "uppercase",
              letterSpacing: "0.06em"
            }}
          >
            Разрешения
          </p>

          <div style={{ display: "grid", gap: "var(--space-1)" }}>
            {groupedPermissions.map(([category, options], categoryIndex) => (
              <div key={category}>
                <p
                  style={{
                    margin: categoryIndex === 0 ? "0 0 var(--space-1)" : "var(--space-3) 0 var(--space-1)",
                    fontSize: "var(--text-xs)",
                    fontWeight: "var(--weight-semibold)",
                    color: "var(--color-text-muted)",
                    textTransform: "uppercase",
                    letterSpacing: "0.06em"
                  }}
                >
                  {category}
                </p>
                {options.map((option) => (
                  <div
                    key={option.value}
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: "var(--space-3)",
                      padding: "var(--space-1) 0"
                    }}
                  >
                    <input
                      data-testid={`permission-${option.value}`}
                      type="checkbox"
                      value={option.value}
                      {...form.register("permissions")}
                      style={{ width: 16, height: 16, accentColor: "var(--color-brand-primary)" }}
                    />
                    <span style={{ flex: 1, fontSize: "var(--text-sm)", color: "var(--color-text-primary)" }}>{option.label}</span>
                    <select
                      data-testid={`permission-scope-${option.value}`}
                      disabled={!selectedSet.has(option.value)}
                      onChange={(event) => {
                        setScopeByCode((prev) => ({
                          ...prev,
                          [option.value]: event.target.value as PermissionScope
                        }));
                      }}
                      style={selectStyle}
                      value={scopeByCode[option.value] ?? option.defaultScope}
                    >
                      {SCOPE_OPTIONS.map((scopeOption) => (
                        <option key={scopeOption.value} value={scopeOption.value}>
                          {scopeOption.label}
                        </option>
                      ))}
                    </select>
                  </div>
                ))}
              </div>
            ))}
          </div>
        </section>
      </EntityFormFrame>
    </Form>
  );
}

function getDefaultScopeMap(): Record<string, PermissionScope> {
  return PERMISSION_OPTIONS.reduce<Record<string, PermissionScope>>((acc, option) => {
    acc[option.value] = option.defaultScope;
    return acc;
  }, {});
}

function mapRoleError(code: string | undefined, message: string, kind: string): string {
  if (code === "validation.role_in_use") {
    return "Роль назначена одной или нескольким должностям.";
  }
  if (code?.startsWith("access.scope")) {
    return "Операция недоступна в текущем контексте.";
  }
  if (kind === "forbidden" || code?.startsWith("access.")) {
    return "Операция запрещена.";
  }
  return message;
}

const selectStyle: CSSProperties = {
  width: 96,
  height: 36,
  border: "1px solid var(--color-border)",
  borderRadius: "var(--radius-input)",
  background: "var(--color-surface)",
  padding: "0 var(--space-2)",
  fontSize: "var(--text-xs)",
  color: "var(--color-text-primary)"
};
