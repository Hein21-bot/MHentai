#!/usr/bin/env bash
set -euo pipefail

URL="${1:-${BACKEND_HEALTH_URL:-}}"

if [ -z "$URL" ]; then
  cat <<'EOF' >&2
Usage: BACKEND_HEALTH_URL=https://your-backend.onrender.com/api/health ./scripts/ping-health.sh
       ./scripts/ping-health.sh https://your-backend.onrender.com/api/health
EOF
  exit 1
fi

curl --fail --silent --show-error --max-time 20 "$URL"
