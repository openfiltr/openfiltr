#!/usr/bin/env bash
# scripts/create-issues.sh
# Seeds GitHub labels, milestones, and issues for the OpenFiltr project.
#
# Usage:
#   DRY_RUN=true bash scripts/create-issues.sh   # preview only
#   bash scripts/create-issues.sh                # create everything
#
# Requires: gh (GitHub CLI), jq
# The GH_TOKEN env var must have issues:write permission.

set -euo pipefail

REPO="${GITHUB_REPOSITORY:-Bigalan09/openfiltr}"
DRY_RUN="${DRY_RUN:-false}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LABELS_FILE="${SCRIPT_DIR}/../.github/project-setup/labels.json"
ISSUES_FILE="${SCRIPT_DIR}/../.github/project-setup/issues.json"

# ── Colours ──────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RED='\033[0;31m'
RESET='\033[0m'

info()    { echo -e "${CYAN}ℹ️  $*${RESET}"; }
success() { echo -e "${GREEN}✅ $*${RESET}"; }
warn()    { echo -e "${YELLOW}⚠️  $*${RESET}"; }
error()   { echo -e "${RED}❌ $*${RESET}" >&2; }

# ── Dependency checks ─────────────────────────────────────────────────────────
for cmd in gh jq; do
  if ! command -v "$cmd" &>/dev/null; then
    error "Required command not found: $cmd"
    exit 1
  fi
done

if [[ "$DRY_RUN" == "true" ]]; then
  warn "DRY RUN mode — no changes will be made to GitHub"
fi

info "Working on repository: ${REPO}"

# ── Step 1: Create milestones ─────────────────────────────────────────────────
MILESTONES=("v1.0" "v1.1" "v1.2")
info "Creating milestones…"

for ms in "${MILESTONES[@]}"; do
  if [[ "$DRY_RUN" == "true" ]]; then
    echo "  [dry-run] Would create milestone: ${ms}"
    continue
  fi
  # Check if milestone already exists
  existing=$(gh api "repos/${REPO}/milestones" --jq ".[] | select(.title==\"${ms}\") | .number" 2>/dev/null || echo "")
  if [[ -n "$existing" ]]; then
    warn "Milestone '${ms}' already exists (number: ${existing}) — skipping"
  else
    gh api "repos/${REPO}/milestones" \
      --method POST \
      --field title="${ms}" \
      --field state="open" \
      --silent \
      && success "Created milestone: ${ms}"
  fi
done

# ── Step 2: Create labels ─────────────────────────────────────────────────────
info "Creating labels from ${LABELS_FILE}…"

label_count=$(jq 'length' "$LABELS_FILE")
info "Found ${label_count} labels to create"

while IFS= read -r label_json; do
  name=$(echo "$label_json"    | jq -r '.name')
  color=$(echo "$label_json"   | jq -r '.color')
  description=$(echo "$label_json" | jq -r '.description')

  if [[ "$DRY_RUN" == "true" ]]; then
    echo "  [dry-run] Would create label: ${name} (#${color})"
    continue
  fi

  # Try to create; if it exists (422), update it
  response=$(gh api "repos/${REPO}/labels" \
    --method POST \
    --field name="${name}" \
    --field color="${color}" \
    --field description="${description}" \
    --silent 2>&1) && result="created" || result="conflict"

  if [[ "$result" == "conflict" ]]; then
    # Update existing label
    encoded_name=$(python3 -c "import urllib.parse; print(urllib.parse.quote('${name}', safe=''))")
    gh api "repos/${REPO}/labels/${encoded_name}" \
      --method PATCH \
      --field color="${color}" \
      --field description="${description}" \
      --silent 2>/dev/null \
      && warn "Updated existing label: ${name}" \
      || warn "Could not update label: ${name}"
  else
    success "Created label: ${name}"
  fi
done < <(jq -c '.[]' "$LABELS_FILE")

# ── Step 3: Build milestone number map ────────────────────────────────────────
declare -A MILESTONE_NUMBERS

if [[ "$DRY_RUN" != "true" ]]; then
  while IFS=$'\t' read -r number title; do
    MILESTONE_NUMBERS["$title"]="$number"
  done < <(gh api "repos/${REPO}/milestones" --jq '.[] | [.number, .title] | @tsv' 2>/dev/null)
fi

# ── Step 4: Create issues ─────────────────────────────────────────────────────
info "Creating issues from ${ISSUES_FILE}…"

issue_count=$(jq 'length' "$ISSUES_FILE")
info "Found ${issue_count} issues to create"

created=0
skipped=0
failed=0

while IFS= read -r issue_json; do
  title=$(echo "$issue_json"     | jq -r '.title')
  body=$(echo "$issue_json"      | jq -r '.body')
  milestone_name=$(echo "$issue_json" | jq -r '.milestone // empty')
  labels_array=$(echo "$issue_json"   | jq -r '.labels | join(",")' 2>/dev/null || echo "")

  if [[ "$DRY_RUN" == "true" ]]; then
    echo "  [dry-run] Would create issue: ${title}"
    echo "            Labels: ${labels_array}"
    echo "            Milestone: ${milestone_name}"
    continue
  fi

  # Check if issue with same title already exists
  existing=$(gh issue list \
    --repo "${REPO}" \
    --search "\"${title}\" in:title" \
    --state all \
    --json title \
    --jq ".[] | select(.title==\"${title}\") | .title" \
    2>/dev/null | head -1)

  if [[ -n "$existing" ]]; then
    warn "Issue already exists: ${title} — skipping"
    ((skipped++)) || true
    continue
  fi

  # Build gh issue create arguments
  gh_args=(
    --repo "${REPO}"
    --title "${title}"
    --body "${body}"
  )

  # Add labels (create them one at a time via --label flag)
  if [[ -n "$labels_array" ]]; then
    IFS=',' read -ra labels_list <<< "$labels_array"
    for lbl in "${labels_list[@]}"; do
      gh_args+=(--label "${lbl}")
    done
  fi

  # Add milestone if it exists
  if [[ -n "$milestone_name" ]] && [[ -n "${MILESTONE_NUMBERS[$milestone_name]:-}" ]]; then
    gh_args+=(--milestone "${milestone_name}")
  fi

  if gh issue create "${gh_args[@]}" --silent 2>/dev/null; then
    success "Created issue: ${title}"
    ((created++)) || true
  else
    error "Failed to create issue: ${title}"
    ((failed++)) || true
  fi

  # Small delay to avoid secondary rate limits
  sleep 0.3

done < <(jq -c '.[]' "$ISSUES_FILE")

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
info "═══════════════════════════════════════════"
if [[ "$DRY_RUN" == "true" ]]; then
  info "DRY RUN complete — no changes were made"
else
  info "Summary:"
  success "  Issues created: ${created}"
  [[ $skipped -gt 0 ]] && warn "  Issues skipped (already exist): ${skipped}"
  [[ $failed  -gt 0 ]] && error "  Issues failed: ${failed}"
fi
info "═══════════════════════════════════════════"
