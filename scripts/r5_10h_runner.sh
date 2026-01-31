#!/usr/bin/env bash
set -euo pipefail

# Keep running if the parent session is closed (best-effort).
trap '' HUP

ROOT="$(git rev-parse --show-toplevel)"
START_TS="${START_TS:-$(date +%s)}"
END_TS="${END_TS:-$((START_TS + 36000))}"
MAX_ITERS="${MAX_ITERS:-}"
SLEEP_SECS="${SLEEP_SECS:-5}"
ITER=0

race_targets=("./pkg/util/..." "./pkg/waveorch/..." "./pkg/remote/..." "./cmd/...")
static_targets=("golangci-lint run ./..." "govulncheck ./..." "staticcheck ./..." "gosec ./...")
script_targets=("./scripts/wave_orch_e2e_smoke.sh" "./scripts/wave_orch_demo_3_agents.sh" "./scripts/wave_orch_demo_multi_project.sh")

mkdir -p /tmp/r5_10h

while [ "$(date +%s)" -lt "$END_TS" ]; do
  ITER=$((ITER + 1))
  RUN_TS="$(date +%Y%m%dT%H%M%S)"
  RUN_DIR="/tmp/r5_10h/$RUN_TS"
  mkdir -p "$RUN_DIR"
  echo "run=$ITER ts=$RUN_TS" | tee "$RUN_DIR/run.txt"

  (cd "$ROOT" && go test ./... 2>&1 | tee "$RUN_DIR/go_test.log" || true)
  (cd "$ROOT" && go vet ./... 2>&1 | tee "$RUN_DIR/go_vet.log" || true)

  (cd "$ROOT" && go test ./pkg/util/... -count=25 -shuffle=on 2>&1 | tee "$RUN_DIR/flake_util.log" || true)
  (cd "$ROOT" && go test ./pkg/waveorch/... -count=25 -shuffle=on 2>&1 | tee "$RUN_DIR/flake_waveorch.log" || true)
  (cd "$ROOT" && go test ./pkg/remote/... -count=25 -shuffle=on 2>&1 | tee "$RUN_DIR/flake_remote.log" || true)

  race_target="${race_targets[$(((ITER - 1) % ${#race_targets[@]}))]}"
  (cd "$ROOT" && go test -race $race_target 2>&1 | tee "$RUN_DIR/race.log" || true)

  if [ $((ITER % 3)) -eq 0 ]; then
    for s in "${script_targets[@]}"; do
      (cd "$ROOT" && $s 2>&1 | tee "$RUN_DIR/$(basename "$s").log" || true)
    done
  else
    s="${script_targets[$(((ITER - 1) % ${#script_targets[@]}))]}"
    (cd "$ROOT" && $s 2>&1 | tee "$RUN_DIR/$(basename "$s").log" || true)
  fi

  static_cmd="${static_targets[$(((ITER - 1) % ${#static_targets[@]}))]}"
  (cd "$ROOT" && bash -lc "$static_cmd" 2>&1 | tee "$RUN_DIR/static.log" || true)

  (cd "$ROOT" && scripts/security/scan_tokens.sh 2>&1 | tee "$RUN_DIR/token_scan.log" || true)

  echo "run=$ITER done" | tee -a "$RUN_DIR/run.txt"

  if [ -n "$MAX_ITERS" ] && [ "$ITER" -ge "$MAX_ITERS" ]; then
    break
  fi

  sleep "$SLEEP_SECS"

done
