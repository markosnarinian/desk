# desk

A workspace manager for local development.

As agents move more programming work to the cloud, local work becomes occasional
and fragmented: temporary clones, playgrounds, exploration, experiments,
throwaway scaffolding, and local prototypes. desk makes those workspaces easy to
create, return to, and retire—including bringing Amp orb work locally.

macOS and Linux. Zsh (including Oh My Zsh), bash, and fish.
Licensed under [GNU AGPL v3](LICENSE).

## Install

**Homebrew** — no Go required:

```sh
brew install markosnarinian/tap/desk && desk init
```

**One-liner** — checksum-verified binary and automatic shell setup, no sudo:

```sh
curl -fsSL https://raw.githubusercontent.com/markosnarinian/desk/main/site/public/install.sh | sh
```

**Go** — Go 1.24+:

```sh
GOBIN="$HOME/.local/bin" go install github.com/markosnarinian/desk/cmd/desk@latest
"$HOME/.local/bin/desk" init
```

Open a new terminal after setup. See [installation](site/content/docs/installation.mdx)
for source builds, version pinning, manual downloads, updates, and uninstalling.

## Use

```sh
desk new                       # empty playground with a generated name
desk new owner/repo             # choose an existing workspace or a new clone
desk list                      # tree grouped by repository, plus playgrounds
desk switch misty              # return by name or unique prefix
desk info                      # current workspace's name, path, and source
desk rename local-test         # optional: name it once you know what it's for
desk finish                    # mark it disposable, without removing it
desk cleanup                   # preview eligible inactive workspaces, then confirm
desk rm                        # explicit removal, even with local-only work
```

With a repository, `--fresh` creates an independent clone, `--reuse` remembers
most-recent reuse, and `--choose` overrides that preference once. Returning never
pulls or resets a clone. Workspaces are plain directories under `~/.desk/workspaces`;
`DESK_HOME` overrides the data root. No daemon runs.

**Test Amp work locally** with Git and an authenticated Amp CLI:

```sh
desk amp                       # browse recent tasks
desk amp THREAD --copy         # editable snapshot, including unpushed work
desk amp THREAD --follow       # live updates; test from a separate terminal
```

Don't edit while following. Stopping restores the pre-sync checkout, not the
latest synced files. Dependencies, services, and secrets are not copied.
See [Test Amp work locally](site/content/docs/amp.mdx).

**Removal:** cleanup considers managed workspaces inactive for at least 72 hours,
excluding local-only Git work, ignored files, inspection failures, detectable
processes, and live Amp follow. Scratch workspaces without a remote require
`finish` first; it never bypasses Git checks. `--yes` skips only confirmation.
Explicit `rm` removes whatever you confirm. Both retain garbage for 72 hours;
any later desk invocation can permanently delete expired garbage. Read
[cleanup and retention](site/content/docs/cleanup.mdx) for activity and safety limits.

Start with the [introduction](site/content/docs/introduction.mdx),
[Quickstart](site/content/docs/quickstart.mdx), or
[command reference](site/content/commands/overview.mdx).

## Development

- `cmd/desk/`: executable and tests, with embedded shell integration; no external Go dependencies.
- `site/`: unified Next.js + Chakra UI documentation, written in MDX.
- `scripts/`: distribution packaging, installer tests, and formula generation.

```sh
go test -race ./... && go vet ./...
scripts/install_test.sh
python3 scripts/distribution_test.py
npm --prefix site ci
npm --prefix site run check
npm --prefix site run build
npm --prefix site run preview
```

The site uses Node 26 in CI and exports to `site/out`. See
[website maintenance](docs/website.md) for development and Vercel configuration.
`.agents/setup` installs the pinned Go toolchain and builds the CLI and site in
Linux amd64 Amp orbs without secrets.

## Releases

`scripts/release.sh v0.1.0` builds four macOS/Linux amd64/arm64 archives in `dist/`,
each containing `desk` and `LICENSE`, plus `SHA256SUMS`. Binaries are not signed
or notarized; checksums are not publisher signatures.

Branch pushes test on Linux and macOS, build the site, and retain commit-labelled
binary artifacts for 30 days. Pushing a stable `vX.Y.Z` tag verifies again, then
publishes the archives, checksums, and generated `desk.rb` formula. The release
remains a draft until every asset is uploaded.

Automatic tap updates require the `HOMEBREW_TAP_TOKEN` Actions secret with
contents-write access to `markosnarinian/homebrew-tap`. Without it, publication
can succeed but the tap job fails; use the release's generated formula for a
manual tap update. The installer supports `DESK_VERSION`, `DESK_INSTALL_DIR`, and
`--no-init`; failed downloads or checksums leave existing binaries unchanged.

Windows, editor integrations, automatic pulls, and environment replication are
deliberately out of scope for now.
