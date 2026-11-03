# desk

A workspace for the local part.

As agents take more programming work to the cloud, local work becomes more
occasional and fragmented: a temporary clone, a playground, a throwaway
scaffold, a local prototype.

desk gives each a workspace with a clear start and finish. Create one, explore
with your usual tools, return when needed, and remove it when you're done.

**Status:** early and deliberately minimal. macOS and Linux; scratch
directories and Git clones with two-word names; automatic shell setup for zsh
(including Oh My Zsh), bash, and fish. Licensed under [GNU AGPL v3](LICENSE).
No public binary releases have been published.

## Install

The release paths below are prepared for the first release. Until it is published,
use the source build below.

**Homebrew** — macOS and Linux, no Go required:

```sh
brew install markosnarinian/tap/desk && desk init
```

**One-liner** — verified binary download and automatic shell setup, no sudo:

```sh
curl -fsSL https://raw.githubusercontent.com/markosnarinian/desk/main/site/public/install.sh | sh
```

**Go** — Go 1.24+, without cloning:

```sh
GOBIN="$HOME/.local/bin" go install github.com/markosnarinian/desk/cmd/desk@latest
"$HOME/.local/bin/desk" init
```

**From source** — Go 1.24+ and Git:

```sh
git clone https://github.com/markosnarinian/desk.git
cd desk
mkdir -p "$HOME/.local/bin"
go build -trimpath -o "$HOME/.local/bin/desk" ./cmd/desk
"$HOME/.local/bin/desk" init
```

Open a new terminal. `desk init` detects your shell and updates its startup
configuration, including Oh My Zsh. Or choose `desk init zsh`, `desk init bash`,
or `desk init fish`. The one-liner and Go examples install into `~/.local/bin`.
See [installation](site/content/docs/installation.mdx) for version pinning,
manual downloads, updating, and uninstalling.

## Use it for real work

Try a small script (requires Node.js), then clear it away:

```sh
desk new
printf 'console.log(2 ** 10)\n' > try.js
node try.js                    # 1024
desk rm                        # asks before deleting; takes you home
```

Need a repository locally? `desk new owner/repo` offers existing workspaces
and a new clone. Use `--fresh` for an independent clone, `--reuse` to remember
a per-repository preference for the most recently used workspace, or `--choose`
to override that preference just this time. No task name is required.

To test Amp work locally, paste a thread URL or ID into `desk amp THREAD`, or
run `desk amp` to browse recent tasks. Choose copy to keep an editable snapshot,
or follow to watch live changes in a fresh clone. This includes uncommitted and
unpushed work without pushing. Requires Git and an installed, authenticated
Amp CLI; dependencies, services, and secrets are not copied.

```sh
desk amp THREAD --copy          # keep a snapshot to edit locally
desk amp THREAD --follow        # foreground; test from a separate terminal
```

While following, use `desk switch NAME` in another terminal and don't edit
locally. Stopping follow restores the pre-sync checkout. See the
[Amp guide](site/content/docs/amp.mdx).

```sh
desk list                      # tree grouped by repository, plus scratch
desk switch misty              # name or unique prefix from your list
desk switch                    # back to the current workspace's top directory
desk info                      # current workspace's name, path, and source
code "$(desk info misty --path)" # open a workspace in another tool
desk rename local-test         # rename the current workspace when useful
desk finish                    # mark the current workspace disposable
desk cleanup                   # preview eligible workspaces, then confirm all
desk rm misty                  # remove explicitly, after confirmation
```

Names start random; `misty` is an example. `desk cleanup` only considers managed
workspaces inactive for at least 72 hours, with no local-only work or detectable
running processes/live Amp follow. Git changes and ignored files (including in
nested repositories), stashes, unpushed commits (including other branches, tags,
and detached HEAD), or failed Git inspection prevent cleanup. Scratch workspaces without a remote need
`desk finish` first; finish never bypasses Git safety checks. `--yes` skips only
confirmation. Cleanup uses locally known remote-tracking refs without fetching;
it cannot guarantee the remote server still has the work.

Pre-upgrade directories without state stay outside cleanup until you explicitly
use them through `desk switch`, select them with `new`, or mark them finished.

Activity comes from saved file/directory changes, recorded new/switch visits,
and detectable processes—not every editor keystroke or unsaved buffer.
Both cleanup and explicit removal move workspaces to `~/.desk/garbage` for
72 hours from removal. Every desk invocation silently collects expired garbage;
there is no daemon. Explicit `desk rm` still asks even when local-only work exists.
See the [cleanup guide](site/content/docs/cleanup.mdx).

Inside a workspace your prompt starts with `(drifting-tundra/src/api)`: the
workspace and your directory inside it. Commands that take a NAME default to
the workspace you're in.

Workspaces are plain directories in `~/.desk/workspaces/<name>` (`DESK_HOME`
overrides `~/.desk`). Workspace metadata lives outside project files in `state/NAME.json`,
with reuse preferences in `preferences.json`. There is no background process.
A repository workspace's source is its Git origin, so `desk new`
recognizes the same repository however its URL is spelled. See the
[shell integration](site/content/docs/shell-integration.mdx) and
[configuration](site/content/docs/configuration.mdx) docs.

## Development and documentation

- `cmd/desk/`: Go executable and tests; no external Go dependencies. Shell
  integration scripts live in `cmd/desk/shell/`, embedded and printed by `desk init --print`.
- `site/`: Next.js + Chakra UI website: landing page, documentation, and a
  command reference, written in MDX.
- `scripts/release.sh`: local release archives and SHA-256 checksums.

```sh
go test -race ./... && go vet ./...
scripts/install_test.sh
python3 scripts/distribution_test.py
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

## Distribution

```sh
scripts/release.sh v0.1.0
```

This creates `dist/desk_v0.1.0_{darwin,linux}_{amd64,arm64}.tar.gz` and
`dist/SHA256SUMS`. Each archive contains `desk` and `LICENSE`. Check the checksum
before extracting; put the executable in a directory on PATH. These binaries are
not signed or notarized. Source builds are the current installation path.

The GitHub Actions workflow tests on Linux and macOS, checks/builds the site,
and independently builds all four binaries on every branch push. Commit-labelled
CI artifacts are retained for 30 days; normal pushes do not publish a stable release.

Pushing a stable `vX.Y.Z` tag runs verification before publishing archives,
checksums, and a generated `desk.rb` Homebrew formula. The release stays a draft
until all assets are uploaded. The tap update runs only after publication.
Before releasing, set the `HOMEBREW_TAP_TOKEN` GitHub Actions secret with
contents-write access limited to `markosnarinian/homebrew-tap`. Without it,
release publication can succeed but the tap-update job fails explicitly. The
generated formula is also attached to the release for manual tap updates.

`site/public/install.sh` is served as `/install.sh` and can also be fetched from
GitHub. It verifies the selected archive before replacing a binary and supports
`DESK_VERSION`, `DESK_INSTALL_DIR`, and `--no-init`.

Website deployment uses the Vercel `desk` project; see
[website maintenance](docs/website.md). Release publication, tap pushes, and
signing require deliberate action; configuring the workflows publishes nothing.

## Deliberately absent, for now

Windows, editor integrations, automatic pulls, and environment replication.
