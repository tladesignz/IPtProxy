module github.com/tladesignz/IPtProxy.git

go 1.15

replace (
	git.torproject.org/pluggable-transports/snowflake.git => ../snowflake
	github.com/Yawning/obfs4.git => ../obfs4
)

require (
	git.torproject.org/pluggable-transports/snowflake.git v0.0.0-00010101000000-000000000000
	github.com/Yawning/obfs4.git v0.0.0-00010101000000-000000000000
	gitlab.com/yawning/obfs4.git v0.0.0-20200410113629-2d8f3c8bbfd7 // indirect
	golang.org/x/mobile v0.0.0-20200801112145-973feb4309de // indirect
)
