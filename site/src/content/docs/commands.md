---
title: Command reference
description: The commands and options implemented by desk.
---

Run `desk help` for the built-in summary. Commands use full 12-character IDs from
`desk list`, not labels or ID prefixes. Flags may appear before or after positional
arguments. Unknown flags and excess arguments fail.

| Command | Behavior |
| --- | --- |
| `desk scratch [NAME] [--no-enter]` | Create an empty workspace, default label `scratch`, and enter it. `--no-enter` only prints the ID. |
| `desk open REPOSITORY [--no-enter]` | Clone or resume the exact normalized source. GitHub `owner/repo` becomes `https://github.com/owner/repo.git`. Accepts HTTPS, `ssh://`, `git@…`, `file://`, and explicit `/`, `./`, or `../` local paths. Local clones have independent object stores. |
| `desk enter ID` | Enter the workspace in an interactive subshell. Clears its done flag. Quarantined workspaces must be restored first. |
| `desk list` | Show IDs, labels, state, first retention reason, and filesystem paths. Repository checks contact origin and can take time; uncertain state is retained. |
| `desk keep ID` | Pin a workspace, including a quarantined one. |
| `desk keep ID --off` | Unpin. Does not mark done. |
| `desk done ID` | Mark an unpinned, non-quarantined workspace as a cleanup candidate. Does not exit a shell or delete anything. |
| `desk gc [--dry-run]` | Explain the first retention reason or proposed quarantine for each workspace. The default is a dry run. |
| `desk gc --apply --quiescent` | Recheck and quarantine eligible, done workspaces. Retained workspaces are skipped, not treated as command failures. Cannot combine `--apply` with `--dry-run`. |
| `desk remove ID --quiescent` | Recheck and quarantine one eligible workspace, even if not marked done. Cannot bypass pins or safety checks. |
| `desk restore ID` | Restore from quarantine and clear done and quarantine age. Pins are preserved. |
| `desk purge ID --quiescent` | Permanently remove an eligible quarantined workspace after seven days. Repeats all retention checks. |
| `desk version` | Print the build version. Source builds default to `dev`. |

## Find a worked example

- [`scratch`, `enter`, and `keep`](/guides/scratch/)
- [`open` and Git authentication](/guides/repositories/)
- [`done`, `gc`, `remove`, `restore`, and `purge`](/guides/cleanup/)
- [Data directories, shell selection, and environment variables](/configuration/)

When a command retains a workspace, start with [troubleshooting](/troubleshooting/)
and the [cleanup guarantees](/safety/).

## Exit status

Successful operations exit 0. Invalid arguments, filesystem errors, and refused
single-workspace removal exit 1. Subshell failures preserve a positive shell exit
code. A successful dry run or batch cleanup can report retained workspaces and
still exit 0. `list` also reports retention issues rather than failing the whole
listing. Read its output before taking action.

## Not implemented

No Windows support, shell integration functions, background cleanup, age-based
collection, worktree creation, or forced deletion. There is no `desk cd`: a child
process cannot change your calling shell's directory.
