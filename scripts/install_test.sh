#!/bin/sh
# Offline installer contract tests: real tarballs, controlled curl and uname.
set -eu
root=$(CDPATH= cd -P "$(dirname "$0")/.." && pwd)
installer=$root/site/public/install.sh
tmp=$(mktemp -d "${TMPDIR:-/tmp}/desk-install-test.XXXXXXXX")
trap 'rm -rf "$tmp"' 0
trap 'exit 1' HUP INT TERM
original_path=$PATH
mkdir -p "$tmp/bin" "$tmp/source" "$tmp/releases/download/v1.2.3" "$tmp/releases/latest/download"
# A restricted PATH also lets us exercise macOS's shasum-only environment.
for tool in sh awk grep mkdir mktemp rm tar gzip chmod mv cp cat ln; do
  ln -s "$(command -v "$tool")" "$tmp/bin/$tool"
done
if command -v sha256sum >/dev/null 2>&1; then
  hash_tool=sha256sum
else
  hash_tool=shasum
fi
real_hash=$(command -v "$hash_tool")
real_shasum=$(command -v shasum || :)
hash() {
  if [ "$hash_tool" = sha256sum ]; then "$real_hash" "$1"; else "$real_hash" -a 256 "$1"; fi
}
# Both checksum command paths calculate real hashes, even on hosts with only one.
cat > "$tmp/bin/shasum" <<'EOF'
#!/bin/sh
printf 'shasum\n' >> "$TEST_HASH_LOG"
[ "$1" = -a ] && [ "$2" = 256 ] || exit 2
shift 2
if [ -n "$TEST_REAL_SHASUM" ]; then
  exec "$TEST_REAL_SHASUM" -a 256 "$@"
elif [ "$TEST_HASH_TOOL" = sha256sum ]; then
  exec "$TEST_REAL_HASH" "$@"
else
  exec "$TEST_REAL_HASH" -a 256 "$@"
fi
EOF
if [ "$hash_tool" = sha256sum ]; then
  ln -s "$real_hash" "$tmp/bin/sha256sum"
fi
cat > "$tmp/bin/uname" <<'EOF'
#!/bin/sh
case $1 in -s) printf '%s\n' "$TEST_OS" ;; -m) printf '%s\n' "$TEST_ARCH" ;; *) exit 2 ;; esac
EOF
cat > "$tmp/bin/curl" <<'EOF'
#!/bin/sh
set -eu
out=
while [ "$#" -gt 0 ]; do
  case $1 in
    -o) out=$2; shift 2 ;;
    --retry|--connect-timeout) shift 2 ;;
    -*) shift ;;
    *) url=$1; shift ;;
  esac
done
printf '%s\n' "$url" >> "$TEST_DOWNLOAD_LOG"
case $url in
  "$TEST_BASE"/*) path=${url#"$TEST_BASE"/} ;;
  *) printf 'Unexpected URL: %s\n' "$url" >&2; exit 22 ;;
esac
case $path in
  *"${TEST_FAIL_DOWNLOAD:-never-match}"*) printf partial > "$out"; exit 22 ;;
esac
cp "$TEST_RELEASES/$path" "$out"
EOF
cat > "$tmp/source/desk" <<'EOF'
#!/bin/sh
case ${1-} in
  init)
    printf '%s\n' "$0" >> "$TEST_INIT_LOG"
    # Would swallow the rest of a piped installer if stdin were not protected.
    cat >/dev/null
    exit "${TEST_INIT_STATUS:-0}"
    ;;
  version) printf '%s version\n' "$0" >> "$TEST_VERSION_LOG"; printf 'v1.2.3\n' ;;
  *) exit 2 ;;
esac
EOF
printf 'Fixture license\n' > "$tmp/source/LICENSE"
chmod +x "$tmp/bin/curl" "$tmp/bin/uname" "$tmp/bin/shasum" "$tmp/source/desk"
for os in linux darwin; do
  for arch in amd64 arm64; do
    name=desk_v1.2.3_${os}_${arch}.tar.gz
    tar -czf "$tmp/releases/download/v1.2.3/$name" -C "$tmp/source" desk LICENSE
    (cd "$tmp/releases/download/v1.2.3" && hash "$name") >> "$tmp/manifest"
  done
done
cp "$tmp/releases/download/v1.2.3/"* "$tmp/releases/latest/download/"
export TEST_REAL_HASH="$real_hash" TEST_HASH_TOOL="$hash_tool"
export TEST_REAL_SHASUM="$real_shasum" TEST_VERSION_LOG="$tmp/version.log"
export TEST_RELEASES="$tmp/releases" TEST_HASH_LOG="$tmp/hash.log"
export TEST_DOWNLOAD_LOG="$tmp/download.log" TEST_INIT_LOG="$tmp/init.log"
export PATH="$tmp/bin" TEST_OS=Linux TEST_ARCH=x86_64
export DESK_RELEASE_BASE_URL=https://mirror.invalid/releases
export TEST_BASE=$DESK_RELEASE_BASE_URL
count=0
reset() {
  rm -rf "$tmp/home" "$tmp/custom bin"
  mkdir -p "$tmp/home"
  export HOME="$tmp/home"
  unset DESK_VERSION DESK_INSTALL_DIR TEST_FAIL_DOWNLOAD TEST_INIT_STATUS
  TEST_OS=Linux; TEST_ARCH=x86_64
  export DESK_RELEASE_BASE_URL=https://mirror.invalid/releases
  TEST_BASE=$DESK_RELEASE_BASE_URL
  cp "$tmp/manifest" "$tmp/releases/download/v1.2.3/SHA256SUMS"
  cp "$tmp/manifest" "$tmp/releases/latest/download/SHA256SUMS"
  : > "$TEST_DOWNLOAD_LOG"; : > "$TEST_INIT_LOG"; : > "$TEST_HASH_LOG"; : > "$TEST_VERSION_LOG"
}
die() { printf 'FAIL: %s\n' "$*" >&2; cat "$tmp/output" >&2; exit 1; }
pass() { count=$((count + 1)); printf 'ok %s - %s\n' "$count" "$*"; }
run() { sh "$installer" "$@" > "$tmp/output" 2>&1; }
installed() {
  destination=${DESK_INSTALL_DIR:-$HOME/.local/bin}
  destination=$(CDPATH= cd -P "$destination" && pwd -P)
  [ "$("$destination/desk" version)" = v1.2.3 ] || die 'installed version'
  [ "$(cat "$TEST_VERSION_LOG")" = "$destination/desk version" ] || die 'version invocation not recorded'
  [ "$(cat "$destination/desk.LICENSE")" = 'Fixture license' ] || die 'license missing'
  [ "$(cat "$TEST_INIT_LOG")" = "$destination/desk" ] || die 'init was not invoked by absolute path'
  grep -q 'Open a new terminal' "$tmp/output" || die 'missing next-terminal instructions'
  for stage in "$destination"/.desk-install.*; do
    [ ! -e "$stage" ] || die 'staging directory leaked'
  done
}
old_binary() {
  mkdir -p "$HOME/.local/bin"
  printf 'existing binary\n' > "$HOME/.local/bin/desk"
  printf 'existing license\n' > "$HOME/.local/bin/desk.LICENSE"
}
unchanged() {
  [ "$(cat "$HOME/.local/bin/desk")" = 'existing binary' ] || die 'overwrote existing binary'
  [ "$(cat "$HOME/.local/bin/desk.LICENSE")" = 'existing license' ] || die 'overwrote existing license'
  [ ! -s "$TEST_INIT_LOG" ] || die 'ran init after failure'
  for stage in "$HOME/.local/bin"/.desk-install.*; do
    [ ! -e "$stage" ] || die 'failure leaked staging directory'
  done
}
reset
unset DESK_RELEASE_BASE_URL
TEST_BASE=https://github.com/markosnarinian/desk/releases
run || die 'default latest install'
installed
grep -q '/download/v1.2.3/desk_v1.2.3_linux_amd64.tar.gz$' "$TEST_DOWNLOAD_LOG" || die 'archive was not pinned to the manifest version'
pass 'default GitHub stable latest, license, absolute init, cleanup'

for platform in Linux:x86_64 Linux:amd64 Linux:aarch64 Linux:arm64 Darwin:x86_64 Darwin:amd64 Darwin:aarch64 Darwin:arm64; do
  reset
  TEST_OS=${platform%:*}; TEST_ARCH=${platform#*:}
  export DESK_VERSION=v1.2.3 DESK_INSTALL_DIR="$tmp/custom bin"
  run || die "$platform install with spaces"
  installed
  case $TEST_OS in Linux) os=linux ;; Darwin) os=darwin ;; esac
  case $TEST_ARCH in x86_64|amd64) arch=amd64 ;; *) arch=arm64 ;; esac
  grep -q "/download/v1.2.3/desk_v1.2.3_${os}_${arch}.tar.gz$" "$TEST_DOWNLOAD_LOG" || die 'wrong target URL'
  pass "$platform explicit version and directory with spaces"
done

reset
ln -s "$HOME" "$tmp/linked home"
export DESK_INSTALL_DIR="$tmp/linked home/.local/bin"
run || die 'install through directory symlink'
installed
pass 'symlinked installation directory uses canonical absolute init path'

reset
export DESK_VERSION=v1.2.3
run --no-init || die '--no-init'
[ ! -s "$TEST_INIT_LOG" ] || die '--no-init ran init'
grep -q 'Shell setup skipped' "$tmp/output" || die 'missing manual instructions'
pass '--no-init'

reset
{ cat "$installer"; printf '\nprintf survived > "$HOME/pipeline-survived"\n'; } | sh > "$tmp/output" 2>&1 || die 'piped installer'
installed
[ "$(cat "$HOME/pipeline-survived")" = survived ] || die 'init consumed piped script stdin'
pass 'installer piped into sh preserves remaining stdin'

for bad in '../v1.2.3' 'v1.2.3/../../escape' '1.2.3' 'v1.2' 'v1.2.3-rc1' '' 'v1.2.3
../../escape'; do
  reset; old_binary
  export DESK_VERSION=$bad
  if run; then die "accepted invalid version $bad"; fi
  unchanged
  [ ! -s "$TEST_DOWNLOAD_LOG" ] || die 'invalid version made network request'
  pass "reject invalid version: $bad"
done

for platform in Windows:x86_64 Linux:riscv64; do
  reset; old_binary
  TEST_OS=${platform%:*}; TEST_ARCH=${platform#*:}
  if run; then die "accepted $platform"; fi
  unchanged
  [ ! -s "$TEST_DOWNLOAD_LOG" ] || die 'unsupported platform made network request'
  pass "reject unsupported $platform"
done

name=desk_v1.2.3_linux_amd64.tar.gz
for problem in mismatch missing prefix duplicate malformed download-manifest download-archive unpack missing-license symlink; do
  reset; old_binary
  export DESK_VERSION=v1.2.3
  manifest=$tmp/releases/download/v1.2.3/SHA256SUMS
  archive=$tmp/releases/download/v1.2.3/$name
  # Restore the pristine archive before every destructive fixture.
  cp "$tmp/releases/latest/download/$name" "$archive"
  case $problem in
    mismatch) printf '%064d  %s\n' 0 "$name" > "$manifest" ;;
    missing) : > "$manifest" ;;
    prefix) hash_value=$(hash "$archive"); printf '%s.extra\n' "$hash_value" > "$manifest" ;;
    duplicate) cat "$tmp/manifest" "$tmp/manifest" > "$manifest" ;;
    malformed) printf 'not-a-hash  %s\n' "$name" > "$manifest" ;;
    download-manifest) export TEST_FAIL_DOWNLOAD=SHA256SUMS ;;
    download-archive) export TEST_FAIL_DOWNLOAD=.tar.gz ;;
    unpack) printf 'not a tarball' > "$archive"; (cd "$tmp/releases/download/v1.2.3" && hash "$name") > "$manifest" ;;
    missing-license) tar -czf "$archive" -C "$tmp/source" desk; (cd "$tmp/releases/download/v1.2.3" && hash "$name") > "$manifest" ;;
    symlink)
      mkdir -p "$tmp/symlink"
      ln -s "$tmp/source/desk" "$tmp/symlink/desk"
      cp "$tmp/source/LICENSE" "$tmp/symlink/LICENSE"
      tar -czf "$archive" -C "$tmp/symlink" desk LICENSE
      (cd "$tmp/releases/download/v1.2.3" && hash "$name") > "$manifest"
      ;;
  esac
  if run; then die "accepted $problem"; fi
  unchanged
  pass "$problem preserves existing binary and license"
done

for problem in missing duplicate unsafe-version; do
  reset; old_binary
  manifest=$tmp/releases/latest/download/SHA256SUMS
  case $problem in
    missing) : > "$manifest" ;;
    duplicate) cat "$tmp/manifest" "$tmp/manifest" > "$manifest" ;;
    unsafe-version) printf '%064d  desk_v1.2.3_../../bad_linux_amd64.tar.gz\n' 0 > "$manifest" ;;
  esac
  if run; then die "accepted latest $problem"; fi
  unchanged
  pass "latest $problem manifest preserves existing install"
done

reset
export DESK_VERSION=v1.2.3
# A valid entry plus a near-match must select only the exact filename.
cp "$tmp/releases/latest/download/$name" "$tmp/releases/download/v1.2.3/$name"
printf '%064d  %s.extra\n' 0 "$name" >> "$tmp/releases/download/v1.2.3/SHA256SUMS"
run || die 'exact checksum with near-match'
installed
pass 'exact checksum entry ignores a near-match'

reset; old_binary
run || die 'replacing existing binary'
installed
pass 'successful upgrade replaces existing binary and retains release license'

reset
rm -f "$tmp/bin/sha256sum"
TEST_OS=Darwin; TEST_ARCH=arm64
run || die 'shasum fallback'
installed
grep -q shasum "$TEST_HASH_LOG" || die 'did not use shasum fallback'
pass 'macOS shasum fallback verifies real checksum'

reset
export TEST_INIT_STATUS=7
if run; then die 'ignored init failure'; fi
[ -x "$HOME/.local/bin/desk" ] || die 'init failure lost installed binary'
grep -q 'shell setup failed' "$tmp/output" || die 'unclear init failure'
pass 'init failure reports successful install and manual recovery'

reset
if run --unknown; then die 'unknown option accepted'; fi
[ ! -s "$TEST_DOWNLOAD_LOG" ] || die 'invalid option downloaded'
pass 'unknown option rejected before downloads'

reset; old_binary
rm -f "$tmp/bin/shasum"
if run; then die 'installed without checksum tooling'; fi
unchanged
[ ! -s "$TEST_DOWNLOAD_LOG" ] || die 'downloaded without checksum tooling'
pass 'missing checksum tooling fails before downloads'
PATH=$original_path
printf 'Passed %s installer tests (offline, isolated HOME).\n' "$count"
