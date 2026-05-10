import { Trash2 } from "lucide-react";
import { useEffect, useMemo } from "react";
import type { CSSProperties } from "react";
import type { Path } from "react-hook-form";
import { useFieldArray, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

import { Button } from "../../components/ui/button";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "../../components/ui/form";
import { Input } from "../../components/ui/input";
import { adaptApiError, getFieldErrors, getRootMessage } from "../../shared/api/errorAdapter";
import { useToast } from "../../shared/ui/feedback";
import { EntityFormFrame, FormActions } from "../../shared/ui/forms";
import type { DivisionCreateRequest, DivisionMediaLink, DivisionPatchRequest, DivisionTreeNode } from "../../shared/api/types";

type DivisionOption = {
  id: string;
  label: string;
};

type DivisionFormValues = {
  parent_id: string;
  short_name: string;
  full_name: string;
  description: string;
  regulation_url: string;
  media_links: DivisionMediaLink[];
};

type DivisionEditFormProps = {
  mode: "create" | "edit";
  node?: DivisionTreeNode;
  defaultParentId?: string;
  parentOptions: DivisionOption[];
  isSubmitting: boolean;
  onCancel: () => void;
  onCreate: (payload: DivisionCreateRequest) => Promise<void>;
  onEdit: (payload: DivisionPatchRequest) => Promise<void>;
};

const divisionFormSchema = z.object({
  parent_id: z.string(),
  short_name: z.string().trim().min(1, "Укажите краткое название."),
  full_name: z.string().trim().min(1, "Укажите полное название."),
  description: z.string(),
  regulation_url: z.string(),
  media_links: z.array(
    z.object({
      platform: z.string().trim().min(1, "Укажите платформу."),
      value: z.string().trim().min(1, "Укажите ссылку.")
    })
  )
});

function createDivisionFormSchema(mode: "create" | "edit") {
  return divisionFormSchema.superRefine((value, context) => {
    if (mode === "create" && !value.parent_id.trim()) {
      context.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Выберите родительское подразделение.",
        path: ["parent_id"]
      });
    }
  });
}

function trimmed(value: string): string {
  return value.trim();
}

function normalizeLinks(links: DivisionMediaLink[]): DivisionMediaLink[] {
  return links.map((item) => ({
    platform: trimmed(item.platform),
    value: trimmed(item.value)
  }));
}

function defaultValues(mode: "create" | "edit", node: DivisionTreeNode | undefined, defaultParentId: string | undefined): DivisionFormValues {
  return {
    parent_id: mode === "edit" ? node?.parent_id ?? "" : defaultParentId ?? "",
    short_name: node?.short_name ?? "",
    full_name: node?.full_name ?? "",
    description: node?.description ?? "",
    regulation_url: node?.regulation_url ?? "",
    media_links: node?.media_links ?? []
  };
}

export function DivisionEditForm({
  mode,
  node,
  defaultParentId,
  parentOptions,
  isSubmitting,
  onCancel,
  onCreate,
  onEdit
}: DivisionEditFormProps) {
  const { showToast } = useToast();
  const defaults = useMemo(
    () => defaultValues(mode, node, defaultParentId),
    [defaultParentId, mode, node]
  );
  const schema = useMemo(() => createDivisionFormSchema(mode), [mode]);

  const form = useForm<DivisionFormValues>({
    resolver: zodResolver(schema),
    defaultValues: defaults
  });

  const { fields, append, remove } = useFieldArray({
    control: form.control,
    name: "media_links"
  });

  useEffect(() => {
    form.reset(defaults);
  }, [defaults, form]);

  async function submit(values: DivisionFormValues) {
    form.clearErrors("root");

    const fullName = trimmed(values.full_name);
    const shortName = trimmed(values.short_name);
    const description = trimmed(values.description);
    const regulationURL = trimmed(values.regulation_url);
    const parentID = trimmed(values.parent_id);
    const mediaLinks = normalizeLinks(values.media_links ?? []);

    try {
      if (mode === "create") {
        const payload: DivisionCreateRequest = {
          parent_id: parentID,
          short_name: shortName,
          full_name: fullName,
          description,
          regulation_url: regulationURL,
          media_links: mediaLinks
        };
        await onCreate(payload);
        showToast("success", "Подразделение создано");
        return;
      }

      const payload: DivisionPatchRequest = {
        short_name: shortName,
        full_name: fullName,
        description,
        regulation_url: regulationURL,
        media_links: mediaLinks
      };
      if (node?.id !== "root" && parentID && parentID !== (node?.parent_id ?? "")) {
        payload.parent_id = parentID;
      }
      await onEdit(payload);
      showToast("success", "Изменения сохранены");
    } catch (error) {
      const apiError = adaptApiError(error, "Не удалось сохранить подразделение.");
      const fieldErrors = getFieldErrors(apiError);
      let hasInlineFieldError = false;
      for (const [field, message] of Object.entries(fieldErrors)) {
        form.setError(field as Path<DivisionFormValues>, {
          type: "server",
          message
        });
        hasInlineFieldError = true;
      }
      if (hasInlineFieldError) {
        return;
      }
      form.setError("root", {
        type: "server",
        message: getRootMessage(apiError)
      });
    }
  }

  const fieldBaseStyle: CSSProperties = {
    width: "100%",
    height: 36,
    border: "1px solid var(--color-border)",
    borderRadius: "var(--radius-input)",
    background: "var(--color-surface)",
    padding: "0 var(--space-3)",
    fontSize: "var(--text-sm)",
    color: "var(--color-text-primary)"
  };

  return (
    <Form {...form}>
      <EntityFormFrame
        actions={(
          <FormActions
            cancelLabel="Отмена"
            isSubmitting={isSubmitting}
            onCancel={onCancel}
            submitLabel={mode === "create" ? "Создать" : "Сохранить"}
            submittingLabel={mode === "create" ? "Создание…" : "Сохранение…"}
          />
        )}
        formId={mode === "create" ? "create-division-form" : "edit-division-form"}
        onSubmit={form.handleSubmit(submit)}
        rootError={form.formState.errors.root?.message}
      >
        <div style={{ display: "grid", gap: "var(--space-3)" }}>
          <FormField
            control={form.control}
            name="full_name"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Полное название</FormLabel>
                <FormControl>
                  <Input data-testid={mode === "create" ? "create-division-full-name" : "edit-division-full-name"} {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="short_name"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Краткое название</FormLabel>
                <FormControl>
                  <Input data-testid={mode === "create" ? "create-division-short-name" : "edit-division-short-name"} {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="description"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Описание</FormLabel>
                <FormControl>
                  <textarea
                    data-testid={mode === "create" ? "create-division-description" : "edit-division-description"}
                    name={field.name}
                    onBlur={field.onBlur}
                    onChange={field.onChange}
                    rows={3}
                    style={{
                      ...fieldBaseStyle,
                      resize: "vertical",
                      height: "auto",
                      minHeight: 76,
                      paddingTop: "var(--space-2)",
                      paddingBottom: "var(--space-2)"
                    }}
                    value={field.value}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="parent_id"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Родительское подразделение</FormLabel>
                <FormControl>
                  <select
                    data-testid={mode === "create" ? "create-division-parent" : "edit-division-parent"}
                    disabled={mode === "edit" && node?.id === "root"}
                    name={field.name}
                    onBlur={field.onBlur}
                    onChange={field.onChange}
                    style={fieldBaseStyle}
                    value={field.value}
                  >
                    <option value="">Выберите подразделение</option>
                    {parentOptions.map((option) => (
                      <option key={option.id} value={option.id}>
                        {option.label}
                      </option>
                    ))}
                  </select>
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />

          <FormField
            control={form.control}
            name="regulation_url"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Ссылка на положение</FormLabel>
                <FormControl>
                  <Input data-testid={mode === "create" ? "create-division-regulation-url" : "edit-division-regulation-url"} {...field} />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </div>

        <div style={{ marginTop: "var(--space-4)" }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "var(--space-2)" }}>
            <span style={{ fontSize: "var(--text-xs)", fontWeight: "var(--weight-semibold)", color: "var(--color-text-muted)", textTransform: "uppercase", letterSpacing: "0.06em" }}>
              Медиа-ссылки
            </span>
            <Button
              onClick={() => append({ platform: "", value: "" })}
              size="sm"
              type="button"
              variant="outline"
            >
              + Ссылка
            </Button>
          </div>

          {!fields.length ? (
            <p style={{ fontSize: "var(--text-sm)", color: "var(--color-text-muted)" }}>
              Ссылки не добавлены.
            </p>
          ) : null}

          <div style={{ display: "grid", gap: "var(--space-2)" }}>
            {fields.map((item, index) => (
              <div
                key={item.id}
                style={{ display: "grid", gridTemplateColumns: "minmax(0,1fr) minmax(0,1fr) auto", gap: "var(--space-2)", alignItems: "start" }}
              >
                <FormField
                  control={form.control}
                  name={`media_links.${index}.platform`}
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Платформа</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <FormField
                  control={form.control}
                  name={`media_links.${index}.value`}
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Ссылка</FormLabel>
                      <FormControl>
                        <Input {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                <Button
                  aria-label="Удалить ссылку"
                  className="w-9 px-0"
                  onClick={() => remove(index)}
                  style={{ marginTop: 26 }}
                  type="button"
                  variant="ghost"
                >
                  <Trash2 size={14} />
                </Button>
              </div>
            ))}
          </div>
        </div>
      </EntityFormFrame>
    </Form>
  );
}
