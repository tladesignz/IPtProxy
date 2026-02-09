package IPtProxy

import (
	"log"

	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/common/event"
	sfp "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/proxy/lib"
	"time"
)

// SnowflakeClientEvents - Interface to get information about clients connecting, disconnecting, failing to connect and
// for statistics of the snowflake proxy. For use with StartSnowflakeProxy
type SnowflakeClientEvents interface {
	// Connected - callback method to handle snowflake proxy client connections.
	Connected()

	// Disconnected - The connection with the client has now been closed,
	// after getting successfully established.
	Disconnected(country string)

	// ConnectionFailed - Rendezvous with a client succeeded, but a data channel has not been created.
	ConnectionFailed()

	// Stats - callback method to handle snowflake proxy client statistics.
	//
	// BEWARE! This is called very often. Before doing anything, make sure there are any non-zero values.
	//
	// @param connectionCount Completed successful connections.
	//
	// @param failedConnectionCount Connections that failed to establish.
	//
	// @param inboundBytes number of inbound `inboundUnit` bytes.
	//
	// @param outboundBytes number of outbound `outboundUnit` bytes.
	//
	// @param inboundUnit unit of inbound bytes. (e.g. "KB")
	//
	// @param outboundUnit unit of outbound bytes. (e.g. "KB")
	//
	// @param summaryInterval time in nanoseconds between summary statistics.
	Stats(
		connectionCount int,
		failedConnectionCount int64,
		inboundBytes, outboundBytes int64,
		inboundUnit, outboundUnit string,
		summaryInterval int64)
}

// SnowflakeProxy - Class to start and stop a Snowflake proxy.
type SnowflakeProxy struct {

	// Capacity - the maximum number of clients a Snowflake will serve. If set to 0, the proxy will accept an unlimited number of clients.
	Capacity int

	// BrokerUrl - Defaults to https://snowflake-broker.torproject.net/, if empty.
	BrokerUrl string

	// RelayUrl - WebSocket relay URL. Defaults to wss://snowflake.bamsoftware.com/, if empty.
	RelayUrl string

	// EphemeralMinPort - limit the range of ports that
	// ICE UDP connections may allocate from.
	// When specifying the range, make sure it's at least 2x as wide
	// as the number of clients that you are hoping to serve concurrently
	// (see the `Capacity` property).
	// If EphemeralMinPort or EphemeralMaxPort is left 0, no limit will be applied.
	EphemeralMinPort int

	// EphemeralMaxPort - limit the range of ports that
	// ICE UDP connections may allocate from.
	// When specifying the range, make sure it's at least 2x as wide
	// as the number of clients that you are hoping to serve concurrently
	// (see the `Capacity` property).
	// If EphemeralMinPort or EphemeralMaxPort is left 0, no limit will be applied.
	EphemeralMaxPort int

	// StunServer - STUN URL. Defaults to stun:stun.l.google.com:19302, if empty.
	StunServer string

	// NatProbeUrl - Defaults to https://snowflake-broker.torproject.net:8443/probe, if empty.
	NatProbeUrl string

	// NATTypeMeasurementInterval is time before NAT type is retested. Defaults to 0, if empty.
	NATTypeMeasurementInterval int64

	// PollInterval - In seconds. How often to ask the broker for a new client. Defaults to 5 seconds if <= 0.
	PollInterval int

	// ClientEvents - A delegate which is called when a client successfully connected, disconnected, failed, or to
	// receive statistics.
	// Will be called on its own thread! You will need to switch to your own UI thread
	// if you want to do UI stuff!
	ClientEvents SnowflakeClientEvents

	// ProxyTypeIdentifier - Identifier for the proxy type. Used for logging and identification purposes.
	// Defaults to "iptproxy", if empty.
	// ATTENTION: This will affect Tor Project statistics. Only change if you talked to Tor Project about it.
	ProxyTypeIdentifier string

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

	if sp.ProxyTypeIdentifier == "" {
		sp.ProxyTypeIdentifier = "iptproxy"
	}

	eventDispatcher := event.NewSnowflakeEventDispatcher()
	eventDispatcher.AddSnowflakeEventListener(sp)

	sp.proxy = &sfp.SnowflakeProxy{
		PollInterval:               time.Duration(sp.PollInterval) * time.Second,
		Capacity:                   uint(sp.Capacity),
		STUNURL:                    sp.StunServer,
		BrokerURL:                  sp.BrokerUrl,
		KeepLocalAddresses:         false,
		RelayURL:                   sp.RelayUrl,
		EphemeralMinPort:           uint16(sp.EphemeralMinPort),
		EphemeralMaxPort:           uint16(sp.EphemeralMaxPort),
		NATProbeURL:                sp.NatProbeUrl,
		NATTypeMeasurementInterval: time.Duration(sp.NATTypeMeasurementInterval),
		ProxyType:                  sp.ProxyTypeIdentifier,
		RelayDomainNamePattern:     "snowflake.torproject.net$",
		AllowNonTLSRelay:           false,
		EventDispatcher:            eventDispatcher,
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
	switch ev := e.(type) {
	case event.EventOnProxyClientConnected:
		if sp.ClientEvents != nil {
			sp.ClientEvents.Connected()
		}

	case event.EventOnProxyConnectionOver:
		if sp.ClientEvents != nil {
			sp.ClientEvents.Disconnected(ev.Country)
		}

	case event.EventOnProxyConnectionFailed:
		if sp.ClientEvents != nil {
			sp.ClientEvents.ConnectionFailed()
		}

	case event.EventOnProxyStats:
		if sp.ClientEvents != nil {
			sp.ClientEvents.Stats(ev.ConnectionCount, int64(ev.FailedConnectionCount), ev.InboundBytes, ev.OutboundBytes,
				ev.InboundUnit, ev.OutboundUnit, ev.SummaryInterval.Nanoseconds())
		}

	default:
	}
}
