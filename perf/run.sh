#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_URL="${PERF_BASE_URL:-http://127.0.0.1:8080}"
BACKEND_PORT="${PERF_BACKEND_PORT:-8080}"
BACKEND_CONTAINER="${PERF_BACKEND_CONTAINER:-}"
SCENARIO="${SCENARIO:-full}"
PERF_PROFILE="${PERF_PROFILE:-default}"
CURRENT_DIR="${SCRIPT_DIR}/results/current"
RAW_DIR="${CURRENT_DIR}/raw"
SUMMARY_CSV="${CURRENT_DIR}/summary.csv"
META_JSON="${CURRENT_DIR}/run_meta.json"
K6_BIN=""

run_python() {
  if command -v python3 >/dev/null 2>&1; then
    python3 "$@"
    return
  fi
  if command -v python >/dev/null 2>&1; then
    python "$@"
    return
  fi
  if command -v py >/dev/null 2>&1; then
    py -3 "$@"
    return
  fi
  if command -v py.exe >/dev/null 2>&1; then
    py.exe -3 "$@"
    return
  fi
  echo "Python 3 is required to parse perf results." >&2
  exit 1
}

resolve_k6_bin() {
  if command -v k6 >/dev/null 2>&1; then
    K6_BIN="$(command -v k6)"
    return
  fi
  if command -v k6.exe >/dev/null 2>&1; then
    K6_BIN="$(command -v k6.exe)"
    return
  fi
  if [[ -x "/c/Program Files/k6/k6.exe" ]]; then
    K6_BIN="/c/Program Files/k6/k6.exe"
    return
  fi
  if [[ -x "/mnt/c/Program Files/k6/k6.exe" ]]; then
    K6_BIN="/mnt/c/Program Files/k6/k6.exe"
    return
  fi
  if [[ -n "${K6_PATH:-}" && -x "${K6_PATH}" ]]; then
    K6_BIN="${K6_PATH}"
  fi
}

require_k6() {
  resolve_k6_bin
  if [[ -z "${K6_BIN}" ]]; then
    echo "k6 is not available in PATH. Install k6 and rerun." >&2
    exit 1
  fi
}

validate_scenario_file() {
  local scenario_name="$1"
  local script_path="${SCRIPT_DIR}/scenarios/${scenario_name}.js"
  if [[ ! -f "${script_path}" ]]; then
    echo "Unknown scenario '${scenario_name}'. Expected file: ${script_path}" >&2
    exit 1
  fi
}

default_case_for_scenario() {
  local scenario_name="$1"
  case "${scenario_name}" in
    smoke)
      echo "4,80"
      ;;
    read_tree)
      echo "4,50"
      ;;
    search)
      echo "4,80"
      ;;
    *)
      echo "4,80"
      ;;
  esac
}

build_cases() {
  if [[ "${SCENARIO}" == "full" ]]; then
    cat <<'EOF'
smoke|1|120|smoke_v1
smoke|2|180|smoke_v2
smoke|4|240|smoke_v4
smoke|8|320|smoke_v8
smoke|16|480|smoke_v16
read_tree|1|100|read_tree_v1
read_tree|2|160|read_tree_v2
read_tree|4|220|read_tree_v4
read_tree|8|300|read_tree_v8
read_tree|16|420|read_tree_v16
search|1|120|search_v1
search|2|180|search_v2
search|4|240|search_v4
search|8|320|search_v8
search|16|480|search_v16
EOF
    return
  fi

  validate_scenario_file "${SCENARIO}"
  local defaults
  defaults="$(default_case_for_scenario "${SCENARIO}")"
  local default_vus="${defaults%%,*}"
  local default_iterations="${defaults##*,}"
  local vus="${PERF_VUS:-${default_vus}}"
  local iterations="${PERF_ITERATIONS:-${default_iterations}}"
  echo "${SCENARIO}|${vus}|${iterations}|${SCENARIO}_v${vus}"
}

get_backend_container_name() {
  if [[ -n "${BACKEND_CONTAINER}" ]]; then
    echo "${BACKEND_CONTAINER}"
    return
  fi

  if command -v docker >/dev/null 2>&1; then
    docker ps --format '{{.Names}}' | grep -E '(^|[-_])(backend|activist-backend)$' | head -n 1 || true
  fi
}

parse_docker_resources() {
  local raw="$1"
  run_python - "${raw}" <<'PY'
import re
import sys

raw = sys.argv[1]
if "|" not in raw:
    raise SystemExit("Invalid docker stats payload")

cpu_raw, mem_raw = raw.split("|", 1)
cpu_raw = cpu_raw.strip().rstrip("%")
try:
    cpu = float(cpu_raw)
except ValueError:
    cpu = 0.0

mem_part = mem_raw.split("/")[0].strip()
match = re.match(r"^([0-9]*\.?[0-9]+)\s*([A-Za-z]+)$", mem_part)
if not match:
    raise SystemExit(f"Unable to parse docker memory usage: {mem_part}")

value = float(match.group(1))
unit = match.group(2).lower()
factors = {
    "b": 1 / (1024 * 1024),
    "kb": 1000 / (1024 * 1024),
    "kib": 1 / 1024,
    "mb": 1000 * 1000 / (1024 * 1024),
    "mib": 1,
    "gb": 1000 * 1000 * 1000 / (1024 * 1024),
    "gib": 1024,
}
if unit not in factors:
    raise SystemExit(f"Unsupported docker memory unit: {unit}")

mem_mb = value * factors[unit]
print(f"{mem_mb:.3f},{cpu:.3f}")
PY
}

read_process_resources() {
  powershell.exe -NoProfile -Command "& {
    \$port = ${BACKEND_PORT}
    \$conn = Get-NetTCPConnection -State Listen -LocalPort \$port -ErrorAction SilentlyContinue | Select-Object -First 1
    if (-not \$conn) { exit 0 }
    \$ownerPid = \$conn.OwningProcess
    \$proc = Get-Process -Id \$ownerPid -ErrorAction SilentlyContinue
    if (-not \$proc) { exit 0 }
    \$perf = Get-CimInstance Win32_PerfFormattedData_PerfProc_Process -Filter \"IDProcess=\$ownerPid\" -ErrorAction SilentlyContinue | Select-Object -First 1
    \$cpu = 0
    if (\$perf) { \$cpu = [double]\$perf.PercentProcessorTime }
    [Console]::Write((\$proc.WorkingSet64.ToString() + \",\" + \$cpu.ToString()))
  }" | tr -d '\r'
}

resolve_backend_resources() {
  local container_name
  container_name="$(get_backend_container_name)"
  if [[ -n "${container_name}" ]]; then
    local usage
    usage="$(docker stats --no-stream --format "{{.CPUPerc}}|{{.MemUsage}}" "${container_name}" 2>/dev/null || true)"
    if [[ -n "${usage}" ]]; then
      parse_docker_resources "${usage}"
      return
    fi
  fi

  local process_stats
  process_stats="$(read_process_resources)"
  if [[ "${process_stats}" =~ ^[0-9]+,[0-9]+(\.[0-9]+)?$ ]]; then
    run_python - "${process_stats}" <<'PY'
import sys

mem_raw, cpu_raw = sys.argv[1].split(",", 1)
mem_mb = int(mem_raw) / (1024 * 1024)
cpu = float(cpu_raw)
print(f"{mem_mb:.3f},{cpu:.3f}")
PY
    return
  fi

  echo "Unable to determine backend resources. Set PERF_BACKEND_CONTAINER or ensure backend listens on PERF_BACKEND_PORT." >&2
  exit 1
}

average_resources() {
  run_python - "$1" "$2" "$3" <<'PY'
import sys

pairs = []
for raw in sys.argv[1:]:
    mem, cpu = raw.split(",", 1)
    pairs.append((float(mem), float(cpu)))

mem_avg = sum(p[0] for p in pairs) / len(pairs)
cpu_avg = sum(p[1] for p in pairs) / len(pairs)
print(f"{mem_avg:.3f},{cpu_avg:.3f}")
PY
}

extract_k6_metrics() {
  local summary_json="$1"
  run_python - "${summary_json}" <<'PY'
import json
import sys

summary_path = sys.argv[1]
with open(summary_path, encoding="utf-8") as handle:
    payload = json.load(handle)

metrics = payload.get("metrics", {})

def values(metric_name: str) -> dict:
    raw = metrics.get(metric_name, {})
    if isinstance(raw, dict) and isinstance(raw.get("values"), dict):
        return raw["values"]
    if isinstance(raw, dict):
        return raw
    return {}

def pick(metric_name: str, *keys: str) -> float:
    data = values(metric_name)
    for key in keys:
        value = data.get(key)
        if isinstance(value, (int, float)):
            return float(value)
    return 0.0

req_count = pick("http_reqs", "count")
rps = pick("http_reqs", "rate")
error_rate = pick("http_req_failed", "rate", "value")
checks_rate = pick("checks", "rate", "value")
avg_ms = pick("http_req_duration", "avg")
p50_ms = pick("http_req_duration", "p(50)", "med")
p95_ms = pick("http_req_duration", "p(95)")
p99_ms = pick("http_req_duration", "p(99)")
max_ms = pick("http_req_duration", "max")
blocked_avg_ms = pick("http_req_blocked", "avg")
waiting_avg_ms = pick("http_req_waiting", "avg")
connecting_avg_ms = pick("http_req_connecting", "avg")
sending_avg_ms = pick("http_req_sending", "avg")
receiving_avg_ms = pick("http_req_receiving", "avg")
tls_avg_ms = pick("http_req_tls_handshaking", "avg")
data_received_mb = pick("data_received", "count") / (1024 * 1024)
data_sent_mb = pick("data_sent", "count") / (1024 * 1024)
vus_max = pick("vus_max", "max", "value")

print(
    f"{req_count:.0f},{rps:.6f},{error_rate:.6f},{checks_rate:.6f},"
    f"{avg_ms:.3f},{p50_ms:.3f},{p95_ms:.3f},{p99_ms:.3f},{max_ms:.3f},"
    f"{blocked_avg_ms:.3f},{waiting_avg_ms:.3f},{connecting_avg_ms:.3f},"
    f"{sending_avg_ms:.3f},{receiving_avg_ms:.3f},{tls_avg_ms:.3f},"
    f"{data_received_mb:.3f},{data_sent_mb:.3f},{vus_max:.3f}"
)
PY
}

run_k6_case() {
  local scenario_name="$1"
  local vus="$2"
  local iterations="$3"
  local profile_label="$4"
  local run_id="$5"
  local timestamp_utc="$6"
  local commit_sha="$7"

  local script_path="${SCRIPT_DIR}/scenarios/${scenario_name}.js"
  validate_scenario_file "${scenario_name}"

  local summary_json="${RAW_DIR}/${profile_label}.json"
  local k6_script_path="${script_path}"
  local k6_summary_path="${summary_json}"

  if [[ "${K6_BIN}" == *.exe ]]; then
    if command -v wslpath >/dev/null 2>&1; then
      k6_script_path="$(wslpath -w "${script_path}")"
      k6_summary_path="$(wslpath -w "${summary_json}")"
    elif command -v cygpath >/dev/null 2>&1; then
      k6_script_path="$(cygpath -w "${script_path}")"
      k6_summary_path="$(cygpath -w "${summary_json}")"
    fi
  fi

  local before_res
  local mid_res
  local after_res
  before_res="$(resolve_backend_resources)"

  local k6_env_args=(
    -e "PERF_BASE_URL=${BASE_URL}"
    -e "PERF_VUS=${vus}"
    -e "PERF_ITERATIONS=${iterations}"
  )
  if [[ -n "${PERF_LOGIN:-}" ]]; then
    k6_env_args+=(-e "PERF_LOGIN=${PERF_LOGIN}")
  fi
  if [[ -n "${PERF_PASSWORD:-}" ]]; then
    k6_env_args+=(-e "PERF_PASSWORD=${PERF_PASSWORD}")
  fi
  if [[ -n "${PERF_AUTH_CREDENTIALS:-}" ]]; then
    k6_env_args+=(-e "PERF_AUTH_CREDENTIALS=${PERF_AUTH_CREDENTIALS}")
  fi
  if [[ -n "${PERF_SKIP_AUTH:-}" ]]; then
    k6_env_args+=(-e "PERF_SKIP_AUTH=${PERF_SKIP_AUTH}")
  fi

  "${K6_BIN}" run \
    "${k6_env_args[@]}" \
    --summary-export "${k6_summary_path}" \
    "${k6_script_path}" >/dev/null &
  local k6_pid=$!

  sleep 2
  if kill -0 "${k6_pid}" >/dev/null 2>&1; then
    mid_res="$(resolve_backend_resources)"
  else
    mid_res="${before_res}"
  fi

  wait "${k6_pid}"
  after_res="$(resolve_backend_resources)"

  local avg_resources
  local metrics
  local memory_avg_mb
  local cpu_avg_pct
  local req_count
  local rps
  local error_rate
  local checks_rate
  local avg_ms
  local p50_ms
  local p95_ms
  local p99_ms
  local max_ms
  local blocked_avg_ms
  local waiting_avg_ms
  local connecting_avg_ms
  local sending_avg_ms
  local receiving_avg_ms
  local tls_avg_ms
  local data_received_mb
  local data_sent_mb
  local vus_max

  avg_resources="$(average_resources "${before_res}" "${mid_res}" "${after_res}")"
  IFS=',' read -r memory_avg_mb cpu_avg_pct <<<"${avg_resources}"

  metrics="$(extract_k6_metrics "${summary_json}")"
  IFS=',' read -r req_count rps error_rate checks_rate avg_ms p50_ms p95_ms p99_ms max_ms blocked_avg_ms waiting_avg_ms connecting_avg_ms sending_avg_ms receiving_avg_ms tls_avg_ms data_received_mb data_sent_mb vus_max <<<"${metrics}"

  printf "%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s\n" \
    "${run_id}" \
    "${timestamp_utc}" \
    "${scenario_name}" \
    "${profile_label}" \
    "${commit_sha}" \
    "${vus}" \
    "${iterations}" \
    "${vus_max}" \
    "${req_count}" \
    "${rps}" \
    "${error_rate}" \
    "${checks_rate}" \
    "${avg_ms}" \
    "${p50_ms}" \
    "${p95_ms}" \
    "${p99_ms}" \
    "${max_ms}" \
    "${blocked_avg_ms}" \
    "${waiting_avg_ms}" \
    "${connecting_avg_ms}" \
    "${sending_avg_ms}" \
    "${receiving_avg_ms}" \
    "${tls_avg_ms}" \
    "${data_received_mb}" \
    "${data_sent_mb}" \
    "${memory_avg_mb}" \
    "${cpu_avg_pct}" \
    "${summary_json}" >> "${SUMMARY_CSV}"

  echo "[case] ${profile_label}: rps=${rps} p95=${p95_ms}ms err=${error_rate} mem=${memory_avg_mb}MB cpu=${cpu_avg_pct}%"
}

require_k6

mapfile -t CASES < <(build_cases)
if (( ${#CASES[@]} == 0 )); then
  echo "No test cases generated." >&2
  exit 1
fi

timestamp_utc="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"
compact_ts="$(date -u +"%Y%m%dT%H%M%SZ")"
sha_short="$(git -C "${SCRIPT_DIR}" rev-parse --short HEAD 2>/dev/null || echo "nogit")"
commit_sha="$(git -C "${SCRIPT_DIR}" rev-parse HEAD 2>/dev/null || echo "unknown")"
branch_name="$(git -C "${SCRIPT_DIR}" rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")"
run_id="${compact_ts}_${SCENARIO}_${sha_short}"

rm -rf "${CURRENT_DIR}"
mkdir -p "${RAW_DIR}"

printf "run_id,timestamp_utc,scenario_name,profile_label,git_sha,vus,iterations,vus_max,req_count,rps,error_rate,checks_rate,avg_ms,p50_ms,p95_ms,p99_ms,max_ms,blocked_avg_ms,waiting_avg_ms,connecting_avg_ms,sending_avg_ms,receiving_avg_ms,tls_avg_ms,data_received_mb,data_sent_mb,memory_avg_mb,cpu_avg_pct,raw_summary_json\n" > "${SUMMARY_CSV}"

for case_def in "${CASES[@]}"; do
  IFS='|' read -r scenario_name vus iterations profile_label <<<"${case_def}"
  run_k6_case "${scenario_name}" "${vus}" "${iterations}" "${profile_label}" "${run_id}" "${timestamp_utc}" "${commit_sha}"
done

run_python - "${META_JSON}" "${run_id}" "${timestamp_utc}" "${SCENARIO}" "${PERF_PROFILE}" "${BASE_URL}" "${commit_sha}" "${branch_name}" "${SUMMARY_CSV}" "${CURRENT_DIR}" "${CASES[@]}" <<'PY'
import json
import sys

meta_path = sys.argv[1]
case_args = sys.argv[11:]

cases = []
for raw in case_args:
    scenario_name, vus, iterations, profile_label = raw.split("|", 3)
    cases.append(
        {
            "scenario_name": scenario_name,
            "profile_label": profile_label,
            "vus": int(vus),
            "iterations": int(iterations),
        }
    )

payload = {
    "run_id": sys.argv[2],
    "timestamp_utc": sys.argv[3],
    "mode": sys.argv[4],
    "profile": sys.argv[5],
    "base_url": sys.argv[6],
    "git_sha": sys.argv[7],
    "branch": sys.argv[8],
    "summary_csv": sys.argv[9],
    "results_dir": sys.argv[10],
    "cases": cases,
}

with open(meta_path, "w", encoding="utf-8") as handle:
    json.dump(payload, handle, indent=2)
    handle.write("\n")
PY

echo "Perf snapshot completed: ${run_id}"
echo "Artifacts:"
echo "- ${META_JSON}"
echo "- ${SUMMARY_CSV}"
echo "- ${RAW_DIR}/*.json"
