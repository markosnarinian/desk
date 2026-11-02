---
title: Cleanup and recovery
description: Preview retention, quarantine finished work, restore mistakes, and understand permanent removal.
---

**Leaving a shell never deletes a workspace.** Cleanup is an explicit decision,
not a timer. Read the [cleanup guarantees](/safety/) for the full contract and
its limits; this guide walks through the everyday workflow.

In the commands below, replace `ID` with a full workspace ID from `desk list`.

## 1. Protect work you are keeping

```sh
desk keep ID
```

Pins block quarantine and purge, including for work already in quarantine.
To unpin later, use `desk keep ID --off`. Unpinning does not mark anything done.

## 2. Mark finished work

Exit your workspace shell, then run:

```sh
desk done ID
desk gc --dry-run
```

`done` marks a cleanup candidate; it does not empty files, stop processes, or
remove the workspace. A pinned workspace must be unpinned before marking it done.

The preview reports either a reason to retain each workspace or that it would
be quarantined. For example, these illustrative IDs represent different states:

```text
8a7b6c5d4e3f retained: not marked done
1a2b3c4d5e6f retained: scratch contains files
9a8b7c6d5e4f would quarantine (requires --apply --quiescent)
```

`desk gc` without flags is also a dry run. Only the first blocking reason is
shown. Resolving it can reveal another; see [troubleshooting](/troubleshooting/).

## 3. Stop everything using the workspace

Before quarantine or purge, stop all relevant shells, editors, dev servers,
watchers, containers using mounted files, and detached processes. This includes
workspaces affected by a batch cleanup, not just the terminal you are using.

The required `--quiescent` flag is **your acknowledgment**, not a process scanner.
desk's session locks catch supervised shells but cannot detect every external
process. If you are unsure, leave the workspace in place.

## 4. Quarantine, do not delete

After stopping those processes:

```sh
desk gc --apply --quiescent
desk list
```

desk rechecks each done workspace and moves eligible directories to quarantine.
Blocked workspaces stay where they are. A successful batch command can still
report retained workspaces; read its output.

To request quarantine for one workspace without marking it done first:

```sh
desk remove ID --quiescent
```

`remove` has the same safety checks. Neither command overrides pins, files,
unpublished work, active sessions, or uncertainty. Nothing has been permanently
deleted at this stage.

## 5. Restore when you need it

```sh
desk restore ID
desk enter ID
```

Restoring moves the directory back, clears its done flag and quarantine age,
and preserves any pin. Use `restore` before entering quarantined work.
Quarantine has no background expiration.

## 6. Purge only when certain

After at least seven days in quarantine, stop any processes still using the
directory and explicitly request permanent removal:

```sh
desk purge ID --quiescent
```

This repeats the retention checks. New files or unavailable remote evidence can
still prevent removal. There is no batch purge, automatic purge, or force flag.
desk cannot restore a purged directory. It is not a backup.
