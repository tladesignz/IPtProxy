module github.com/tladesignz/IPtProxy.git

go 1.16

replace (
	git.torproject.org/pluggable-transports/snowflake.git => ../snowflake
	gitlab.com/yawning/obfs4.git => ../obfs4
)

require (
	git.torproject.org/pluggable-transports/snowflake.git v1.1.0
	gitlab.com/yawning/obfs4.git v0.0.0-20210511220700-e330d1b7024b
	golang.org/x/mobile v0.0.0-20210924032853-1c027f395ef7 // indirect
)
