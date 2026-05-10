import "./MetaBlock.css";

import type { ReactNode } from "react";

export type MetaField = {
  label: ReactNode;
  value: ReactNode;
};

export type MetaBlockProps = {
  fields: MetaField[];
  columns?: 1 | 2;
  "data-testid"?: string;
};

export function MetaBlock({ fields, columns = 1, "data-testid": testId }: MetaBlockProps) {
  return (
    <dl
      className={`meta-block meta-block--columns-${columns}`}
      data-testid={testId}
    >
      {fields.map((field, index) => (
        <div className="meta-block__field" key={index}>
          <dt className="meta-block__label">{field.label}</dt>
          <dd className="meta-block__value">{field.value}</dd>
        </div>
      ))}
    </dl>
  );
}
