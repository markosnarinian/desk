---
title: Quickstart
description: Create a workspace, leave it safely, and preview explicit cleanup.
---

First [install desk](/installation/). This walkthrough creates an experiment,
leaves it safely, and previews cleanup. Nothing here permanently deletes data.

## 1. Start a workspace

```sh
desk scratch experiment
```

desk prints an ID and enters an interactive subshell in a new directory.
Inside that shell:

```sh
pwd
printf 'an idea\n' > notes.txt
exit
```

You are back in your original shell. **Your file is still there.**

## 2. Find and resume it

```sh
desk list
desk enter ID
```

Replace `ID` with the full 12-character ID from the listing, not the label
`experiment`. Inside the resumed shell, `cat notes.txt` shows your idea.
Run `exit` again to return.

## 3. Protect or finish it

```sh
desk keep ID
```

A pin blocks removal. When you are ready to consider cleanup, unpin and mark done:

```sh
desk keep ID --off
desk done ID
desk gc --dry-run
```

Expect `scratch contains files`: your `notes.txt` still matters. Marking done
does not delete it. There is no automatic cleanup on exit or in the background.

## Next steps

- [Scratch workspaces](/guides/scratch/): keep useful results and create without entering.
- [Repository workspaces](/guides/repositories/): open `owner/repo`, use SSH, and resume a checkout.
- [Cleanup and recovery](/guides/cleanup/): stop processes, quarantine eligible work, and restore it.
- [Cleanup guarantees](/safety/): exactly what desk checks and what it cannot promise.

Cleanup is deliberately separate from this first run. desk is not a backup.
