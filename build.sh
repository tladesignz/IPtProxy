#!/bin/sh

cd "$(dirname "$0")"

if test -e "IPtProxy.framework"; then
    echo "--- No build necessary, IPtProxy.framework already exists."
    exit
fi

# Install dependencies. Go itself is a prerequisite.
echo "--- Golang 1.15 or up needs to be installed! Try 'brew install go' if we fail further down!\n"
echo "--- Installing gomobile...\n"
go get -v golang.org/x/mobile/cmd/gomobile

# Fetch submodules obfs4 and snowflake.
echo "\n\n--- Fetching Obfs4proxy and Snowflake dependencies...\n"
if test -e ".git"; then
    # There's a .git directory - we must be in the development pod.
    git submodule update --init --recursive
    cd obfs4
    git restore .
    cd ../snowflake
    git restore .
    cd ..
else
    # No .git directory - That's a normal install.
    git clone https://github.com/Yawning/obfs4.git
    cd obfs4
    git checkout --force 2d8f3c8b
    cd ..
    git clone https://git.torproject.org/pluggable-transports/snowflake.git
    cd snowflake
    git checkout --force 2d43dd26
    cd ..
fi

# Apply patches.
echo "\n\n--- Apply patches to Obfs4proxy and Snowflake...\n"
patch --directory=obfs4 --strip=1 < obfs4.patch
patch --directory=snowflake --strip=1 < snowflake.patch

# Compile framework.
echo "\n\n--- Compile IPtProxy.framework...\n"
export PATH=~/go/bin:$PATH
cd IPtProxy.go
gomobile bind -target=ios -o ../IPtProxy.framework -v

echo "\n\n--- Done."
