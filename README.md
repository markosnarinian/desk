# desk

A place for the work in between. A small Go CLI that keeps occasional local
development in managed workspaces, out of your way.

**Status:** initial implementation. macOS and Linux; independent Git clones and
scratch directories; interactive subshell entry; explicit, conservative cleanup.
The [GitHub repository](https://github.com/markosnarinian/desk) is private while
licensing is finalized. No public binary releases have been published.

## Build and try it

Requires Go 1.23+; repository operations also need Git.

```sh
mkdir -p "$HOME/.local/bin"
(cd cli && go build -trimpath -o "$HOME/.local/bin/desk" .)
export PATH="$HOME/.local/bin:$PATH"
desk scratch experiment
# Inside the workspace: use your tools, then exit to return.
```

```sh
desk open octocat/Hello-World
desk list
desk enter ID
desk keep ID
desk keep ID --off
desk done ID
desk gc --dry-run
```

Replace `ID` with a full ID from `desk list`. `scratch` and `open` also support
`--no-enter` for scripting. `open` resumes the same normalized source without
pulling or resetting. `enter` clears the workspace's done flag.

## Cleanup is a separate decision

**Git-clean does not mean disposable.** Read the [safety contract](docs/safety.md).
desk retains ignored and untracked files, local-only history, local-only branch
and tag names, pins, active desk sessions, and uncertain state. It checks origin's
current advertised refs, not only cached tracking branches.

`gc` defaults to a dry run. Stop all shells, editors, servers, and detached
processes using the affected workspaces before using `--quiescent`:

```sh
desk gc --apply --quiescent
desk remove ID --quiescent
desk restore ID
desk purge ID --quiescent
```

`remove` and applied `gc` quarantine; they do not permanently delete. `purge`
requires seven days in quarantine and repeats safety checks. Nothing happens
automatically on shell exit. There is no force flag. Session locks cannot detect
every external process, which is why your quiescence acknowledgment is required.

Storage defaults to `~/.local/share/desk` on Linux or
`~/Library/Application Support/desk` on macOS. `XDG_DATA_HOME` is respected on
Linux; `DESK_HOME` overrides the root on either OS and must be absolute. Do not use
OS-managed temporary storage for work you want to resume. Test fixtures use
temporary storage only to isolate disposable tests.

## Development and documentation

- `cli/`: Go executable and tests; no external Go dependencies.
- `site/`: Astro + Starlight landing page, installation, quickstart, task guides,
  command reference, configuration, troubleshooting, and cleanup guarantees.
  The site imports `docs/safety.md` directly.
- `scripts/release.sh`: local release archives and SHA-256 checksums.

```sh
(cd cli && go test -race ./... && go vet ./...)
npm --prefix site ci
npm --prefix site run check
npm --prefix site run build
npm --prefix site run dev
```

The site uses Node 26 in CI and follows the official
[Astro](https://docs.astro.build/en/install-and-setup/) and
[Starlight](https://starlight.astro.build/manual-setup/) setup. Search defaults to
Pagefind until an Algolia DocSearch index and search-only credentials are supplied.
Use `npm --prefix site run preview` to test the built search. Canonical URLs use
`SITE_URL` or Vercel's production URL; without either, sitemap generation is skipped.
See [website maintenance](docs/website.md) for Vercel and Algolia configuration.
The current Astro/Starlight combination also emits a harmless duplicate `/404`
route warning; the built `404.html` is checked in the browser.

`.agents/setup` installs the pinned Go toolchain and builds the CLI and site in
Linux amd64 Amp orbs. It needs no secrets. Run it from the repository root.

## Distribution, without publication

```sh
scripts/release.sh v0.1.0
```

This creates `dist/desk_v0.1.0_{darwin,linux}_{amd64,arm64}.tar.gz` and
`dist/SHA256SUMS`. Each archive contains `desk` and `SAFETY.md`. Check the checksum
before extracting; put the executable in a directory on PATH. These binaries are
not signed or notarized. Source builds are the current installation path.

The GitHub Actions workflow tests on Linux and macOS, checks/builds the site,
and builds all four binaries as downloadable CI artifacts. It does not publish
releases or deploy the site. Website deployment uses the Vercel `desk` project;
see [website maintenance](docs/website.md) for GitHub connection and build settings.
Release publication and signing are separate decisions.

## Deliberately absent

Windows, shell functions, background cleanup, worktree creation/optimization,
automatic pulls, and age-based workspace collection.
