# Performance Baseline Runbook

This directory tracks append-only API performance baselines for SC-4/SC-5.

## Prerequisites

- Backend API is running and reachable (default: `http://127.0.0.1:8080`).
- Either:
  - `k6` is installed locally, or
  - Docker is installed (runner uses `grafana/k6` fallback).
- Python 3 with `matplotlib`:
  - `py -3 -m pip install matplotlib`

## Run perf baseline (append one row)

```bash
bash perf/run.sh
```

Each successful run appends exactly one line to `perf/results.csv` with:

- `timestamp_utc`
- `p50_ms`
- `p95_ms`
- `p99_ms`
- `backend_memory_mb`

The script never rewrites or reorders existing rows.

## Generate charts from CSV history

```bash
py -3 perf/plot.py
```

Charts are written to `perf/charts/`:

- `p50_ms.png`
- `p95_ms.png`
- `p99_ms.png`
- `backend_memory_mb.png`

## Environment overrides

- `PERF_BASE_URL`: API base URL for load test.
- `PERF_BACKEND_PORT`: Port used to detect backend process memory (default `8080`).
- `PERF_BACKEND_CONTAINER`: Docker container name to use for memory sampling.

Examples:

```bash
PERF_BASE_URL=http://127.0.0.1:8080 bash perf/run.sh
PERF_BACKEND_CONTAINER=activist-backend bash perf/run.sh
```
