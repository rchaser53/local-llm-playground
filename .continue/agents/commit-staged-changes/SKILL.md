---
name: commit-staged-changes
description: Use this skill when the user wants you to inspect currently staged git changes, write an appropriate commit message, and create the commit.
---

# Commit Staged Changes

This skill helps create a commit from the current staged changes.

## When to Use This Skill

Use this skill when the user:

- Asks you to commit what is already staged
- Asks for a commit message based on `git add` contents
- Wants you to inspect the staged diff and commit it
- Wants a concise commit message without manually writing it

Do not use this skill when the user wants to stage files first, rewrite commit history, or amend an existing commit unless they ask for that explicitly.

## Core Rules

- Only use staged changes as the source of truth
- Never include unstaged or untracked changes in the commit
- If there are no staged changes, stop and tell the user nothing is staged
- Prefer the repository's existing commit style over generic conventions
- If the staged diff clearly contains multiple unrelated changes, call that out before committing

## Workflow

### Step 1: Confirm There Is Something Staged

Run:

```bash
git status --short
git diff --cached --stat
```

If `git diff --cached --stat` is empty, tell the user there is nothing to commit and stop.

### Step 2: Inspect the Staged Diff Only

Run:

```bash
git diff --cached --unified=3 --no-color
git diff --cached --name-status
```

Use only this staged diff to infer intent. Do not inspect unstaged changes unless the user explicitly asks.

### Step 3: Check Local Commit Style

Run:

```bash
git log --oneline -n 10
```

Infer the local style from recent commits.

Default preferences unless the repository clearly uses something else:

- Short English subject line
- Describe the actual change, not the tool used
- Keep the subject concise and specific

If the repository clearly uses conventional commits, follow that pattern. Otherwise, use a plain summary line.

### Step 4: Draft the Commit Message

Base the message on the dominant user-facing or codebase-facing change in the staged diff.

Prioritize this order when summarizing intent:

1. Primary behavior change
2. New capability
3. Fix or cleanup
4. Supporting docs or config

Good message patterns:

- `Add staged commit helper skill`
- `Update website summary skill instructions`
- `Fix logging config for final response output`

Avoid vague messages such as:

- `update files`
- `misc changes`
- `fix stuff`

### Step 5: Commit

If the user asked you to commit, create the commit directly with the generated message:

```bash
git commit -m "<message>"
```

Only pause to ask the user if one of these is true:

- The staged diff appears to bundle unrelated changes
- The repository has a strong message convention but the correct type or scope is ambiguous
- The commit would be risky because the diff looks incomplete

Otherwise, proceed.

### Step 6: Report Back

Tell the user:

- The commit message you used
- The resulting commit hash
- Whether anything remains unstaged or uncommitted

## Response Guidelines

Keep the response concise and operational.

If the commit succeeds, report the exact message and short hash.

If the commit does not succeed, explain why briefly and include the command output summary.

## Example Triggers

- "今 stage している内容をコミットして"
- "git add 済みの差分からコミットメッセージ考えてそのまま commit して"
- "staged changes を見て適切にコミットして"
- "index の内容だけでコミットを作って"

## Constraints

- Never run `git add .` or restage files unless the user asks
- Never amend or force-push unless the user asks
- Never invent details that are not visible in the staged diff
- Prefer one clear subject line; add a body only when the diff is large and the extra context is genuinely useful