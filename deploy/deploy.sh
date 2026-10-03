#!/bin/zsh
#
# Build Bushwhack for photon (linux/arm64), install it and restart the unit.
# First-time setup (user, port, unit, vhosts) is in DEPLOY.md; this script
# only ships new builds.
#
# Usage:
#   ./deploy/deploy.sh [--dry-run]
#
# Environment:
#   EC2_HOST   ssh alias of the target box. Default `photon`.

set -euo pipefail

die() { print -u2 -r -- "error: $*"; exit 1; }

DRY_RUN=0
[[ "${1:-}" == "--dry-run" ]] && DRY_RUN=1

EC2_HOST="${EC2_HOST:-photon}"
SITE_URL="https://bushwhack.sstools.co"
ROOT="${0:A:h:h}"
BIN="${ROOT}/bin/bushwhack-linux-arm64"

cd "$ROOT"
print "==> Testing"
go vet ./... && go test ./... >/dev/null || die "tests failed"

print "==> Building ${BIN:t}"
make linux >/dev/null
file "$BIN" | grep -q 'ARM aarch64' || die "${BIN} is not a linux/arm64 binary"

if (( DRY_RUN )); then
    print "==> Dry run: would upload ${BIN:t} to ${EC2_HOST} and restart bushwhack"
    exit 0
fi

print "==> Uploading to ${EC2_HOST}"
scp -q "$BIN" "${EC2_HOST}:/tmp/bushwhack.new"

print "==> Installing and restarting"
ssh "$EC2_HOST" '
    set -e
    sudo install -m 0755 -o root -g root /tmp/bushwhack.new /usr/local/bin/bushwhack
    rm -f /tmp/bushwhack.new
    sudo restorecon /usr/local/bin/bushwhack 2>/dev/null || true
    sudo systemctl restart bushwhack
    sleep 1
    systemctl is-active --quiet bushwhack || { sudo journalctl -u bushwhack -n 30 --no-pager; exit 1; }
'

print "==> Verifying"
curl -fsS "${SITE_URL}/healthz"
code=$(curl -s -o /dev/null -w '%{http_code}' "${SITE_URL}/make")
[[ "$code" == "200" ]] || die "/make returned ${code}"
code=$(curl -s -o /dev/null -w '%{http_code}' -F pattern=classic-jack "${SITE_URL}/api/process")
[[ "$code" == "200" ]] || die "/api/process returned ${code}"
print "==> Done. ${SITE_URL}/ is live."
