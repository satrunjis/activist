import { AlertCircle } from "lucide-react";

import { EmptyStateCard, SectionSkeleton } from "../feedback";

export type AsyncState = "idle" | "loading" | "error" | "empty" | "success";

export type AsyncStateViewProps<TData> = {
  state: AsyncState;
  data?: TData;
  errorMessage?: string | null;
  onRetry?: () => void;
  loadingView?: React.ReactNode;
  emptyView?: React.ReactNode;
  idleView?: React.ReactNode;
  children: (data: TData) => React.ReactNode; // success renderer
};

export function AsyncStateView<TData>({
  state,
  data,
  errorMessage,
  onRetry,
  loadingView,
  emptyView,
  idleView,
  children
}: AsyncStateViewProps<TData>) {
  if (state === "idle") {
    return <>{idleView ?? null}</>;
  }

  if (state === "loading") {
    return <>{loadingView ?? <SectionSkeleton rows={5} />}</>;
  }

  if (state === "error") {
    return (
      <EmptyStateCard
        body={errorMessage ?? "Не удалось загрузить данные."}
        cta={onRetry ? { label: "Повторить", onClick: onRetry } : undefined}
        heading="Ошибка загрузки"
        icon={<AlertCircle />}
      />
    );
  }

  if (state === "empty") {
    return <>{emptyView ?? <EmptyStateCard />}</>;
  }

  return <>{children(data as TData)}</>;
}
