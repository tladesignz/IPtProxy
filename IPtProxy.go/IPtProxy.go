package IPtProxy

import (
	"fmt"
	snowflakeclient "git.torproject.org/pluggable-transports/snowflake.git/client"
	"github.com/Yawning/obfs4.git/obfs4proxy"
	"os"
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

// Start the Obfs4Proxy.
//goland:noinspection GoUnusedExportedFunction
func StartObfs4Proxy() {
	if obfs4ProxyRunning {
		return
	}

	obfs4ProxyRunning = true

	fixEnv()

	go obfs4proxy.Main()
}

// Start the Snowflake client.
//goland:noinspection GoUnusedExportedFunction
func StartSnowflake(ice, url, front, logFile string, logToStateDir, keepLocalAddresses, unsafeLogging bool, maxPeers int) {
	_, _ = fmt.Fprint(os.Stdout, "Try to start snowflake.", snowflakeRunning)

	if snowflakeRunning {
		return
	}

	snowflakeRunning = true

	fixEnv()

	go snowflakeclient.Start(ice, url, front, logFile, logToStateDir, keepLocalAddresses, unsafeLogging, maxPeers)
}

//goland:noinspection GoUnusedExportedFunction
func StopSnowflake() {
	_, _ = fmt.Fprint(os.Stdout, "Try to stop snowflake.", snowflakeRunning)

	if !snowflakeRunning {
		return
	}

	go snowflakeclient.Stop()

	snowflakeRunning = false
}

// Hack: Set some environment variables that are either
// required, or values that we want. have to do this here, since we can only
// launch this in a thread and the manipulation of environment variables
// from within an iOS app won't end up in goptlib properly.
func fixEnv() {
	_ = os.Setenv("TOR_PT_CLIENT_TRANSPORTS", "meek_lite,obfs2,obfs3,obfs4,scramblesuit,snowflake")
	_ = os.Setenv("TOR_PT_MANAGED_TRANSPORT_VER", "1")

	tmpdir := os.Getenv("TMPDIR")
	if tmpdir == "" {
		os.Exit(1)
	}
	_ = os.Setenv("TOR_PT_STATE_LOCATION", tmpdir+"/pt_state")
}
