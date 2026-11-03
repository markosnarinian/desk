#!/usr/bin/env bash
set -euo pipefail

version="${1:?Usage: scripts/release.sh VERSION [OUTPUT_DIRECTORY]}"
if [[ ! "$version" =~ ^[a-zA-Z0-9][a-zA-Z0-9._-]*$ ]]; then
  echo 'Version must contain only letters, digits, dots, underscores, and hyphens.' >&2
  exit 1
fi
root="$(cd "$(dirname "$0")/.." && pwd)"
output="${2:-$root/dist}"
mkdir -p "$output"
output="$(cd "$output" && pwd)"
stage="$(mktemp -d)"
trap 'rm -rf "$stage"' EXIT

for target in darwin/amd64 darwin/arm64 linux/amd64 linux/arm64; do
  os="${target%/*}"
  arch="${target#*/}"
  name="desk_${version}_${os}_${arch}"
  mkdir -p "$stage/$name"
  (
    cd "$root/cli"
    CGO_ENABLED=0 GOOS="$os" GOARCH="$arch" go build \
      -trimpath -buildvcs=false -ldflags "-s -w -X main.version=$version" \
      -o "$stage/$name/desk" .
  )
  tar -czf "$output/$name.tar.gz" -C "$stage/$name" desk
done
(
  cd "$output"
  if command -v sha256sum >/dev/null; then
    sha256sum "desk_${version}_"*.tar.gz > SHA256SUMS
  else
    shasum -a 256 "desk_${version}_"*.tar.gz > SHA256SUMS
  fi
)
echo "Built four archives and SHA256SUMS in $output. Nothing published."
