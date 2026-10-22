#!/bin/bash
# Session start for WAG's cloud sessions:
#  1. fetch the Go modules, so that tests and tools run at once;
#  2. restore the private workspace (github.com/Austin-Routt/wag-workspace)
#     next to WAG, and set WAG_WORKSPACE. See its README.
# It never fails the session: a step that cannot run says why, and the
# session goes on without it.
set -uo pipefail

if [ "${CLAUDE_CODE_REMOTE:-}" != "true" ]; then
  exit 0
fi

repo="${CLAUDE_PROJECT_DIR:-$(pwd)}"
ws="${WAG_WORKSPACE_DIR:-$(dirname "$repo")/wag-workspace}"
url="${WAG_WORKSPACE_URL:-https://github.com/Austin-Routt/wag-workspace}"

if command -v go >/dev/null 2>&1 && [ -f "$repo/wag/go.mod" ]; then
  (cd "$repo/wag" && go mod download) || echo "session-start: go mod download failed; go test will fetch the modules itself."
fi

if [ -d "$ws/.git" ]; then
  if ! timeout 300 git -C "$ws" pull -q --ff-only 2>/dev/null; then
    echo "session-start: the workspace at $ws could not be updated from its remote; it is used as it is."
  fi
elif timeout 600 git clone -q "$url" "$ws" 2>/dev/null; then
  echo "session-start: the workspace is restored at $ws."
else
  echo "session-start: the private workspace ($url) could not be cloned. Attach Austin-Routt/wag-workspace to the session (add_repo, push access) and run .claude/hooks/session-start.sh again."
fi

if [ -d "$ws/.git" ] && [ -n "${CLAUDE_ENV_FILE:-}" ]; then
  echo "export WAG_WORKSPACE=\"$ws\"" >> "$CLAUDE_ENV_FILE"
fi
exit 0
