module github.com/tladesignz/IPtProxy.git

go 1.16

replace (
	git.torproject.org/pluggable-transports/snowflake.git/v2 => ../snowflake
	github.com/pion/dtls/v2 => github.com/pion/dtls/v2 v2.0.12
	gitlab.com/yawning/obfs4.git => ../obfs4
)

require (
	git.torproject.org/pluggable-transports/snowflake.git/v2 v2.0.1
	gitlab.com/yawning/obfs4.git v0.0.0-20210511220700-e330d1b7024b
	golang.org/x/mobile v0.0.0-20211207041440-4e6c2922fdee // indirect
)
