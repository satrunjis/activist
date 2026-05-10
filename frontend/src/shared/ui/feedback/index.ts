/**
 * Shared feedback primitives for loading and empty states.
 *
 * Design contract: D-04, D-06, D-11, D-12, D-13 (09-CONTEXT.md)
 */

export { SkeletonBlock } from "./SkeletonBlock";
export type { SkeletonBlockProps } from "./SkeletonBlock";

export { SectionSkeleton } from "./SectionSkeleton";
export type { SectionSkeletonProps } from "./SectionSkeleton";

export { SoftRefreshWrapper } from "./SoftRefreshWrapper";
export type { SoftRefreshWrapperProps } from "./SoftRefreshWrapper";

export { EmptyStateCard } from "./EmptyStateCard";
export type { EmptyStateCardProps } from "./EmptyStateCard";

export { InlineError } from "./InlineError";
export type { InlineErrorProps } from "./InlineError";

export { GlobalApiBanner } from "./GlobalApiBanner";
export type { GlobalApiBannerProps } from "./GlobalApiBanner";

export { ToastProvider, useToast } from "./ToastProvider";
export type { ToastVariant } from "./ToastProvider";

export { ConfirmDialog } from "./ConfirmDialog";
export type { ConfirmDialogProps } from "./ConfirmDialog";
