module github.com/tladesignz/IPtProxy.git

go 1.16

replace (
	git.torproject.org/pluggable-transports/snowflake.git/v2 => ../snowflake
	github.com/pion/dtls/v2 => github.com/pion/dtls/v2 v2.0.12
	gitlab.com/yawning/obfs4.git => ../obfs4
)

require (
	git.torproject.org/pluggable-transports/snowflake.git/v2 v2.3.0
	gitlab.com/yawning/obfs4.git v0.0.0-20220204003609-77af0cba934d
	golang.org/x/mobile v0.0.0-20220518205345-8578da9835fd // indirect
	golang.org/x/tools v0.1.8-0.20211022200916-316ba0b74098 // indirect
)
