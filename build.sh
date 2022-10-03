#!/bin/sh

TARGET=ios,iossimulator,macos
OUTPUT=IPtProxy.xcframework

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

# Fetch submodules obfs4 and snowflake.
printf '\n\n--- Fetching Obfs4proxy and Snowflake dependencies...\n'
if test -e ".git"; then
    # There's a .git directory - we must be in the development pod.
    git submodule update --init --recursive
    cd obfs4 || exit 1
    git reset --hard
    cd ../snowflake || exit 1
    git reset --hard
    cd ..
else
    # No .git directory - That's a normal install.
    git clone https://gitlab.com/yawning/obfs4.git
    cd obfs4 || exit 1
    git checkout --force --quiet 336a71d6
    cd ..
    git clone https://git.torproject.org/pluggable-transports/snowflake.git
    cd snowflake || exit 1
    git checkout --force --quiet 36f03dfd
    cd ..
fi

# Apply patches.
printf '\n\n--- Apply patches to Obfs4proxy and Snowflake...\n'
patch --directory=obfs4 --strip=1 < obfs4.patch
patch --directory=snowflake --strip=1 < snowflake.patch

# Compile framework.
printf '\n\n--- Compile %s...\n' "$OUTPUT"
export PATH=~/go/bin:$PATH
cd IPtProxy.go || exit 1

gomobile init

MACOSX_DEPLOYMENT_TARGET=11.0 gomobile bind -target=$TARGET -o ../$OUTPUT -iosversion 11.0 -androidapi 19 -v

printf '\n\n--- Done.\n\n'
