package IPtProxy

import (
	snowflakeclient "git.torproject.org/pluggable-transports/snowflake.git/client"
	snowflakeproxy "git.torproject.org/pluggable-transports/snowflake.git/proxy"
	"github.com/Yawning/obfs4.git/obfs4proxy"
	"os"
	"runtime"
)

//goland:noinspection GoUnusedConst
const MeekSocksPort = 47352

//goland:noinspection GoUnusedConst
const Obfs2SocksPort = 47353

//goland:noinspection GoUnusedConst
const Obfs3SocksPort = 47354

//goland:noinspection GoUnusedConst
const Obfs4SocksPort = 47351

//goland:noinspection GoUnusedConst
const ScramblesuitSocksPort = 47355

//goland:noinspection GoUnusedConst
const SnowflakeSocksPort = 52610

var obfs4ProxyRunning = false
var snowflakeRunning = false
var snowflakeProxyRunning = false

// Override TOR_PT_STATE_LOCATION, which defaults to "$TMPDIR/pt_state".
var StateLocation string

func init() {
	if runtime.GOOS == "android" {
		StateLocation = "/data/local/tmp"
	} else {
		StateLocation = os.Getenv("TMPDIR")
	}

	StateLocation += "/pt_state"
}

// Start the Obfs4Proxy.
//
// @param logLevel Log level (ERROR/WARN/INFO/DEBUG). Defaults to ERROR if empty string.
//
// @param enableLogging Log to TOR_PT_STATE_LOCATION/obfs4proxy.log.
//
// @param unsafeLogging Disable the address scrubber.
//
//goland:noinspection GoUnusedExportedFunction
func StartObfs4Proxy(logLevel string, enableLogging, unsafeLogging bool) {
	if obfs4ProxyRunning {
		return
	}

	obfs4ProxyRunning = true

	fixEnv()

	go obfs4proxy.Start(&logLevel, &enableLogging, &unsafeLogging)
}

// Stop the Obfs4Proxy.
//goland:noinspection GoUnusedExportedFunction
func StopObfs4Proxy() {
	if !obfs4ProxyRunning {
		return
	}

	go obfs4proxy.Stop()

	obfs4ProxyRunning = false
}

// Start the Snowflake client.
//
// @param ice Comma-separated list of ICE servers.
//
// @param url URL of signaling broker.
//
// @param front Front domain.
//
// @param logFile Name of log file. OPTIONAL
//
// @param logToStateDir Resolve the log file relative to Tor's PT state dir.
//
// @param keepLocalAddresses Keep local LAN address ICE candidates.
//
// @param unsafeLogging Prevent logs from being scrubbed.
//
// @param maxPeers Capacity for number of multiplexed WebRTC peers. DEFAULTs to 1 if less than that.
//
//goland:noinspection GoUnusedExportedFunction
func StartSnowflake(ice, url, front, logFile string, logToStateDir, keepLocalAddresses, unsafeLogging bool, maxPeers int) {
	if snowflakeRunning {
		return
	}

	snowflakeRunning = true

	fixEnv()

	go snowflakeclient.Start(&ice, &url, &front, &logFile, &logToStateDir, &keepLocalAddresses, &unsafeLogging, &maxPeers)
}

// Stop the Snowflake client.
//goland:noinspection GoUnusedExportedFunction
func StopSnowflake() {
	if !snowflakeRunning {
		return
	}

	go snowflakeclient.Stop()

	snowflakeRunning = false
}

// Start the Snowflake proxy.
//
// @param capacity Maximum concurrent clients. OPTIONAL. Defaults to 10, if 0.
//
// @param broker Broker URL. OPTIONAL. Defaults to https://snowflake-broker.bamsoftware.com/, if empty.
//
// @param relay WebSocket relay URL. OPTIONAL. Defaults to wss://snowflake.bamsoftware.com/, if empty.
//
// @param stun STUN URL. OPTIONAL. Defaults to stun:stun.stunprotocol.org:3478, if empty.
//
// @param logFile Name of log file. OPTIONAL
//
// @param keepLocalAddresses Keep local LAN address ICE candidates.
//
// @param unsafeLogging Prevent logs from being scrubbed.
//
//goland:noinspection GoUnusedExportedFunction
func StartSnowflakeProxy(capacity int, broker, relay, stun, logFile string, keepLocalAddresses, unsafeLogging bool) {
	if snowflakeProxyRunning {
		return
	}

	snowflakeProxyRunning = true

	fixEnv()

	go snowflakeproxy.Start(uint(capacity), broker, relay, stun, logFile, unsafeLogging, keepLocalAddresses)
}

// Stop the Snowflake proxy.
//goland:noinspection GoUnusedExportedFunction
func StopSnowflakeProxy() {
	if !snowflakeProxyRunning {
		return
	}

	go snowflakeproxy.Stop()

	snowflakeProxyRunning = false
}

// Hack: Set some environment variables that are either
// required, or values that we want. have to do this here, since we can only
// launch this in a thread and the manipulation of environment variables
// from within an iOS app won't end up in goptlib properly.
//
// Note: This might be called multiple times when using different fuctions here,
// but that doesn't necessarily mean, that the values set are independent each
// time this is called. It's still the ENVIRONMENT, we're changing here, so there might
// be race conditions.
func fixEnv() {
	_ = os.Setenv("TOR_PT_CLIENT_TRANSPORTS", "meek_lite,obfs2,obfs3,obfs4,scramblesuit,snowflake")
	_ = os.Setenv("TOR_PT_MANAGED_TRANSPORT_VER", "1")

	_ = os.Setenv("TOR_PT_STATE_LOCATION", StateLocation)
}
