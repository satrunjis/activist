/**
 * SkeletonBlock — generic animated loading placeholder.
 *
 * Design contract: D-04, D-06 from 09-CONTEXT.md
 * - Use for any content area that is awaiting data (list rows, text, avatars).
 * - Never leave a blank-white surface while loading; always render a skeleton instead.
 *
 * Usage:
 *   <SkeletonBlock height={20} width="60%" />
 *   <SkeletonBlock height={40} className="rounded-full" />
 */

import type { CSSProperties } from "react";

export interface SkeletonBlockProps {
  /** Height in pixels or any valid CSS value. Defaults to 16px. */
  height?: number | string;
  /** Width as a number (px) or any CSS value. Defaults to "100%". */
  width?: number | string;
  /** Additional CSS class names. */
  className?: string;
  /** Inline style overrides. */
  style?: CSSProperties;
}

export function SkeletonBlock({ height = 16, width = "100%", className, style }: SkeletonBlockProps) {
  const resolvedHeight = typeof height === "number" ? `${height}px` : height;
  const resolvedWidth = typeof width === "number" ? `${width}px` : width;

  return (
    <div
      aria-hidden="true"
      className={className}
      style={{
        height: resolvedHeight,
        width: resolvedWidth,
        borderRadius: "var(--radius-input)",
        background: "linear-gradient(90deg, #E8EAED 25%, #F4F5F7 50%, #E8EAED 75%)",
        backgroundSize: "200% 100%",
        animation: "shimmer 1.4s ease infinite",
        ...style,
      }}
    />
  );
}
