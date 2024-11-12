package IPtProxy

import (
	"log"

	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/common/event"
	sfp "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/proxy/lib"
)

// SnowflakeProxy - Class to start and stop a Snowflake proxy.
type SnowflakeProxy struct {

	// Capacity - the maximum number of clients a Snowflake will serve. If set to 0, the proxy will accept an unlimited number of clients.
	Capacity int

	// BrokerUrl - Defaults to https://snowflake-broker.torproject.net/, if empty.
	BrokerUrl string

	// RelayUrl - WebSocket relay URL. Defaults to wss://snowflake.bamsoftware.com/, if empty.
	RelayUrl string

	// StunServer - STUN URL. Defaults to stun:stun.l.google.com:19302, if empty.
	StunServer string

	// NatProbeUrl - Defaults to https://snowflake-broker.torproject.net:8443/probe, if empty.
	NatProbeUrl string

	isRunning bool
	proxy     *sfp.SnowflakeProxy

	// OnConnected - Callback for when a client has successfully connected to the proxy
	OnConnected func()
}

type proxyEventCallbacks struct {
	onConnectedCallback func()
}

func (c *proxyEventCallbacks) OnNewSnowflakeEvent(e event.SnowflakeEvent) {
	switch e.(type) {
	case event.EventOnProxyClientConnected:
		c.onConnectedCallback()
	default:
	}
}

// Start - Start the Snowflake proxy.
func (sp *SnowflakeProxy) Start() {
	if sp.isRunning {
		return
	}
	if sp.Capacity < 1 {
		sp.Capacity = 0
	}

	dispatcher := event.NewSnowflakeEventDispatcher()
	callbacks := &proxyEventCallbacks{
		onConnectedCallback: sp.OnConnected,
	}
	dispatcher.AddSnowflakeEventListener(callbacks)

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
		EventDispatcher:        dispatcher,
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

// Stop - Stop the Snowflake proxy.
func (sp *SnowflakeProxy) Stop() {
	if sp.isRunning {
		sp.proxy.Stop()
		sp.isRunning = false
		sp.proxy = nil
	}
}

// IsRunning - Checks to see if a snowflake proxy is running in your app.
func (sp *SnowflakeProxy) IsRunning() bool {
	return sp.isRunning
}
