#!/usr/bin/env bash
set -euo pipefail

# Minimal orchestrator to ensure Xiaohongshu MCP is running and logged in.
# - Starts server on PORT if not listening
# - Ensures a valid MCP session
# - Prompts QR scan (saved to /tmp/xhs_login_qr.png) and polls until logged in
# - If extra args are provided, runs them after login (e.g., your other services)
#
# Usage examples:
#   scripts/xhs-ready.sh
#   PORT=18061 scripts/xhs-ready.sh echo "MCP ready on $PORT"
#
# Requirements: curl, jq, rg (ripgrep), base64, lsof, open (macOS)

PORT=${PORT:-18060}
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
ROOT_DIR=$(cd "${SCRIPT_DIR}/.." && pwd)

DEFAULT_BIN=""
for candidate in \
  "${ROOT_DIR}/mcp/xiaohongshu-mcp" \
  "${ROOT_DIR}/xiaohongshu-mcp-darwin-arm64"
do
  if [[ -x "${candidate}" ]]; then
    DEFAULT_BIN="${candidate}"
    break
  fi
done

if [[ -z "${DEFAULT_BIN}" ]]; then
  DEFAULT_BIN="${ROOT_DIR}/mcp/xiaohongshu-mcp"
fi

BIN=${BIN:-"${DEFAULT_BIN}"}
CHROME_BIN_DEFAULT="/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"

log() { printf "[xhs-ready] %s\n" "$*"; }
fail() { printf "[xhs-ready] ERROR: %s\n" "$*" >&2; exit 1; }

[[ -x "${BIN}" ]] || fail "Binary not found or not executable: ${BIN}"

# 1) Start server if needed
if ! lsof -tiTCP:"$PORT" -sTCP:LISTEN >/dev/null 2>&1; then
  log "Starting MCP server on :$PORT (GUI mode)"
  if [[ -x "$CHROME_BIN_DEFAULT" ]]; then
    nohup "$BIN" -headless=false -port ":$PORT" -bin "$CHROME_BIN_DEFAULT" >> xiaohongshu-mcp.log 2>&1 & echo $! > xiaohongshu-mcp.pid
  else
    nohup "$BIN" -headless=false -port ":$PORT" >> xiaohongshu-mcp.log 2>&1 & echo $! > xiaohongshu-mcp.pid
  fi
  # Wait until listening
  for i in {1..30}; do
    sleep 0.3
    if lsof -tiTCP:"$PORT" -sTCP:LISTEN >/dev/null 2>&1; then
      break
    fi
    if [[ $i -eq 30 ]]; then
      fail "Server did not start listening on :$PORT"
    fi
  done
else
  log "Server already listening on :$PORT"
fi

# 2) Initialize MCP session
INIT='{"jsonrpc":"2.0","id":"1","method":"initialize","params":{"clientInfo":{"name":"xhs-ready","version":"0.1.0"},"protocolVersion":"2024-09-18","capabilities":{}}}'
SESSION=$(curl -sS -D - -o /dev/null -X POST "http://127.0.0.1:${PORT}/mcp" \
  -H 'content-type: application/json' --data-binary "$INIT" | rg -o "Mcp-Session-Id: (.+)" -r '$1' | tr -d '\r')
[[ -n "$SESSION" ]] || fail "Failed to obtain Mcp-Session-Id"
log "SESSION=$SESSION"

call_tool() {
  local name="$1"; shift
  local args_json="$1"; shift || true
  local id="$RANDOM$RANDOM"
  local payload; payload=$(jq -n --arg name "$name" --argjson args "$args_json" '{jsonrpc:"2.0", id: ("" + $ENV.id), method:"tools/call", params:{name:$name, arguments:$args}}' 2>/dev/null || true)
  # Fallback if jq --argjson fails (e.g., empty)
  if [[ -z "$payload" ]]; then
    payload=$(jq -n --arg name "$name" '{jsonrpc:"2.0", id:"42", method:"tools/call", params:{name:$name, arguments:{}}}')
  fi
  curl -sS -X POST "http://127.0.0.1:${PORT}/mcp" \
    -H 'content-type: application/json' \
    -H "Mcp-Session-Id: $SESSION" \
    -d "$payload"
}

# 3) Check login; if not, request QR and poll
RES=$(call_tool check_login_status '{}')
IS_LOGGED=$(echo "$RES" | jq -r '.result.content[]? | select(.type=="text") | .text | test("IsLoggedIn:true")')
if [[ "$IS_LOGGED" != "true" ]]; then
  log "Not logged in. Requesting QR..."
  QR=$(call_tool get_login_qrcode '{}')
  B64=$(echo "$QR" | jq -r '.result.content[]? | select(.type=="image") | .data // empty')
  if [[ -n "$B64" ]]; then
    echo "$B64" | base64 -D > /tmp/xhs_login_qr.png 2>/dev/null || echo "$B64" | base64 -d > /tmp/xhs_login_qr.png || true
    if [[ -s /tmp/xhs_login_qr.png ]]; then
      log "QR saved to /tmp/xhs_login_qr.png"
      (open /tmp/xhs_login_qr.png >/dev/null 2>&1 || true)
    fi
  fi
  log "Please scan with the Xiaohongshu App within 5 minutes..."
  # poll up to 120s
  for i in {1..120}; do
    sleep 1
    RES=$(call_tool check_login_status '{}') || true
    IS_LOGGED=$(echo "$RES" | jq -r '.result.content[]? | select(.type=="text") | .text | test("IsLoggedIn:true")')
    if [[ "$IS_LOGGED" == "true" ]]; then
      break
    fi
  done
fi

[[ "$IS_LOGGED" == "true" ]] || fail "Login not completed in time"
log "Login OK"

# 4) Optionally run post-login command(s)
if [[ $# -gt 0 ]]; then
  log "Running post-login command: $*"
  exec "$@"
else
  log "MCP ready on :$PORT (SESSION=$SESSION)"
fi
