import { Button } from "../../../components/ui/button";
import { useEntityFormFrameContext } from "./EntityFormFrame";

type FormActionsProps = {
  submitLabel: string;
  submittingLabel?: string;
  cancelLabel?: string;
  onCancel?: () => void;
  isSubmitting?: boolean;
  disableSubmit?: boolean;
};

export function FormActions({
  submitLabel,
  submittingLabel,
  cancelLabel = "Отмена",
  onCancel,
  isSubmitting = false,
  disableSubmit = false
}: FormActionsProps) {
  const onlyPrimary = !onCancel;
  const { formId } = useEntityFormFrameContext();
  const submitTestId = mapSubmitTestId(formId);

  return (
    <div
      style={
        onlyPrimary
          ? { display: "flex", marginTop: "var(--space-4)", width: "100%" }
          : { display: "flex", justifyContent: "flex-end", gap: "var(--space-2)", marginTop: "var(--space-4)" }
      }
    >
      {onCancel ? (
        <Button disabled={isSubmitting} onClick={onCancel} type="button" variant="outline">
          {cancelLabel}
        </Button>
      ) : null}
      <Button
        data-testid={submitTestId}
        disabled={disableSubmit || isSubmitting}
        isLoading={isSubmitting}
        style={onlyPrimary ? { width: "100%" } : undefined}
        type="submit"
        variant="primary"
      >
        {isSubmitting ? submittingLabel ?? submitLabel : submitLabel}
      </Button>
    </div>
  );
}

function mapSubmitTestId(formId: string | undefined): string | undefined {
  if (!formId) {
    return undefined;
  }

  if (formId === "create-role-form" || formId === "edit-role-form") {
    return "role-submit";
  }
  if (formId === "create-division-form") {
    return "create-division-submit";
  }
  if (formId === "edit-division-form") {
    return "edit-division-submit";
  }
  if (formId === "create-position-form") {
    return "create-position-submit";
  }
  if (formId === "assign-member-form") {
    return "assign-submit";
  }
  return undefined;
}

export type { FormActionsProps };
