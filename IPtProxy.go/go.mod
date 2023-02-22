module github.com/tladesignz/IPtProxy.git

go 1.16

replace (
	gitlab.com/yawning/obfs4.git => ../obfs4
	git.torproject.org/pluggable-transports/snowflake.git/v2 => ../snowflake
	github.com/pion/dtls/v2 => github.com/pion/dtls/v2 v2.0.12
)

require (
	gitlab.com/yawning/obfs4.git v0.0.0
	git.torproject.org/pluggable-transports/snowflake.git/v2 v2.5.1
	golang.org/x/mobile v0.0.0-20221110043201-43a038452099
	golang.org/x/tools v0.1.12 // indirect
)
