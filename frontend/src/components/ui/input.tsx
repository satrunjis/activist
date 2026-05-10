import * as React from "react";

import { cn } from "../../lib/utils";

const Input = React.forwardRef<HTMLInputElement, React.ComponentProps<"input">>(
  ({ className, type = "text", ...props }, ref) => {
    return (
      <input
        className={cn(
          "w-full h-9 px-3 py-2 text-sm font-body bg-surface border border-border rounded-input text-text-primary placeholder:text-text-muted transition-colors duration-120 focus:outline-none focus:border-border-focus focus:ring-2 focus:ring-brand-primary/20 disabled:bg-surface-subtle disabled:text-text-disabled disabled:cursor-not-allowed disabled:opacity-65 aria-invalid:border-border-error aria-invalid:ring-2 aria-invalid:ring-error/15 file:border-0 file:bg-transparent file:text-sm file:font-medium",
          className
        )}
        ref={ref}
        type={type}
        {...props}
      />
    );
  }
);
Input.displayName = "Input";

export { Input };
