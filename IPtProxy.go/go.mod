module github.com/tladesignz/IPtProxy.git

go 1.19

replace (
	gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird => ../lyrebird
	gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2 => ../snowflake
)

require (
	gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird v0.0.0-00010101000000-000000000000
	gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2 v2.6.0
)

require (
	golang.org/x/mobile v0.0.0-20230531173138-3c911d8e3eda // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/sync v0.0.0-20220819030929-7fc1605a5dde // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/tools v0.1.12 // indirect
)
