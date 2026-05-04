# commit-staged-changes

This repository packages a single agent skill that inspects currently staged git changes, drafts an appropriate commit message, and creates the commit.

## Included Files

- SKILL.md

## What It Does

- Checks whether staged changes exist before committing
- Reviews only the staged diff, not unstaged work
- Infers the local commit message style from recent history
- Creates a commit with a concise, context-aware message

## Install

After publishing this folder as its own Git repository, install it with the Skills CLI:

```bash
npx skills add <owner>/<repo>
```

To preview locally before publishing:

```bash
npx skills add . --list
```

## Example Prompt

```text
今 stage している内容を見て、適切なコミットメッセージでそのまま commit して
```

## Notes

- The skill is designed for agents that support the shared agent skills format.
- If you publish this as a single-skill repository, keeping SKILL.md at the repository root is the simplest structure for skills.sh compatibility.