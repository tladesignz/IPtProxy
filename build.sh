#!/bin/sh

TARGET=ios
OUTPUT=IPtProxy.framework

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
echo "--- Golang 1.15 or up needs to be installed! Try 'brew install go' on MacOS or snap install go on Linux if we fail further down!\n"
echo "--- Installing gomobile...\n"
go get -v golang.org/x/mobile/cmd/gomobile

# Fetch submodules obfs4 and snowflake.
echo "\n\n--- Fetching Obfs4proxy and Snowflake dependencies...\n"
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
    git clone https://github.com/Yawning/obfs4.git
    cd obfs4 || exit 1
    git checkout --force --quiet 2d8f3c8b
    cd ..
    git clone https://git.torproject.org/pluggable-transports/snowflake.git
    cd snowflake || exit 1
    git checkout --force --quiet 2d43dd26
    cd ..
fi

# Apply patches.
echo "\n\n--- Apply patches to Obfs4proxy and Snowflake...\n"
patch --directory=obfs4 --strip=1 < obfs4.patch
patch --directory=snowflake --strip=1 < snowflake.patch

# Compile framework.
echo "\n\n--- Compile $OUTPUT...\n"
export PATH=~/go/bin:$PATH
cd IPtProxy.go || exit 1
gomobile init
gomobile bind -target=$TARGET -o ../$OUTPUT -v

echo "\n\n--- Done."
