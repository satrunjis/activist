#!/usr/bin/env python3
from __future__ import annotations

import csv
from dataclasses import dataclass
from html import escape
from pathlib import Path

try:
    import matplotlib

    matplotlib.use("Agg")
    import matplotlib.pyplot as plt
except ImportError as exc:
    raise SystemExit(
        "matplotlib is required. Install with: py -3 -m pip install matplotlib"
    ) from exc


@dataclass(frozen=True)
class Sample:
    scenario_name: str
    profile_label: str
    vus: float
    iterations: float
    req_count: float
    rps: float
    error_rate: float
    avg_ms: float
    p95_ms: float
    p99_ms: float
    memory_avg_mb: float
    cpu_avg_pct: float


def load_summary_rows(summary_csv: Path) -> tuple[list[dict[str, str]], list[str]]:
    with summary_csv.open("r", encoding="utf-8", newline="") as handle:
        reader = csv.DictReader(handle)
        if reader.fieldnames is None:
            raise SystemExit(f"CSV has no header row: {summary_csv}")
        rows = list(reader)
        return rows, list(reader.fieldnames)


def load_samples(summary_csv: Path) -> list[Sample]:
    if not summary_csv.exists():
        raise SystemExit(f"Missing current summary: {summary_csv}. Run `SCENARIO=full bash perf/run.sh` first.")

    expected = {
        "scenario_name",
        "profile_label",
        "vus",
        "iterations",
        "req_count",
        "rps",
        "error_rate",
        "avg_ms",
        "p95_ms",
        "p99_ms",
        "memory_avg_mb",
        "cpu_avg_pct",
    }

    rows, fieldnames = load_summary_rows(summary_csv)
    missing = expected.difference(fieldnames)
    if missing:
        raise SystemExit(f"CSV missing required columns: {sorted(missing)}")

    samples: list[Sample] = []
    for row_index, row in enumerate(rows, start=2):
        try:
            samples.append(
                Sample(
                    scenario_name=row["scenario_name"],
                    profile_label=row["profile_label"],
                    vus=float(row["vus"]),
                    iterations=float(row["iterations"]),
                    req_count=float(row["req_count"]),
                    rps=float(row["rps"]),
                    error_rate=float(row["error_rate"]),
                    avg_ms=float(row["avg_ms"]),
                    p95_ms=float(row["p95_ms"]),
                    p99_ms=float(row["p99_ms"]),
                    memory_avg_mb=float(row["memory_avg_mb"]),
                    cpu_avg_pct=float(row["cpu_avg_pct"]),
                )
            )
        except Exception as exc:
            raise SystemExit(f"Invalid data at {summary_csv}:{row_index}: {exc}") from exc

    if not samples:
        raise SystemExit("Current summary is empty. Run `SCENARIO=full bash perf/run.sh` first.")

    return samples


def grouped(samples: list[Sample]) -> dict[str, list[Sample]]:
    buckets: dict[str, list[Sample]] = {}
    for sample in samples:
        buckets.setdefault(sample.scenario_name, []).append(sample)
    for key in buckets:
        buckets[key] = sorted(buckets[key], key=lambda s: s.vus)
    return buckets


def render_line_by_vus(
    samples: list[Sample],
    y_key: str,
    title: str,
    ylabel: str,
    output_path: Path,
) -> None:
    fig, axis = plt.subplots(figsize=(10, 5))
    palette = ["#2563eb", "#dc2626", "#059669", "#7c3aed", "#ea580c"]

    for idx, (scenario_name, bucket) in enumerate(sorted(grouped(samples).items())):
        axis.plot(
            [s.vus for s in bucket],
            [getattr(s, y_key) for s in bucket],
            marker="o",
            linewidth=2,
            label=scenario_name,
            color=palette[idx % len(palette)],
        )

    axis.set_title(title)
    axis.set_xlabel("VUs")
    axis.set_ylabel(ylabel)
    axis.grid(alpha=0.35, linestyle="--", linewidth=0.7)
    axis.legend(loc="best")
    fig.tight_layout()
    fig.savefig(output_path, dpi=160)
    plt.close(fig)


def render_scatter(
    samples: list[Sample],
    x_key: str,
    y_key: str,
    title: str,
    xlabel: str,
    ylabel: str,
    output_path: Path,
) -> None:
    fig, axis = plt.subplots(figsize=(10, 5))
    palette = ["#2563eb", "#dc2626", "#059669", "#7c3aed", "#ea580c"]

    for idx, (scenario_name, bucket) in enumerate(sorted(grouped(samples).items())):
        axis.scatter(
            [getattr(s, x_key) for s in bucket],
            [getattr(s, y_key) for s in bucket],
            label=scenario_name,
            s=70,
            alpha=0.85,
            color=palette[idx % len(palette)],
        )

    axis.set_title(title)
    axis.set_xlabel(xlabel)
    axis.set_ylabel(ylabel)
    axis.grid(alpha=0.35, linestyle="--", linewidth=0.7)
    axis.legend(loc="best")
    fig.tight_layout()
    fig.savefig(output_path, dpi=160)
    plt.close(fig)


def write_semicolon_csv(rows: list[dict[str, str]], fieldnames: list[str], output_path: Path) -> None:
    with output_path.open("w", encoding="utf-8", newline="") as handle:
        writer = csv.DictWriter(handle, fieldnames=fieldnames, delimiter=";")
        writer.writeheader()
        writer.writerows(rows)


def safe_float(raw: str) -> float | None:
    try:
        return float(raw)
    except Exception:
        return None


def lerp(a: int, b: int, t: float) -> int:
    return int(round(a + (b - a) * t))


def gradient_color(t: float, low_rgb: tuple[int, int, int], high_rgb: tuple[int, int, int]) -> str:
    bounded = max(0.0, min(1.0, t))
    r = lerp(low_rgb[0], high_rgb[0], bounded)
    g = lerp(low_rgb[1], high_rgb[1], bounded)
    b = lerp(low_rgb[2], high_rgb[2], bounded)
    return f"rgb({r}, {g}, {b})"


def metric_cell_style(
    value: str,
    key: str,
    mins: dict[str, float],
    maxs: dict[str, float],
    higher_is_better: set[str],
) -> str:
    numeric = safe_float(value)
    if numeric is None or key not in mins or key not in maxs:
        return ""

    min_v = mins[key]
    max_v = maxs[key]
    if max_v <= min_v:
        ratio = 0.5
    else:
        ratio = (numeric - min_v) / (max_v - min_v)
    if key in higher_is_better:
        ratio = 1.0 - ratio

    color = gradient_color(ratio, (232, 245, 233), (255, 235, 238))
    return f' style="background:{color};"'


def render_html_report(rows: list[dict[str, str]], fieldnames: list[str], output_path: Path) -> None:
    numeric_keys = [
        "vus",
        "iterations",
        "req_count",
        "rps",
        "error_rate",
        "checks_rate",
        "avg_ms",
        "p50_ms",
        "p95_ms",
        "p99_ms",
        "max_ms",
        "memory_avg_mb",
        "cpu_avg_pct",
    ]
    higher_is_better = {"rps", "checks_rate"}

    mins: dict[str, float] = {}
    maxs: dict[str, float] = {}
    for key in numeric_keys:
        values = [safe_float(row.get(key, "")) for row in rows]
        filtered = [value for value in values if value is not None]
        if filtered:
            mins[key] = min(filtered)
            maxs[key] = max(filtered)

    visible_columns = [
        "scenario_name",
        "profile_label",
        "vus",
        "iterations",
        "req_count",
        "rps",
        "error_rate",
        "checks_rate",
        "avg_ms",
        "p50_ms",
        "p95_ms",
        "p99_ms",
        "max_ms",
        "memory_avg_mb",
        "cpu_avg_pct",
        "raw_summary_json",
    ]
    columns = [column for column in visible_columns if column in fieldnames]

    top_row = max(rows, key=lambda row: safe_float(row.get("rps", "0")) or 0.0) if rows else None
    worst_row = max(rows, key=lambda row: safe_float(row.get("p95_ms", "0")) or 0.0) if rows else None

    cards = []
    if top_row is not None:
        cards.append(
            f'<div class="card"><div class="k">Max RPS</div><div class="v">{escape(top_row.get("rps",""))}</div>'
            f'<div class="s">{escape(top_row.get("profile_label",""))}</div></div>'
        )
    if worst_row is not None:
        cards.append(
            f'<div class="card"><div class="k">Max p95</div><div class="v">{escape(worst_row.get("p95_ms",""))} ms</div>'
            f'<div class="s">{escape(worst_row.get("profile_label",""))}</div></div>'
        )
    cards.append(f'<div class="card"><div class="k">Rows</div><div class="v">{len(rows)}</div></div>')

    header_html = "".join(f"<th>{escape(column)}</th>" for column in columns)
    body_parts: list[str] = []
    for row in rows:
        cells: list[str] = []
        for column in columns:
            value = row.get(column, "")
            style = metric_cell_style(value, column, mins, maxs, higher_is_better)
            if column == "raw_summary_json":
                cell_content = f'<span class="path">{escape(value)}</span>'
            else:
                cell_content = escape(value)
            cells.append(f"<td{style}>{cell_content}</td>")
        body_parts.append("<tr>" + "".join(cells) + "</tr>")

    html = f"""<!doctype html>
<html lang="ru">
<head>
  <meta charset="utf-8" />
  <title>Perf Snapshot Report</title>
  <style>
    :root {{
      --fg: #1f2937;
      --panel: #ffffff;
      --line: #dbe3ee;
      --accent: #1d4ed8;
    }}
    body {{
      margin: 0;
      font: 14px/1.4 "Segoe UI", "Inter", sans-serif;
      color: var(--fg);
      background: linear-gradient(140deg, #eef4ff 0%, #f8fafc 45%, #ecfeff 100%);
    }}
    .wrap {{ max-width: 1900px; margin: 24px auto; padding: 0 16px; }}
    h1 {{ margin: 0 0 8px; font-size: 24px; }}
    .sub {{ color: #475569; margin-bottom: 16px; }}
    .cards {{ display: flex; gap: 10px; margin-bottom: 14px; flex-wrap: wrap; }}
    .card {{
      background: var(--panel);
      border: 1px solid var(--line);
      border-radius: 10px;
      padding: 10px 12px;
      min-width: 170px;
      box-shadow: 0 2px 10px rgba(30, 41, 59, 0.05);
    }}
    .card .k {{ color: #64748b; font-size: 12px; }}
    .card .v {{ font-size: 22px; font-weight: 700; color: var(--accent); }}
    .card .s {{ color: #334155; font-size: 12px; }}
    .table-box {{
      border: 1px solid var(--line);
      border-radius: 12px;
      overflow: auto;
      background: var(--panel);
      box-shadow: 0 4px 14px rgba(15, 23, 42, 0.06);
    }}
    table {{ width: max-content; min-width: 100%; border-collapse: collapse; }}
    thead th {{
      position: sticky;
      top: 0;
      background: #eaf0fb;
      color: #0f172a;
      border-bottom: 1px solid var(--line);
      z-index: 1;
    }}
    th, td {{
      border-right: 1px solid var(--line);
      border-bottom: 1px solid var(--line);
      padding: 7px 10px;
      white-space: nowrap;
      text-align: right;
      font-variant-numeric: tabular-nums;
    }}
    th:first-child, td:first-child {{ text-align: left; }}
    th:nth-child(2), td:nth-child(2) {{ text-align: left; }}
    tr:nth-child(even) td {{ background-color: rgba(241, 245, 249, 0.45); }}
    .path {{
      display: inline-block;
      max-width: 560px;
      overflow: hidden;
      text-overflow: ellipsis;
      color: #334155;
    }}
  </style>
</head>
<body>
  <div class="wrap">
    <h1>Perf Full Snapshot</h1>
    <div class="sub">Источник: summary.csv, раскладка по колонкам + цветовая подсветка ключевых метрик.</div>
    <div class="cards">{"".join(cards)}</div>
    <div class="table-box">
      <table>
        <thead><tr>{header_html}</tr></thead>
        <tbody>{"".join(body_parts)}</tbody>
      </table>
    </div>
  </div>
</body>
</html>
"""
    output_path.write_text(html, encoding="utf-8")


def main() -> None:
    script_dir = Path(__file__).resolve().parent
    current_dir = script_dir / "results" / "current"
    summary_csv = current_dir / "summary.csv"
    charts_dir = script_dir / "charts"
    charts_dir.mkdir(parents=True, exist_ok=True)

    rows, fieldnames = load_summary_rows(summary_csv)
    samples = load_samples(summary_csv)

    generated: list[Path] = []

    rps_vs_vus = charts_dir / "rps_vs_vus.png"
    render_line_by_vus(samples, "rps", "Throughput vs load", "RPS", rps_vs_vus)
    generated.append(rps_vs_vus)

    p95_vs_vus = charts_dir / "p95_vs_vus.png"
    render_line_by_vus(samples, "p95_ms", "Latency vs load", "p95 latency (ms)", p95_vs_vus)
    generated.append(p95_vs_vus)

    error_vs_vus = charts_dir / "error_rate_vs_vus.png"
    render_line_by_vus(samples, "error_rate", "Error rate vs load", "Error rate", error_vs_vus)
    generated.append(error_vs_vus)

    memory_vs_rps = charts_dir / "memory_vs_rps.png"
    render_scatter(
        samples,
        "rps",
        "memory_avg_mb",
        "Resource dependency: memory vs throughput",
        "RPS",
        "Memory avg (MB)",
        memory_vs_rps,
    )
    generated.append(memory_vs_rps)

    cpu_vs_rps = charts_dir / "cpu_vs_rps.png"
    render_scatter(
        samples,
        "rps",
        "cpu_avg_pct",
        "Resource dependency: CPU vs throughput",
        "RPS",
        "CPU avg (%)",
        cpu_vs_rps,
    )
    generated.append(cpu_vs_rps)

    p95_vs_cpu = charts_dir / "p95_vs_cpu.png"
    render_scatter(
        samples,
        "cpu_avg_pct",
        "p95_ms",
        "Latency dependency: p95 vs CPU",
        "CPU avg (%)",
        "p95 latency (ms)",
        p95_vs_cpu,
    )
    generated.append(p95_vs_cpu)

    semicolon_csv = current_dir / "summary_semicolon.csv"
    write_semicolon_csv(rows, fieldnames, semicolon_csv)

    html_report = current_dir / "report.html"
    render_html_report(rows, fieldnames, html_report)

    print("Generated charts:")
    for path in generated:
        print(f"- {path}")
    print("Generated table reports:")
    print(f"- {semicolon_csv}")
    print(f"- {html_report}")
    print("Data source:")
    print(f"- {summary_csv}")


if __name__ == "__main__":
    main()
