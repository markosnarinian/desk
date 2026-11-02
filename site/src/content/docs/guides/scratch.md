---
title: Scratch workspaces
description: Try an idea, resume it later, and keep useful files without cluttering your home directory.
---

Use a scratch workspace when you need a directory, not a repository: reproduce a
bug, try a package, or sketch a script. desk manages the directory. You choose
your tools and install dependencies yourself.

## Start an experiment

```sh
desk scratch json-experiment
```

This prints a workspace ID and enters your interactive shell in the new directory.
The name is a label, not a path. Creating another scratch workspace with the same
name creates a separate directory and ID.

Inside that shell, try a small example:

```sh
printf '{"answer":42}\n' > sample.json
cat sample.json
printf 'Workspace: %s\n' "$DESK_WORKSPACE"
exit
```

`exit` returns to your original shell and directory. It does **not** mark the
workspace done or delete `sample.json`.

## Resume tomorrow

```sh
desk list
desk enter ID
```

Replace `ID` with the full 12-character ID from the listing. The file is still
there. `desk enter` clears an earlier done flag, so resumed work is no longer a
batch-cleanup candidate.

For an experiment you want to protect explicitly, exit the workspace and pin it:

```sh
desk keep ID
```

A pin blocks quarantine and purge. It is not a backup.

## Keep the result, not the directory

Copy anything valuable to a location you own, verify that copy, and only then
remove the original files you no longer need. For example, from inside the
workspace, with your chosen destination directory already created:

```sh
cp -i sample.json "$HOME/Documents/sample.json"
cmp sample.json "$HOME/Documents/sample.json"
```

Check the destination first so you do not overwrite an existing file. desk will
retain a scratch workspace as long as **any** entry remains, including dotfiles
or empty subdirectories. It does not guess which files are disposable.

When the directory is empty, leave the shell and follow
[cleanup and recovery](/guides/cleanup/). Never empty a workspace just to make a
cleanup check pass if you still need its contents.

## Create without entering

For scripts or when you want to enter later:

```sh
id=$(desk scratch api-experiment --no-enter)
desk keep "$id"
desk enter "$id"
```

`--no-enter` prints only the new ID to standard output. Use `desk list` to find
its filesystem path. desk does not provide a shell function or change the
directory of your calling shell.
