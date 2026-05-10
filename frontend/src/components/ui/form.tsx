import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import {
  Controller,
  FormProvider,
  useFormContext,
  type ControllerProps,
  type FieldPath,
  type FieldValues
} from "react-hook-form";

import { cn } from "../../lib/utils";
import { Label } from "./label";

const Form = FormProvider;

type FormFieldContextValue<
  TFieldValues extends FieldValues = FieldValues,
  TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>
> = {
  name: TName;
};

const FormFieldContext = React.createContext<FormFieldContextValue>({} as FormFieldContextValue);

function FormField<TFieldValues extends FieldValues, TName extends FieldPath<TFieldValues>>(
  props: ControllerProps<TFieldValues, TName>
) {
  return (
    <FormFieldContext.Provider value={{ name: props.name }}>
      <Controller {...props} />
    </FormFieldContext.Provider>
  );
}

type FormItemContextValue = {
  id: string;
};

const FormItemContext = React.createContext<FormItemContextValue>({} as FormItemContextValue);

const FormItem = React.forwardRef<HTMLDivElement, React.HTMLAttributes<HTMLDivElement>>(
  ({ className, ...props }, ref) => {
    const id = React.useId();
    return (
      <FormItemContext.Provider value={{ id }}>
        <div ref={ref} className={cn("space-y-0", className)} {...props} />
      </FormItemContext.Provider>
    );
  }
);
FormItem.displayName = "FormItem";

type FormLabelProps = React.ComponentPropsWithoutRef<typeof Label> & {
  required?: boolean;
};

const FormLabel = React.forwardRef<React.ElementRef<typeof Label>, FormLabelProps>(
  ({ className, children, required = false, ...props }, ref) => {
    const { formItemId } = useFormField();
    return (
      <Label
        ref={ref}
        className={cn(
          "block [margin-bottom:var(--space-1)] [font-size:var(--text-sm)] [font-weight:var(--weight-semibold)] [color:var(--color-text-primary)]",
          className
        )}
        htmlFor={formItemId}
        {...props}
      >
        {children}
        {required ? <span aria-hidden="true" className="[color:var(--color-error)]"> *</span> : null}
      </Label>
    );
  }
);
FormLabel.displayName = "FormLabel";

const FormControl = React.forwardRef<React.ElementRef<typeof Slot>, React.ComponentPropsWithoutRef<typeof Slot>>(
  ({ ...props }, ref) => {
    const { error, formItemId, formDescriptionId, formMessageId } = useFormField();
    return (
      <Slot
        ref={ref}
        id={formItemId}
        aria-describedby={error ? `${formDescriptionId} ${formMessageId}` : formDescriptionId}
        aria-invalid={Boolean(error)}
        {...props}
      />
    );
  }
);
FormControl.displayName = "FormControl";

const FormDescription = React.forwardRef<HTMLParagraphElement, React.HTMLAttributes<HTMLParagraphElement>>(
  ({ className, ...props }, ref) => {
    const { formDescriptionId } = useFormField();
    return (
      <p
        ref={ref}
        id={formDescriptionId}
        className={cn("mt-[4px] [font-size:var(--text-sm)] [color:var(--color-text-muted)]", className)}
        {...props}
      />
    );
  }
);
FormDescription.displayName = "FormDescription";

const FormMessage = React.forwardRef<HTMLParagraphElement, React.HTMLAttributes<HTMLParagraphElement>>(
  ({ className, children, ...props }, ref) => {
    const { error, formMessageId } = useFormField();
    const body = error ? String(error.message ?? "") : children;
    if (!body) {
      return null;
    }
    return (
      <p
        ref={ref}
        id={formMessageId}
        className={cn("mt-[4px] [font-size:var(--text-xs)] [font-weight:var(--weight-medium)] [color:var(--color-error-text)]", className)}
        {...props}
      >
        {body}
      </p>
    );
  }
);
FormMessage.displayName = "FormMessage";

type RootErrorProps = React.HTMLAttributes<HTMLDivElement> & {
  message?: React.ReactNode;
};

const RootError = React.forwardRef<HTMLDivElement, RootErrorProps>(
  ({ className, message, children, ...props }, ref) => {
    const body = message ?? children;
    if (!body) {
      return null;
    }
    return (
      <div
        ref={ref}
        className={cn(
          "[background:var(--color-error-subtle)] [border:1px_solid_var(--color-border-error)] [border-radius:var(--radius-input)] px-[var(--space-3)] py-[var(--space-2)] [font-size:var(--text-sm)] [color:var(--color-error-text)]",
          className
        )}
        role="alert"
        {...props}
      >
        {body}
      </div>
    );
  }
);
RootError.displayName = "RootError";

function useFormField() {
  const fieldContext = React.useContext(FormFieldContext);
  const itemContext = React.useContext(FormItemContext);
  const { getFieldState, formState } = useFormContext();
  const fieldState = getFieldState(fieldContext.name, formState);

  if (!fieldContext) {
    throw new Error("useFormField must be used inside <FormField>.");
  }

  const { id } = itemContext;

  return {
    id,
    name: fieldContext.name,
    formItemId: `${id}-form-item`,
    formDescriptionId: `${id}-form-item-description`,
    formMessageId: `${id}-form-item-message`,
    ...fieldState
  };
}

export { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage, RootError, useFormField };
