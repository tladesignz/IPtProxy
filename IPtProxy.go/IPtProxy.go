package IPtProxy

import (
	snowflakeclient "git.torproject.org/pluggable-transports/snowflake.git/client"
	snowflakeproxy "git.torproject.org/pluggable-transports/snowflake.git/proxy"
	"gitlab.com/yawning/obfs4.git/obfs4proxy"
	"net"
	"os"
	"runtime"
	"strconv"
	"time"
)

var meekPort = 47000

// Port where Obfs4proxy will provide its Meek service.
// Only use this after calling StartObfs4Proxy! It might have changed after that!
func MeekPort() int {
	return meekPort
}

var obfs2Port = 47100

// Port where Obfs4proxy will provide its Obfs2 service.
// Only use this property after calling StartObfs4Proxy! It might have changed after that!
func Obfs2Port() int {
	return obfs2Port
}

var obfs3Port = 47200

// Port where Obfs4proxy will provide its Obfs3 service.
// Only use this property after calling StartObfs4Proxy! It might have changed after that!
func Obfs3Port() int {
	return obfs3Port
}

var obfs4Port = 47300

// Port where Obfs4proxy will provide its Obfs4 service.
// Only use this property after calling StartObfs4Proxy! It might have changed after that!
func Obfs4Port() int {
	return obfs4Port
}

var scramblesuitPort = 47400

// Port where Obfs4proxy will provide its Scramblesuit service.
// Only use this property after calling StartObfs4Proxy! It might have changed after that!
func ScramblesuitPort() int {
	return scramblesuitPort
}

var snowflakePort = 52000

// Port where Snowflike will provide its service.
// Only use this property after calling StartSnowflake! It might have changed after that!
func SnowflakePort() int {
	return snowflakePort
}

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
// This will test, if the default ports are available. If not, it will increment them until there is.
// Only use the port properties after calling this, they might have been changed!
//
// @param logLevel Log level (ERROR/WARN/INFO/DEBUG). Defaults to ERROR if empty string.
//
// @param enableLogging Log to TOR_PT_STATE_LOCATION/obfs4proxy.log.
//
// @param unsafeLogging Disable the address scrubber.
//
// @return Port number where Obfs4Proxy will listen on for Obfs4(!), if no error happens during start up.
//	If you need the other ports, check MeekPort, Obfs2Port, Obfs3Port and ScramblesuitPort properties!
//
//
//goland:noinspection GoUnusedExportedFunction
func StartObfs4Proxy(logLevel string, enableLogging, unsafeLogging bool) int {
	if obfs4ProxyRunning {
		return obfs4Port
	}

	obfs4ProxyRunning = true

	for !isAvailable(meekPort) {
		meekPort++
	}

	if meekPort >= obfs2Port {
		obfs2Port = meekPort + 1
	}

	for !isAvailable(obfs2Port) {
		obfs2Port++
	}

	if obfs2Port >= obfs3Port {
		obfs3Port = obfs2Port + 1
	}

	for !isAvailable(obfs3Port) {
		obfs3Port++
	}

	if obfs3Port >= obfs4Port {
		obfs4Port = obfs3Port + 1
	}

	for !isAvailable(obfs4Port) {
		obfs4Port++
	}

	if obfs4Port >= scramblesuitPort {
		scramblesuitPort = obfs4Port + 1
	}

	for !isAvailable(scramblesuitPort) {
		scramblesuitPort++
	}

	fixEnv()

	go obfs4proxy.Start(&meekPort, &obfs2Port, &obfs3Port, &obfs4Port, &scramblesuitPort, &logLevel, &enableLogging, &unsafeLogging)

	return obfs4Port
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
// @return Port number where Snowflake will listen on, if no error happens during start up.
//
//goland:noinspection GoUnusedExportedFunction
func StartSnowflake(ice, url, front, logFile string, logToStateDir, keepLocalAddresses, unsafeLogging bool, maxPeers int) int {
	if snowflakeRunning {
		return snowflakePort
	}

	snowflakeRunning = true

	for !isAvailable(snowflakePort) {
		snowflakePort++
	}

	fixEnv()

	go snowflakeclient.Start(&snowflakePort, &ice, &url, &front, &logFile, &logToStateDir, &keepLocalAddresses, &unsafeLogging, &maxPeers)

	return snowflakePort
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
// Note: This might be called multiple times when using different functions here,
// but that doesn't necessarily mean, that the values set are independent each
// time this is called. It's still the ENVIRONMENT, we're changing here, so there might
// be race conditions.
func fixEnv() {
	_ = os.Setenv("TOR_PT_CLIENT_TRANSPORTS", "meek_lite,obfs2,obfs3,obfs4,scramblesuit,snowflake")
	_ = os.Setenv("TOR_PT_MANAGED_TRANSPORT_VER", "1")

	_ = os.Setenv("TOR_PT_STATE_LOCATION", StateLocation)
}

func isAvailable(port int) bool {
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))

	conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)

	if err != nil {
		return true
	}

	err = conn.Close()

	return false
}
