#!/bin/sh

cd "$(dirname "$0")" || exit 1

cd "IPtProxy.go" || exit 1

gomobile init

MACOSX_DEPLOYMENT_TARGET=11.0 gomobile bind -target=ios,iossimulator,macos -o "../IPtProxy.xcframework" -iosversion=12.0 -v -tags=netcgo -trimpath
