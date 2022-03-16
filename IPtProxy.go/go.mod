module github.com/tladesignz/IPtProxy.git

go 1.16

replace (
	git.torproject.org/pluggable-transports/snowflake.git/v2 => ../snowflake
	github.com/pion/dtls/v2 => github.com/pion/dtls/v2 v2.0.12
	gitlab.com/yawning/obfs4.git => ../obfs4
	www.bamsoftware.com/git/dnstt.git => ../dnstt
)

require (
	git.torproject.org/pluggable-transports/snowflake.git/v2 v2.1.0
	gitlab.com/yawning/obfs4.git v0.0.0-20220204003609-77af0cba934d
	golang.org/x/mobile v0.0.0-20220307220422-55113b94f09c // indirect
	golang.org/x/tools v0.1.8-0.20211022200916-316ba0b74098 // indirect
	www.bamsoftware.com/git/dnstt.git v0.0.0-00010101000000-000000000000
)
