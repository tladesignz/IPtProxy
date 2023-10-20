cd "$(dirname "$0")" || exit 1

cd lyrebird || exit 1
git reset --hard
cd ../snowflake || exit 1
git reset --hard
cd ..

patch --directory="lyrebird" --strip=1 < lyrebird.patch
patch --directory="snowflake" --strip=1 < snowflake.patch

cd "IPtProxy.go" || exit 1

gomobile init

MACOSX_DEPLOYMENT_TARGET=11.0 gomobile bind -target=ios,iossimulator,macos -o "../IPtProxy.xcframework" -iosversion=12.0 -v -tags=netcgo -trimpath
