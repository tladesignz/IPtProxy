package IPtProxy

import (
	"log"

	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/common/event"
	sfp "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/proxy/lib"
)

type SnowflakeProxy struct {
	Capacity    int
	BrokerUrl   string
	RelayUrl    string
	StunServer  string
	NatProbeUrl string

	isRunning bool
	proxy     *sfp.SnowflakeProxy
}

func (sp *SnowflakeProxy) Start() {
	if sp.isRunning {
		return
	}
	if sp.Capacity < 1 {
		sp.Capacity = 0
	}

	sp.proxy = &sfp.SnowflakeProxy{
		Capacity:               uint(sp.Capacity),
		STUNURL:                sp.StunServer,
		BrokerURL:              sp.BrokerUrl,
		KeepLocalAddresses:     false,
		RelayURL:               sp.RelayUrl,
		NATProbeURL:            sp.NatProbeUrl,
		ProxyType:              "iptproxy",
		RelayDomainNamePattern: "snowflake.torproject.net$",
		AllowNonTLSRelay:       false,
		EventDispatcher:        event.NewSnowflakeEventDispatcher(),
	}

	go func() {
		sp.isRunning = true
		err := sp.proxy.Start()
		if err != nil {
			sp.isRunning = false
			log.Fatal(err)
		}
	}()
}

func (sp *SnowflakeProxy) Stop() {
	if sp.isRunning {
		sp.proxy.Stop()
		sp.isRunning = false
		sp.proxy = nil
	}
}

func (sp *SnowflakeProxy) IsRunning() bool {
	return sp.isRunning
}
