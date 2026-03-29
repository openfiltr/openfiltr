# Codex Agent Workflow

This README is the source of truth for Codex agents working in this repository.

Follow this flow every time you make a real code change.

## 1. Start from the latest `origin/main`

```bash
git fetch origin main
mkdir -p .worktrees
git worktree add .worktrees/<task-name> -b codex/<task-name> origin/main
cd .worktrees/<task-name>
```

Use a short, scoped `<task-name>`. Do not reuse an unrelated feature branch.

## 2. Read the repository state before coding

```bash
sed -n '1,200p' CHANGELOG.md
for f in docs/adr/*.md; do sed -n '1,220p' "$f"; done
go test ./...
```

If `go` is unavailable locally, use Docker:

```bash
docker run --rm -v "$PWD":/work -w /work golang:1.24 go test ./...
```

Do not start coding until you have done this.

## 3. Make the change

- Keep the scope tight.
- Use British English in comments, docs, commit messages, and PR text.
- Use `rg` for search.
- Prefer small fixes with tests over broad refactors.

## 4. Verify the change

Run the narrowest relevant checks first, then the broader suite:

```bash
go test ./internal/dns ./internal/storage
go test ./...
```

Add any other targeted checks the change needs.

## 5. Commit the work

Every commit must be signed off:

```bash
git add <files>
git commit -s -m "fix: short summary"
```

Do not create a PR if there are no code or documentation changes to commit.

## 6. Push to `origin`

Push the branch you created for the task:

```bash
git push -u origin codex/<task-name>
```

If the push fails, fix that before claiming the task is complete.

## 7. Open the GitHub pull request

Use GitHub CLI so the PR is actually visible on GitHub:

```bash
gh pr create \
  --base main \
  --head codex/<task-name> \
  --title "fix: short summary" \
  --body-file <prepared-pr-body>
```

If you do not have a prepared PR body file, create one first:

```bash
cat > /tmp/pr-body.md <<'EOF'
## Summary
- concise bullet

## Testing
- go test ./...
EOF
```

Then run:

```bash
gh pr create \
  --base main \
  --head codex/<task-name> \
  --title "fix: short summary" \
  --body-file /tmp/pr-body.md
```

## 8. Confirm the result

After opening the PR, confirm all of the following:

```bash
git status --short
git rev-parse --short HEAD
gh pr view --web
```

- `git status --short` should be clean.
- `gh pr view --web` should open the real GitHub PR, not just a local draft message.

## Non-negotiable rules

- Do not push directly to `main`.
- Do not stop after only drafting a PR message locally.
- Do not say a PR exists unless the branch is pushed and `gh pr create` has succeeded.
- Do not open a PR for empty or unrelated changes.

## 9. Track migration execution in-repo

For bbolt migration work, keep `docs/codex-agents/bbolt-migration-tracker.md` updated in the same PR:

- move issue status forward
- add the implementation PR link
- capture concise behaviour notes for future agents

## 10. Changelog requirements for all implementation PRs

Every implementation PR must include a `CHANGELOG.md` update under `## [Unreleased]`.

- Use Keep a Changelog categories
- Include concrete context (what changed and why)
- State in the PR body that the changelog text was AI-checked
