package IPtProxy

import (
	"log"

	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/common/event"
	sfp "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/proxy/lib"
	"time"
)

// SnowflakeClientConnected - Interface to use when clients connect
// to the snowflake proxy. For use with StartSnowflakeProxy
type SnowflakeClientConnected interface {
	// Connected - callback method to handle snowflake proxy client connections.
	Connected()
}

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

	// PollInterval - In seconds. How often to ask the broker for a new client. Defaults to 5 seconds, if <= 0.
	PollInterval int

	// ClientConnected - A delegate which is called when a client successfully connected.
	// Will be called on its own thread! You will need to switch to your own UI thread,
	// if you want to do UI stuff!
	ClientConnected SnowflakeClientConnected

	isRunning bool
	proxy     *sfp.SnowflakeProxy
}

// Start - Start the Snowflake proxy.
func (sp *SnowflakeProxy) Start() {
	if sp.isRunning {
		return
	}
	if sp.Capacity < 1 {
		sp.Capacity = 0
	}

	if sp.PollInterval < 0 {
		sp.PollInterval = 0
	}

	eventDispatcher := event.NewSnowflakeEventDispatcher()
	eventDispatcher.AddSnowflakeEventListener(sp)

	sp.proxy = &sfp.SnowflakeProxy{
		PollInterval:           time.Duration(sp.PollInterval) * time.Second,
		Capacity:               uint(sp.Capacity),
		STUNURL:                sp.StunServer,
		BrokerURL:              sp.BrokerUrl,
		KeepLocalAddresses:     false,
		RelayURL:               sp.RelayUrl,
		NATProbeURL:            sp.NatProbeUrl,
		ProxyType:              "iptproxy",
		RelayDomainNamePattern: "snowflake.torproject.net$",
		AllowNonTLSRelay:       false,
		EventDispatcher:        eventDispatcher,
	}

	go func() {
		sp.isRunning = true
		err := sp.proxy.Start()
		if err != nil {
			sp.isRunning = false
			eventDispatcher.RemoveSnowflakeEventListener(sp)
			log.Print(err)
		}
	}()
}

// Stop - Stop the Snowflake proxy.
func (sp *SnowflakeProxy) Stop() {
	if sp.isRunning {
		sp.proxy.Stop()
		sp.proxy.EventDispatcher.RemoveSnowflakeEventListener(sp)
		sp.isRunning = false
		sp.proxy = nil
	}
}

// IsRunning - Checks to see if a snowflake proxy is running in your app.
func (sp *SnowflakeProxy) IsRunning() bool {
	return sp.isRunning
}

func (sp *SnowflakeProxy) OnNewSnowflakeEvent(e event.SnowflakeEvent) {
	switch e.(type) {
	case event.EventOnProxyClientConnected:
		if sp.ClientConnected != nil {
			sp.ClientConnected.Connected()
		}

	default:
	}
}
