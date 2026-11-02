---
title: Troubleshooting
description: Understand retention messages, Git authentication, missing workspaces, and subshell behavior.
---

Start with `desk list` and `desk gc --dry-run`. They show the first reason a
workspace is retained. Do not delete files or Git history just to satisfy cleanup.

## Cleanup retains a Git-clean repository

`git status` is only one check. A clean checkout can still contain work that
exists nowhere else.

| Retention reason | What to check |
| --- | --- |
| Tracked changes, untracked files, or ignored files | Run `git status --short --ignored`. Preserve `.env`, databases, build output, and other files you need. Ignored does not mean disposable. |
| Local-only branch or tag name | Inspect `git branch` and `git tag`. desk requires each local name on origin; commits alone are not enough. |
| Local-only commits, including refs or reflogs | Inspect your refs, stashes, and `git reflog`. A deleted branch can leave recoverable history. Preserve work before changing it. |
| Origin unavailable | Check connectivity and your normal Git authentication. An offline or deleted remote cannot establish recovery. |
| No locally verifiable commits advertised by origin | Empty repositories or missing local evidence are retained. desk does not fetch automatically. |
| Content filters, submodules, extra worktrees, or shallow history | These need manual retention because desk cannot prove their data is recoverable from the commit graph alone. |

To inspect the remote evidence without changing your checkout:

```sh
git remote -v
git ls-remote --heads --tags origin
```

Run these inside the affected workspace. desk checks **origin's current refs**,
not just cached `origin/*` tracking refs. Pushing to another remote does not
necessarily resolve retention. A stale fetch is not proof that a remote still
holds your work.

There is no force flag. If you intentionally want to retain configuration,
hooks, or other personal data in `.git`, pin the workspace.

## Scratch contains files

An empty scratch workspace has no directory entries at all. Dotfiles and empty
subdirectories count. Inspect hidden files with `ls -la`, preserve what matters,
and leave uncertain files alone. desk never recursively guesses which scratch
contents are safe to remove.

## Active session or unavailable session lock

Leave all desk subshells using that workspace with `exit`, and stop other tools
that use the directory. More than one desk shell can be open at once. An
unavailable lock also causes retention; it is not permission to bypass the check.

`--quiescent` does not override an active session. It acknowledges that you have
also stopped external processes desk cannot reliably detect.

## A workspace is not marked done

Batch cleanup considers only work explicitly marked done:

```sh
desk done ID
desk gc --dry-run
```

Entering it again clears done. A pin must be removed with `desk keep ID --off`
before `done` is accepted. Removing a pin alone does not mark it done.

## A workspace is quarantined

Restore before trying to enter or reopen it:

```sh
desk restore ID
desk enter ID
```

If purge says the quarantine is too young, wait until at least seven days have
elapsed. If its timestamp is missing, restore and quarantine it again to begin
a known waiting period. Do not edit the registry to bypass it.

## Git authentication fails

Check the same remote outside desk with `git ls-remote URL`, using your normal
credential helper or SSH agent. GitHub shorthand uses HTTPS; if your credentials
use SSH, pass the SSH URL instead. This may create a different workspace because
SSH and HTTPS sources are not treated as aliases.

desk disables Git terminal password prompts. SSH and askpass prompts may still
appear. Never put a token in a repository URL. After a failed clone, use the
workspace ID in the error to inspect anything left behind.

## A workspace is missing from the listing

Check the [data-root configuration](/configuration/). A different `DESK_HOME`,
Linux `XDG_DATA_HOME`, or user account selects a different registry. Setting a
different root does not migrate existing workspaces.

If desk reports a missing directory, it retains the registry entry rather than
inventing a replacement or deleting evidence. Do not manually edit the registry
or reuse the ID. Recover from your independent backup if files were removed by
another program; desk is not a backup system.

## The shell starts in the wrong directory

desk initially starts the subshell in the workspace. Your shell's startup files
can then run `cd`. Inspect those files or try an installed alternative shell
using `SHELL=/bin/bash desk enter ID`. `SHELL` cannot contain flags or arguments.

There is no `desk cd`: a child executable cannot change the directory of your
calling shell. Use `exit` to return to it.
