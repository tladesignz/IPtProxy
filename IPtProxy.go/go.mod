module github.com/tladesignz/IPtProxy.git

go 1.16

replace (
	git.torproject.org/pluggable-transports/snowflake.git => ../snowflake
	gitlab.com/yawning/obfs4.git => ../obfs4
)

require (
	git.torproject.org/pluggable-transports/snowflake.git v0.0.0-00010101000000-000000000000
	gitlab.com/yawning/obfs4.git v0.0.0-00010101000000-000000000000
	golang.org/x/mobile v0.0.0-20210710064935-76c259c465ba // indirect
)
