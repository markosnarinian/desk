---
title: Repository workspaces
description: Clone with Git, resume the same managed checkout, and preserve work before cleanup.
---

Use `desk open` for an occasional local fix, investigation, or contribution.
desk creates an independent Git clone and enters an interactive shell there.
Git must be installed and your normal Git credentials must already work.

## Clone from GitHub

```sh
desk open octocat/Hello-World
```

`owner/repo` is shorthand for `https://github.com/owner/repo.git`. Within the
workspace, use Git normally:

```sh
git status
git log -1 --oneline
exit
```

No GitHub CLI is required. desk does not install the repository's dependencies,
run its setup scripts, or configure a development environment.

## Resume without resetting

Run the same command again:

```sh
desk open octocat/Hello-World
```

desk resumes the existing checkout for that source. It does **not** pull, reset,
switch branches, or discard changes. You decide when to run `git fetch` or
`git pull`. Use `desk list` and `desk enter ID` to resume by ID instead.

If you previously used `desk done`, entering clears that flag. Opening with
`--no-enter` does not enter a shell and does not clear it.

## Other sources and authentication

These are alternative examples; choose the source you actually need:

```sh
desk open https://github.com/octocat/Hello-World.git
desk open git@github.com:octocat/Hello-World.git
desk open ssh://git@example.com/team/project.git
desk open ./existing-project
desk open /absolute/path/to/project
```

Use an explicit `./`, `../`, or absolute path for local repositories. Local
clones have independent object stores, rather than Git's local-clone hardlink
optimization. Existing directories are cloned, never adopted or moved.

GitHub shorthand and its exact HTTPS expansion identify the same source. Other
URL aliases are **not** deduplicated: SSH and HTTPS can produce separate
workspaces for the same remote repository.

For private repositories, configure your Git credential helper or SSH agent
first. desk disables Git's terminal password prompts, but SSH or askpass tools
may still ask for credentials. Do not embed tokens in URLs: URLs become part of
the saved workspace source. See [authentication troubleshooting](/troubleshooting/#git-authentication-fails).

## Finish a contribution safely

Commit and publish work using your usual Git workflow before considering cleanup.
That can mean pushing to your own fork, not necessarily the original repository.
desk checks **origin** as its recovery source; a push to a different remote alone
does not prove recovery to desk.

Even if `git status` says clean, desk retains local-only branch names, commits,
tags, and reflog history. Ignored files such as `.env`, databases, and dependency
directories also block cleanup. See [retention reasons](/troubleshooting/#cleanup-retains-a-git-clean-repository).

Keep or back up anything valuable, stop processes, then use the
[cleanup guide](/guides/cleanup/). There is no force option.

## Empty repositories and failed clones

An empty repository can be opened and resumed, but desk retains it during
cleanup because it has no remotely verifiable commits.

If cloning fails, desk preserves any remaining files as a scratch workspace and
prints its ID in the error. Inspect it with `desk enter ID`. A later `desk open`
can attempt a new clone; it does not overwrite the failed attempt. An empty
failed attempt can be quarantined like any empty scratch workspace.
