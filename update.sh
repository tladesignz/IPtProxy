#!/bin/sh

cd "$(dirname "$0")" || exit 1

cd dnstt || exit
git pull

cd ../IPtProxy.go || exit
go get gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib@latest
go get gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird@latest
go get gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2@latest
go get golang.org/x/net@latest
go get -tool golang.org/x/mobile/cmd/gomobile@latest
go mod tidy
