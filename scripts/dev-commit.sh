#!/usr/bin/env bash
set -euo pipefail
MSG=${1:-"chore: automated dnsbro checkpoint"}

git add -A
git commit -m "$MSG"
git push
