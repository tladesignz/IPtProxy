package IPtProxy

import (
	snowflakeclient "git.torproject.org/pluggable-transports/snowflake.git/client"
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

var snowflakeRunning = false
var obfs4ProxyRunning = false

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
// - parameter logLevel: Log level (ERROR/WARN/INFO/DEBUG). Defaults to ERROR if empty string.
//
// - parameter enableLogging: Log to TOR_PT_STATE_LOCATION/obfs4proxy.log.
//
// - parameter unsafeLogging: Disable the address scrubber.
//
//goland:noinspection GoUnusedExportedFunction
func StartObfs4Proxy(logLevel string, enableLogging, unsafeLogging bool) {
	if obfs4ProxyRunning {
		return
	}

	obfs4ProxyRunning = true

	fixEnv()

	go obfs4proxy.Start(logLevel, &enableLogging, &unsafeLogging)
}

// Start the Snowflake client.
//
// - parameter ice: Comma-separated list of ICE servers.
//
// - parameter url: URL of signaling broker.
//
// - parameter front: Front domain.
//
// - parameter logFilename: Name of log file. OPTIONAL
//
// - parameter logToStateDir: Resolve the log file relative to Tor's PT state dir.
//
// - parameter keepLocalAddresses: Keep local LAN address ICE candidates.
//
// - parameter unsafeLogging: Prevent logs from being scrubbed.
//
// - parameter max: Capacity for number of multiplexed WebRTC peers. DEFAULTs to 1 if less than that.
//
//goland:noinspection GoUnusedExportedFunction
func StartSnowflake(ice, url, front, logFile string, logToStateDir, keepLocalAddresses, unsafeLogging bool, maxPeers int) {
	if snowflakeRunning {
		return
	}

	snowflakeRunning = true

	fixEnv()

	go snowflakeclient.Start(ice, url, front, logFile, &logToStateDir, &keepLocalAddresses, &unsafeLogging, maxPeers)
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
