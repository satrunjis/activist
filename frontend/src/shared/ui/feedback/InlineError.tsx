export type InlineErrorProps = {
  message: string | null | undefined;
  "data-testid"?: string;
};

export function InlineError({ message, "data-testid": dataTestId }: InlineErrorProps) {
  if (!message) {
    return null;
  }

  return (
    <p
      data-testid={dataTestId}
      style={{
        margin: 0,
        marginTop: "var(--space-1)",
        color: "var(--color-error-text)",
        fontSize: "var(--text-sm)"
      }}
    >
      {message}
    </p>
  );
}
