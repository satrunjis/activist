import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "../../components/ui/form";
import { Input } from "../../components/ui/input";
import { request } from "../../shared/api/client";
import { adaptApiError } from "../../shared/api/errorAdapter";
import { EntityFormFrame, FormActions } from "../../shared/ui/forms";
import type { AssignMemberRequest } from "../../shared/api/types";

const schema = z.object({
  user_id: z.string().trim().min(1, "Укажите ID пользователя.")
});

type FormValues = z.infer<typeof schema>;

type AssignMemberFormProps = {
  positionId: string;
  positionTitle: string;
  onSuccess: () => void;
  onCancel: () => void;
  onForbidden?: (message: string) => void;
};

const FORBIDDEN_SCOPE_MESSAGE = "Операция недоступна в текущем контексте.";

export function AssignMemberForm({
  positionId,
  positionTitle,
  onSuccess,
  onCancel,
  onForbidden
}: AssignMemberFormProps) {
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      user_id: ""
    }
  });

  async function submit(values: FormValues) {
    form.clearErrors("root");

    const payload: AssignMemberRequest = {
      user_id: values.user_id.trim(),
      position_id: positionId
    };

    try {
      await request("/api/v1/memberships", {
        method: "POST",
        body: payload
      });
      onSuccess();
    } catch (error) {
      const apiError = adaptApiError(error, "Не удалось назначить участника.");
      if (apiError.kind === "conflict") {
        form.setError("root", {
          type: "server",
          message: "Пользователь уже назначен на эту должность."
        });
        return;
      }
      if (apiError.kind === "forbidden") {
        const message = FORBIDDEN_SCOPE_MESSAGE;
        form.setError("root", {
          type: "server",
          message
        });
        onForbidden?.(message);
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
            submitLabel="Назначить"
            submittingLabel="Назначение…"
          />
        )}
        formId="assign-member-form"
        onSubmit={form.handleSubmit(submit)}
        rootError={form.formState.errors.root?.message}
      >
        <p className="muted-text">
          Должность: <code>{positionTitle}</code>
        </p>
        <FormField
          control={form.control}
          name="user_id"
          render={({ field }) => (
            <FormItem>
              <FormLabel>ID пользователя</FormLabel>
              <FormControl>
                <Input data-testid="assign-user-id" placeholder="Введите ID пользователя" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
      </EntityFormFrame>
    </Form>
  );
}
