# desk

A place for the work in between. A small Go CLI that keeps occasional local
development in managed workspaces, out of your way.

**Status:** early and deliberately minimal. macOS and Linux; scratch
directories and Git clones with two-word names; shell integration for zsh,
oh-my-zsh, bash, and fish. The [GitHub repository](https://github.com/markosnarinian/desk)
is private while licensing is finalized. No public binary releases have been published.

## Build and try it

Requires Go 1.24+; repositories also need Git.

```sh
mkdir -p "$HOME/.local/bin"
(cd cli && go build -trimpath -o "$HOME/.local/bin/desk" .)
export PATH="$HOME/.local/bin:$PATH"
eval "$(desk init zsh)"   # or bash; fish: desk init fish | source
```

```sh
desk new                       # empty workspace, e.g. drifting-tundra; cd's there
desk new octocat/Hello-World   # clone, or return to the existing clone
desk list                      # * marks the workspace you're in
desk cd misty                  # by name or unique prefix; no name: workspace top
desk path misty                # print the path, for scripts
desk rm                        # current workspace (or NAME), after confirmation
```

Inside a workspace your prompt starts with `(drifting-tundra/src/api)`: the
workspace and your directory inside it. Commands that take a NAME default to
the workspace you're in.

Workspaces are plain directories in `~/.desk/workspaces/<name>` (`DESK_HOME`
overrides `~/.desk`). There is no registry, background process, or automatic
cleanup. A repository workspace's source is its Git origin, so `desk new`
recognizes the same repository however its URL is spelled. See the
[shell integration](site/content/docs/shell-integration.mdx) and
[configuration](site/content/docs/configuration.mdx) docs.

## Development and documentation

- `cli/`: Go executable and tests; no external Go dependencies. Shell
  integration scripts live in `cli/shell/`, embedded and printed by `desk init`.
- `site/`: Next.js + Chakra UI website: landing page, documentation, and a
  command reference, written in MDX.
- `scripts/release.sh`: local release archives and SHA-256 checksums.

```sh
(cd cli && go test -race ./... && go vet ./...)
npm --prefix site ci
npm --prefix site run check
npm --prefix site run build
npm --prefix site run dev
```

The site uses Node 26 in CI and builds to a static export in `site/out`. Use
`npm --prefix site run preview` to serve the built site, including its search.
Canonical URLs use `SITE_URL` or Vercel's production URL. See
[website maintenance](docs/website.md) for content layout and Vercel settings.

`.agents/setup` installs the pinned Go toolchain and builds the CLI and site in
Linux amd64 Amp orbs. It needs no secrets. Run it from the repository root.

## Distribution, without publication

```sh
scripts/release.sh v0.1.0
```

This creates `dist/desk_v0.1.0_{darwin,linux}_{amd64,arm64}.tar.gz` and
`dist/SHA256SUMS`. Each archive contains `desk`. Check the checksum
before extracting; put the executable in a directory on PATH. These binaries are
not signed or notarized. Source builds are the current installation path.

The GitHub Actions workflow tests on Linux and macOS, checks/builds the site,
and builds all four binaries as downloadable CI artifacts. It does not publish
releases or deploy the site. Website deployment uses the Vercel `desk` project;
see [website maintenance](docs/website.md) for GitHub connection and build settings.
Release publication and signing are separate decisions.

## Deliberately absent, for now

Windows, editor and Amp integrations, automatic pulls, and any automatic cleanup.
