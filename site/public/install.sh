#!/bin/sh
# Install a verified release without Go or elevated privileges.
set -eu

fail() { printf 'desk: %s\n' "$*" >&2; exit 1; }
no_init=false
case ${1-} in
  '') ;;
  --no-init) no_init=true; shift ;;
  *) fail 'Usage: install.sh [--no-init]' ;;
esac
[ "$#" -eq 0 ] || fail 'Usage: install.sh [--no-init]'

valid_version() {
  case $1 in ''|*[!v0-9.]*) return 1 ;; esac
  printf '%s\n' "$1" | LC_ALL=C grep -Eq '^v[0-9]+\.[0-9]+\.[0-9]+$'
}
version=${DESK_VERSION-}
if [ "${DESK_VERSION+x}" = x ]; then
  valid_version "$version" || fail 'DESK_VERSION must be vX.Y.Z (for example v1.2.3).'
fi
case $(uname -s) in
  Darwin) os=darwin ;;
  Linux) os=linux ;;
  *) fail 'Supported operating systems: macOS and Linux.' ;;
esac
case $(uname -m) in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) fail 'Supported architectures: amd64 (x86_64) and arm64 (aarch64).' ;;
esac
command -v curl >/dev/null 2>&1 || fail 'curl is required.'
if command -v sha256sum >/dev/null 2>&1; then
  checksum() { sha256sum "$1"; }
elif command -v shasum >/dev/null 2>&1; then
  checksum() { shasum -a 256 "$1"; }
else
  fail 'sha256sum or shasum is required to verify the download.'
fi

base=${DESK_RELEASE_BASE_URL:-https://github.com/markosnarinian/desk/releases}
base=${base%/}
if [ -n "$version" ]; then
  url=$base/download/$version
else
  url=$base/latest/download
fi
install_dir=${DESK_INSTALL_DIR:-${HOME:?HOME must be set}/.local/bin}
case $install_dir in /*) ;; *) install_dir=$PWD/$install_dir ;; esac
mkdir -p "$install_dir" || fail "Cannot create $install_dir. Choose a writable DESK_INSTALL_DIR; sudo is not needed."
install_dir=$(CDPATH= cd -P "$install_dir" && pwd -P)
# Stage on the destination filesystem so the final rename is atomic.
stage=$(mktemp -d "$install_dir/.desk-install.XXXXXXXX") || fail 'Cannot create installation staging directory.'
trap 'rm -rf "$stage"' 0
trap 'exit 1' HUP INT TERM
download() {
  curl -fSL --retry 2 --connect-timeout 15 -o "$2" "$1" ||
    fail "Download failed: $1. Check the release exists and your connection; the existing Desk binary was not changed."
}
download "$url/SHA256SUMS" "$stage/SHA256SUMS"
if [ -z "$version" ]; then
  # GitHub's latest endpoint selects the latest stable (not prerelease) release.
  archive=$(LC_ALL=C awk -v suffix="_${os}_${arch}.tar.gz" '
    NF == 2 && $2 ~ /^desk_v[0-9]+\.[0-9]+\.[0-9]+_/ {
      name = $2
      if (substr(name, length(name) - length(suffix) + 1) == suffix) {
        print name
        count++
      }
    }
    END { if (count != 1) exit 1 }
  ' "$stage/SHA256SUMS") || fail 'The latest release manifest must contain exactly one matching Desk archive.'
  version=${archive#desk_}
  version=${version%_${os}_${arch}.tar.gz}
  valid_version "$version" || fail 'The latest release manifest contains an invalid version.'
fi
# Pin the archive to the manifest's version if a new latest release appears.
url=$base/download/$version
archive=desk_${version}_${os}_${arch}.tar.gz
expected=$(LC_ALL=C awk -v name="$archive" '
  NF == 2 && $2 == name { count++; hash = $1 }
  END {
    if (count != 1 || length(hash) != 64 || hash !~ /^[0-9a-fA-F]+$/) exit 1
    print tolower(hash)
  }
' "$stage/SHA256SUMS") || fail "Missing, invalid, or duplicate checksum for exactly $archive."
download "$url/$archive" "$stage/archive.tar.gz"
actual=$(checksum "$stage/archive.tar.gz") || fail 'Could not calculate the archive checksum.'
actual=${actual%% *}
[ "$actual" = "$expected" ] || fail 'Checksum mismatch; the existing Desk binary was not changed.'
mkdir "$stage/unpacked"
tar -xzf "$stage/archive.tar.gz" -C "$stage/unpacked" desk LICENSE ||
  fail 'Could not unpack desk and LICENSE; the existing Desk binary was not changed.'
for file in desk LICENSE; do
  [ -f "$stage/unpacked/$file" ] && [ ! -L "$stage/unpacked/$file" ] ||
    fail "Release $file must be a regular file; the existing Desk binary was not changed."
done
chmod 755 "$stage/unpacked/desk" || fail 'Cannot make Desk executable.'
[ ! -d "$install_dir/desk" ] || fail 'The destination desk is a directory.'
[ ! -d "$install_dir/desk.LICENSE" ] || fail 'The destination desk.LICENSE is a directory.'
mv -f "$stage/unpacked/LICENSE" "$install_dir/desk.LICENSE" || fail 'Cannot install the Desk license.'
mv -f "$stage/unpacked/desk" "$install_dir/desk" || fail 'Cannot install Desk.'
printf 'Installed Desk %s at %s\nLicense: %s\n' "$version" "$install_dir/desk" "$install_dir/desk.LICENSE"
if [ "$no_init" = false ]; then
  # Never let init consume the remainder of a script piped into sh.
  "$install_dir/desk" init </dev/null ||
    fail "Desk was installed, but shell setup failed. Run \"$install_dir/desk\" init manually."
else
  printf 'Shell setup skipped. Run "%s/desk" init when ready.\n' "$install_dir"
fi
printf 'Open a new terminal to use Desk shell integration.\n'
case :${PATH-}: in
  *:"$install_dir":*) ;;
  *) printf 'Add %s to PATH in your shell configuration if it is not already there.\n' "$install_dir" ;;
esac
