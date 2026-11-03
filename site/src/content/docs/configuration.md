---
title: Storage and configuration
description: Find managed directories, choose a separate data root, and understand shell and Git configuration.
---

desk keeps managed data under `~/.desk` and reserves `~/.config/desk/` for
configuration on both Linux and macOS. It currently has no configuration file or
daemon, so it does not create or read the configuration directory. Settings are
supplied through a small set of environment variables.

## Find your workspaces

```sh
desk list
```

The `PATH` column is the actual directory location, including quarantined
workspaces. Names are labels; generated 12-character IDs identify workspaces.

The default data root is `~/.desk` on both Linux and macOS. `XDG_DATA_HOME`
does not change this location.

The root contains `workspaces/`, `quarantine/`, `locks/`, and a JSON registry.
Do not edit the registry, replace workspace directories with symlinks, or share
a root between machines. Moving the root during a session is unsupported.

## Choose a separate root

`DESK_HOME` overrides the data root on both platforms. It must be absolute.
For example, use a separate persistent root to try desk without mixing in your
normal workspaces:

```sh
export DESK_HOME="$HOME/.desk-trial"
desk scratch trial
```

Keep the export in effect when you list, resume, or clean up that trial. Each
root has its own registry. Changing the variable does **not** move or delete
existing workspaces. `unset DESK_HOME` returns to `~/.desk`; the
trial remains in its separate root.

### Existing installations

Earlier versions defaulted to `$XDG_DATA_HOME/desk` (or `~/.local/share/desk`)
on Linux and `~/Library/Application Support/desk` on macOS. desk does not
automatically migrate or delete those directories. Set `DESK_HOME` to the
absolute path of your old root to continue using it.

To adopt the new default, first stop all desk sessions and other processes using
the workspaces, then move the entire old root to `~/.desk` only if that destination
does not already exist. Do not merge registries or move individual workspaces.
Unset `DESK_HOME` and check `desk list` afterward.

Do not put resumable work in OS-managed temporary directories such as `/tmp`.
Another tool or a reboot may remove it independently of desk's safety checks.
Never run desk with `sudo`.

## Shell behavior

| Variable | Meaning |
| --- | --- |
| `SHELL` | Executable used for the interactive subshell; falls back to `/bin/sh` |
| `DESK_WORKSPACE` | Set inside a desk subshell to its workspace ID |
| `DESK_HOME` | Optional absolute data root |

To choose a shell for one entry, with that executable installed:

```sh
SHELL=/bin/bash desk enter ID
```

`SHELL` must be an executable path, not a command containing arguments. desk runs
it with `-i` and the workspace as its initial directory. Shell startup files may
change that directory. Your original shell stays outside; `exit` returns to it.
desk does not inject a prompt or install shell integration functions.

## Git behavior

Git's credential helpers, SSH configuration, `GIT_SSH`, `GIT_SSH_COMMAND`, and
`GIT_ASKPASS` work normally. Other inherited `GIT_*` variables are removed from
Git subprocesses so they cannot redirect checks to a different checkout.
Git terminal password prompts are disabled.

Repository inspections contact origin and can take time, including during
`desk list`. An unavailable remote means retention, not deletion. desk does not
automatically fetch or pull. See [repository workspaces](/guides/repositories/).

## Backups and uninstalling

Back up valuable files and Git work independently. A workspace pin prevents
desk cleanup, not disk failure or another program deleting files.

Removing the executable leaves all workspaces untouched. Before manually removing
the data root, inspect `desk list` and recover everything you need, including
quarantined directories. Deleting the root yourself bypasses all desk guarantees.
