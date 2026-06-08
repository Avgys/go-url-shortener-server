#!/usr/bin/env bash
set -euo pipefail

# Load-test POST / (plain-text shorten). Uses vegeta JSON targets (inline body).
#
# Usage:
#   ./scripts/vegeta/shorten.sh
#   BASE_URL=http://127.0.0.1:8080 RATE=100 DURATION=30s COUNT=1000 ./scripts/vegeta/shorten.sh

BASE_URL="${BASE_URL:-http://localhost:8080}"
RATE="${RATE:-50}"
DURATION="${DURATION:-30s}"
COUNT="${COUNT:-500}"
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

generate_targets() {
	local i id body body_b64
	for ((i = 1; i <= COUNT; i++)); do
		id="$(random_id)"
		body="${LONG_HOST}/${id}"
		if command -v base64 >/dev/null 2>&1; then
			body_b64="$(printf '%s' "$body" | base64 | tr -d '\n')"
		else
			body_b64="$(printf '%s' "$body" | openssl base64 -A)"
		fi
		printf '{"method":"POST","url":"%s/","body":"%s","header":{"Content-Type":["text/plain"]}}\n' \
			"$BASE_URL" "$body_b64"
	done
}

echo "shorten attack: base=$BASE_URL rate=$RATE duration=$DURATION targets=$COUNT" >&2

if [[ -n "$OUTPUT" ]]; then
	generate_targets | vegeta attack -format=json -rate="$RATE" -duration="$DURATION" -output="$OUTPUT"
	vegeta report "$OUTPUT"
else
	generate_targets | vegeta attack -format=json -rate="$RATE" -duration="$DURATION" | vegeta report
fi
