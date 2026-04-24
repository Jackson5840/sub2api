#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  tools/starsly-reapply.sh <source-feature-branch> <new-feature-branch> [upstream-ref]

Example:
  git checkout main
  git fetch upstream --tags
  git merge upstream/main
  tools/starsly-reapply.sh feature/starsly-0.1.117 feature/starsly-0.1.118 upstream/main

What it does:
  1. Verifies the working tree is clean.
  2. Creates <new-feature-branch> from the current HEAD.
  3. Finds custom commits on <source-feature-branch> after its merge-base with upstream-ref.
  4. Cherry-picks those commits in order.

If a conflict happens:
  - Resolve files in VS Code.
  - Run: git cherry-pick --continue
  - Continue manually with the remaining commits if needed.
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -lt 2 || $# -gt 3 ]]; then
  usage
  exit 1
fi

source_branch="$1"
target_branch="$2"
upstream_ref="${3:-upstream/main}"

if ! git rev-parse --verify --quiet "$source_branch" >/dev/null; then
  echo "Source branch not found: $source_branch" >&2
  exit 1
fi

if ! git rev-parse --verify --quiet "$upstream_ref" >/dev/null; then
  echo "Upstream ref not found: $upstream_ref" >&2
  exit 1
fi

if git rev-parse --verify --quiet "$target_branch" >/dev/null; then
  echo "Target branch already exists: $target_branch" >&2
  exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "Working tree is not clean. Commit or stash changes first." >&2
  git status --short >&2
  exit 1
fi

base="$(git merge-base "$source_branch" "$upstream_ref")"
mapfile -t commits < <(git rev-list --reverse "${base}..${source_branch}")

if [[ ${#commits[@]} -eq 0 ]]; then
  echo "No custom commits found on $source_branch after merge-base with $upstream_ref." >&2
  exit 1
fi

echo "Creating $target_branch from current HEAD: $(git rev-parse --short HEAD)"
git checkout -b "$target_branch"

echo "Reapplying ${#commits[@]} commits from $source_branch:"
for commit in "${commits[@]}"; do
  echo "  - $(git log --format='%h %s' -n 1 "$commit")"
done

for commit in "${commits[@]}"; do
  git cherry-pick "$commit"
done

echo "Done. Current branch:"
git branch --show-current
git log --oneline --decorate -n 12
