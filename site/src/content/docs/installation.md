---
title: Installation
description: Build and install desk on macOS or Linux.
---

desk is a standalone Go executable for macOS and Linux, on Intel/AMD 64-bit and
ARM64 machines. Git must be installed to use repositories. There is no daemon,
shell function, or runtime dependency on Go after building.

## Install from this checkout

**No public binary release has been published yet.** The GitHub repository is
currently private while licensing is finalized. With repository access and Git
authentication configured, clone it into a location of your choice:

```sh
git clone https://github.com/markosnarinian/desk.git
cd desk
```

An existing Amp checkout works too. Build with Go 1.23 or newer from the repository root:

```sh
mkdir -p "$HOME/.local/bin"
(cd cli && go build -trimpath -o "$HOME/.local/bin/desk" .)
export PATH="$HOME/.local/bin:$PATH"
desk version
```

Add the PATH export to your shell's startup file to keep it for future terminals.
Never install or run desk with sudo. `desk version` prints `dev` for a source build.

## Release archives

The repository includes a release builder for `darwin` and `linux`, each for
`amd64` and `arm64`. An approved future release can provide these archives and
SHA-256 checksums. Until then, build from source; there is no curl installer,
Homebrew formula, or public download endpoint.

To use an archive built locally, verify it against its accompanying `SHA256SUMS`
(`sha256sum -c` on Linux, or `shasum -a 256 -c` on macOS), extract the archive for
your OS/architecture, then install its `desk` executable in a directory on PATH.

## Uninstall

Remove the executable from the directory where you installed it. That does not
delete workspaces. Use `desk list` and recover anything you need **before** removing
the application-data directory yourself. Read [cleanup guarantees](/safety/) first.
