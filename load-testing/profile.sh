#!/bin/sh

CURRENT_TIMESTAMP=$(date +%Y%m%d-%H%M%S)
CURRENT_COMMIT=$(git rev-parse --short HEAD)
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
OUT_DIR="$SCRIPT_DIR/perf-${CURRENT_TIMESTAMP}-${CURRENT_COMMIT}"

mkdir -p "$OUT_DIR"

sleep 15
curl -o $OUT_DIR/cpu.pprof   "http://localhost:6060/debug/pprof/profile?seconds=30"
curl -o $OUT_DIR/block.pprof "http://localhost:6060/debug/pprof/block"
curl -o $OUT_DIR/heap.pprof  "http://localhost:6060/debug/pprof/heap?gc=1"