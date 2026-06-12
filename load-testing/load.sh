#!/bin/sh

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
(
cd "$SCRIPT_DIR" || exit 1
vegeta attack -targets="$SCRIPT_DIR/attack.txt" -duration=90s -rate=100/s
) | vegeta report