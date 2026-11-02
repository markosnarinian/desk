desk manages disposable **directories**, not development environments. It supports
macOS and Linux. It never deletes on shell exit and runs no background cleanup.

## Ownership and storage

The data root is `$DESK_HOME` when set to an absolute path, otherwise
`$XDG_DATA_HOME/desk` on Linux (default `~/.local/share/desk`) and
`~/Library/Application Support/desk` on macOS. The root is private to the user.
desk owns only directories it creates under `workspaces/<id>` and its registry,
locks, and quarantine. It does not adopt existing directories or follow workspace
symlinks. Workspace names are labels; generated IDs are directory names.
Do not edit the registry, replace managed directories, or share the root between
machines. Moving the data root while a session is running is unsupported.

## Sessions

`scratch` creates an empty workspace; `open` clones or resumes a repository;
`enter ID` resumes either kind. Each enters `$SHELL` (or `/bin/sh`) as an
interactive subshell, with `DESK_WORKSPACE` set to its ID. `exit` returns to the
calling shell. Creation also accepts `--no-enter` for scripting.
Session locks block cleanup while desk is supervising a shell. Multiple sessions
are permitted. Exiting is not a request to delete anything.

Locks cannot prove that an editor, detached process, or a shell surviving a killed
desk process has stopped. Before **any** quarantine or purge, the user must supply
`--quiescent`, acknowledging that all processes using the affected workspaces have
stopped. Do not race cleanup with external file writes. desk cannot enforce this
against other applications. These limits are why automatic cleanup is absent.

## Removal and retention

`keep ID` pins; `keep ID --off` unpins. `done ID` marks an unpinned workspace as
a cleanup candidate; it neither exits a shell nor deletes anything. `remove ID`
explicitly requests quarantine of that workspace without requiring `done`.
Neither `done` nor `remove` overrides a pin or a safety check. There is no force
option and no age-based deletion.

`gc` defaults to a dry run. Only `gc --apply --quiescent` moves eligible, done
workspaces to quarantine. An empty scratch workspace is eligible; any file,
including a dotfile, retains it. A repository must:

- be an ordinary independent clone with its own `.git` directory;
- have no tracked changes, untracked files, ignored files, submodules, tracked
  content filters (including Git LFS), or extra worktrees. Separate content stores
  cannot be verified through Git's commit graph alone;
- have every local branch name present on origin, and every local tag present
  there with the exact same object ID (including annotated tag metadata);
- have no local commits reachable through HEAD, any ref, or reflogs that are not
  reachable from a branch or tag **currently advertised by origin**. This includes
  local-only branches, tags, stashes, detached commits, and recoverable history.
  Remote evidence comes from `git ls-remote`, not stale tracking refs. If the remote
  is unavailable, or the evidence cannot be checked locally, retain the workspace.

Git metadata (configuration, hooks, index, reflogs, and other `.git` contents) is
part of the disposable checkout, not a place for personal files. A pin is required
if you need to preserve it. Git checks disable optional index writes and replacement
objects. Checks may be conservative: retaining a safe workspace is preferable to
losing data. Origin is a recovery source, not a guaranteed permanent backup.

## Quarantine and permanent removal

Quarantine is a same-filesystem rename into `quarantine/<id>`. `list` shows it;
`restore ID` returns it to managed storage and clears its done flag. No background
timer empties quarantine. `purge ID --quiescent` permanently removes only an already
quarantined workspace, at least seven days after quarantine. It repeats the safety
checks, so files added in quarantine or a remote that disappeared block purging.
There is no bypass for unsafe contents: recover or move them somewhere you own,
then retry. Never run desk with sudo.

Registry updates use atomic replacement and operations are serialized. Location
is inferred from the actual directory so interruption after a quarantine/restore
rename remains recoverable. desk is not a backup or protection against malicious
local filesystem mutation, disk failure, or simultaneous edits by other tools.
