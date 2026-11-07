#!/usr/bin/env bash
# Prints the latest commit SHA on the main branch of the upstream
# ghostty repo. Compares it against the pinned commit in
# generator.go and reports whether an update is needed.

set -euo pipefail

REPO="ghostty-org/ghostty"
BRANCH="main"
GENERATOR="generator.go"

# Get the currently pinned commit from generator.go.
pinned=$(grep -oP '(?<=ghosttyCommit = ")[0-9a-f]+' "$GENERATOR" 2>/dev/null || true)
if [ -z "$pinned" ]; then
    echo "ERROR: could not find ghosttyCommit in $GENERATOR" >&2
    exit 1
fi

# Fetch the latest commit from GitHub.
latest=$(git ls-remote "https://github.com/${REPO}.git" "refs/heads/${BRANCH}" \
    | awk '{print $1}')
if [ -z "$latest" ]; then
    echo "ERROR: could not fetch latest commit from ${REPO}" >&2
    exit 1
fi

echo "pinned:  $pinned"
echo "latest:  $latest"

if [ "$pinned" = "$latest" ]; then
    echo "status:  up-to-date"
else
    echo "status:  update-available"
fi
