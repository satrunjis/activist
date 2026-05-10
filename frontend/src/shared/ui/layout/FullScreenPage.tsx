export type FullScreenPageProps = {
  children: React.ReactNode;
  "data-testid"?: string;
};

export function FullScreenPage({ children, "data-testid": dataTestId }: FullScreenPageProps) {
  return (
    <div
      data-testid={dataTestId}
      style={{
        height: "calc(100vh - var(--topbar-height))",
        minHeight: 0,
        display: "flex",
        overflow: "hidden"
      }}
    >
      {children}
    </div>
  );
}
