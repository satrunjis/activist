import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "../../lib/utils";

const buttonVariants = cva(
  "inline-flex shrink-0 items-center justify-center whitespace-nowrap font-body font-medium leading-tight transition-colors duration-120 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-primary/20 disabled:pointer-events-none disabled:opacity-65",
  {
    variants: {
      variant: {
        primary: "bg-brand-primary text-text-inverse border-0 hover:bg-brand-primary-hover rounded-card",
        outline: "bg-transparent border border-border-strong text-text-primary hover:bg-surface-subtle rounded-card",
        ghost: "bg-transparent border-0 text-text-primary hover:bg-surface-subtle rounded-card",
        destructive: "bg-error hover:bg-error text-text-inverse rounded-card border-0"
      },
      size: {
        sm: "h-7 px-3 text-sm",
        default: "h-9 px-3 text-sm",
        lg: "h-11 px-5 text-base"
      }
    },
    defaultVariants: {
      variant: "primary",
      size:    "default"
    }
  }
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
  isLoading?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, type = "button", disabled, isLoading = false, children, ...props }, ref) => {
    const Comp = asChild ? Slot : "button";
    const content = isLoading ? (
      <span className="flex items-center gap-2">
        <span
          aria-hidden="true"
          className="h-3 w-3 animate-spin rounded-full border-2 border-current/30 border-t-current"
        />
        <span>{children}</span>
      </span>
    ) : (
      <span className="flex items-center gap-2">{children}</span>
    );

    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        disabled={disabled || isLoading}
        type={type}
        {...props}
      >
        {content}
      </Comp>
    );
  }
);
Button.displayName = "Button";

export { Button, buttonVariants };
