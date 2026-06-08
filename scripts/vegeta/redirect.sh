#!/usr/bin/env bash
set -euo pipefail

# Load-test GET /{short} (307 redirect; -redirects=-1 stops at Location header).
#
# Usage:
#   ./scripts/vegeta/redirect.sh
#   SHORT_PATH=/abc12xyz ./scripts/vegeta/redirect.sh
#   SEED_COUNT=10 RATE=100 DURATION=30s ./scripts/vegeta/redirect.sh
#
# Requires: vegeta, curl

BASE_URL="${BASE_URL:-http://localhost:8080}"
RATE="${RATE:-150}"
DURATION="${DURATION:-10s}"
SEED_COUNT="${SEED_COUNT:-1}"
SHORT_PATH="${SHORT_PATH:-}"
LONG_HOST="${LONG_HOST:-http://ofdafnyylfqe.biz/page}"
OUTPUT="${OUTPUT:-}"

random_id() {
	if command -v openssl >/dev/null 2>&1; then
		openssl rand -hex 4
	elif command -v uuidgen >/dev/null 2>&1; then
		uuidgen | tr -d '-' | cut -c1-8
	else
		echo "$RANDOM$RANDOM"
	fi
}

seed_short_path() {
	local id long short full
	id="$(random_id)"
	long="${LONG_HOST}/${id}"
	full="$(curl -sf -X POST -H "Content-Type: text/plain" -d "$long" "${BASE_URL}/")"
	short="${full#"${BASE_URL}"}"
	short="${short#http://localhost:8080}"
	short="${short#https://localhost:8080}"
	echo "${short#/}"
}

generate_targets() {
	local -a paths=()
	local i path

	if [[ -n "$SHORT_PATH" ]]; then
		paths=("$SHORT_PATH")
	else
		echo "seeding ${SEED_COUNT} short URL(s)..." >&2
		for ((i = 1; i <= SEED_COUNT; i++)); do
			path="$(seed_short_path)"
			echo "  -> ${BASE_URL}/${path}" >&2
			paths+=("$path")
		done
	fi

	for path in "${paths[@]}"; do
		printf 'GET %s/%s\n' "$BASE_URL" "${path#/}"
	done
}

echo "redirect attack: base=$BASE_URL rate=$RATE duration=$DURATION" >&2

# -redirects=-1: do not follow Location; count 3xx as success
if [[ -n "$OUTPUT" ]]; then
	generate_targets | vegeta attack -rate="$RATE" -duration="$DURATION" -redirects=-1 -output="$OUTPUT"
	vegeta report "$OUTPUT"
else
	generate_targets | vegeta attack -rate="$RATE" -duration="$DURATION" -redirects=-1 | vegeta report
fi
