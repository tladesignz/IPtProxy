#!/bin/sh

TARGET=ios,iossimulator,macos
OUTPUT=IPtProxy.xcframework
TEMPDIR="$TMPDIR/IPtProxy"

if test "$1" = "android"; then
  TARGET=android
  OUTPUT=IPtProxy.aar
fi

cd "$(dirname "$0")" || exit 1

if test -e $OUTPUT; then
    echo "--- No build necessary, $OUTPUT already exists."
    exit
fi

# Install dependencies. Go itself is a prerequisite.
printf '\n--- Golang 1.21 or up needs to be installed! Try "brew install go" on MacOS or "snap install go" on Linux if we fail further down!'
printf '\n--- Installing gomobile...\n'
go install golang.org/x/mobile/cmd/gomobile@latest

# Prepare build environment
printf '\n\n--- Prepare build environment at %s...\n' "$TEMPDIR"
CURRENT=$PWD
rm -rf "$TEMPDIR"
mkdir -p "$TEMPDIR"
cp -a IPtProxy.go "$TEMPDIR/"


# Fetch submodules lyrebird and snowflake.
printf '\n\n--- Fetching Lyrebird and Snowflake dependencies...\n'
if test -e ".git"; then
    # There's a .git directory - we must be in the development pod.
    git submodule update --init --recursive
    cd lyrebird || exit 1
    git reset --hard
    cp -a . "$TEMPDIR/lyrebird"
    cd ../snowflake || exit 1
    git reset --hard
    cp -a . "$TEMPDIR/snowflake"
    cd ..
else
    # No .git directory - That's a normal install.
    git clone https://gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird.git "$TEMPDIR/lyrebird"
    cd "$TEMPDIR/lyrebird" || exit 1
    git checkout --force --quiet aab4891
    git clone https://gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake.git "$TEMPDIR/snowflake"
    cd "$TEMPDIR/snowflake" || exit 1
    git checkout --force --quiet b130151
    cd "$CURRENT" || exit 1
fi

# Apply patches.
printf '\n\n--- Apply patches to Lyrebird and Snowflake...\n'
patch --directory="$TEMPDIR/lyrebird" --strip=1 < lyrebird.patch
patch --directory="$TEMPDIR/snowflake" --strip=1 < snowflake.patch

# Compile framework.
printf '\n\n--- Compile %s...\n' "$OUTPUT"
export PATH=~/go/bin:$PATH
cd "$TEMPDIR/IPtProxy.go" || exit 1

gomobile init

MACOSX_DEPLOYMENT_TARGET=11.0 gomobile bind -target=$TARGET -ldflags="-s -w" -o "$CURRENT/$OUTPUT" -iosversion=12.0 -androidapi=21 -v -tags=netcgo -trimpath

### Note:
# $ go tool link -h
#  -s	disable symbol table
#  -w	disable DWARF generation
#
# -> Saves > 50% of file size on all targets!
# See https://github.com/guardianproject/orbot/pull/1061

rm -rf "$TEMPDIR"

printf '\n\n--- Done.\n\n'
