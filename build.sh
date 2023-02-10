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
printf '\n--- Golang 1.16 or up needs to be installed! Try "brew install go" on MacOS or "snap install go" on Linux if we fail further down!'
printf '\n--- Installing gomobile...\n'
go install golang.org/x/mobile/cmd/gomobile@latest

# Prepare build environment
printf '\n\n--- Prepare build environment at %s...\n' "$TEMPDIR"
CURRENT=$PWD
rm -rf "$TEMPDIR"
mkdir -p "$TEMPDIR"
cp -a IPtProxy.go "$TEMPDIR/"


# Fetch submodules obfs4 and snowflake.
printf '\n\n--- Fetching Obfs4proxy and Snowflake dependencies...\n'
if test -e ".git"; then
    # There's a .git directory - we must be in the development pod.
    git submodule update --init --recursive
    cd obfs4 || exit 1
    git reset --hard
    cp -a . "$TEMPDIR/obfs4"
    cd ../snowflake || exit 1
    git reset --hard
    cp -a . "$TEMPDIR/snowflake"
    cd ..
else
    # No .git directory - That's a normal install.
    git clone https://gitlab.com/yawning/obfs4.git "$TEMPDIR/obfs4"
    cd "$TEMPDIR/obfs4" || exit 1
    git checkout --force --quiet 336a71d
    git clone https://git.torproject.org/pluggable-transports/snowflake.git "$TEMPDIR/snowflake"
    cd "$TEMPDIR/snowflake" || exit 1
    git checkout --force --quiet 7b77001
    cd "$CURRENT" || exit 1
fi

# Apply patches.
printf '\n\n--- Apply patches to Obfs4proxy and Snowflake...\n'
patch --directory="$TEMPDIR/obfs4" --strip=1 < obfs4.patch
patch --directory="$TEMPDIR/snowflake" --strip=1 < snowflake.patch

# Compile framework.
printf '\n\n--- Compile %s...\n' "$OUTPUT"
export PATH=~/go/bin:$PATH
cd "$TEMPDIR/IPtProxy.go" || exit 1

gomobile init

MACOSX_DEPLOYMENT_TARGET=11.0 gomobile bind -target=$TARGET -o "$CURRENT/$OUTPUT" -iosversion=11.0 -androidapi=19 -v -tags=netcgo -trimpath

rm -rf "$TEMPDIR"

printf '\n\n--- Done.\n\n'
