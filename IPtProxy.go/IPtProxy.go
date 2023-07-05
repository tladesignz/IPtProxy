package IPtProxy

import (
	"errors"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/cmd/lyrebird"
	snowflakeclient "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/client"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/common/event"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/common/safelog"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/common/version"
	sfp "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/proxy/lib"
	"io"
	"io/fs"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

var meekPort = 47000

// MeekPort - Port where Lyrebird will provide its Meek service.
// Only use this after calling StartLyrebird! It might have changed after that!
//
//goland:noinspection GoUnusedExportedFunction
func MeekPort() int {
	return meekPort
}

var obfs2Port = 47100

// Obfs2Port - Port where Lyrebird will provide its Obfs2 service.
// Only use this property after calling StartLyrebird! It might have changed after that!
//
//goland:noinspection GoUnusedExportedFunction
func Obfs2Port() int {
	return obfs2Port
}

var obfs3Port = 47200

// Obfs3Port - Port where Lyrebird will provide its Obfs3 service.
// Only use this property after calling StartLyrebird! It might have changed after that!
//
//goland:noinspection GoUnusedExportedFunction
func Obfs3Port() int {
	return obfs3Port
}

var obfs4Port = 47300

// Obfs4Port - Port where Lyrebird will provide its Obfs4 service.
// Only use this property after calling StartLyrebird! It might have changed after that!
//
//goland:noinspection GoUnusedExportedFunction
func Obfs4Port() int {
	return obfs4Port
}

var scramblesuitPort = 47400

// ScramblesuitPort - Port where Lyrebird will provide its Scramblesuit service.
// Only use this property after calling StartLyrebird! It might have changed after that!
//
//goland:noinspection GoUnusedExportedFunction
func ScramblesuitPort() int {
	return scramblesuitPort
}

var snowflakePort = 52000

// SnowflakePort - Port where Snowflake will provide its service.
// Only use this property after calling StartSnowflake! It might have changed after that!
//
//goland:noinspection GoUnusedExportedFunction
func SnowflakePort() int {
	return snowflakePort
}

var lyrebirdRunning = false
var snowflakeRunning = false
var snowflakeProxy *sfp.SnowflakeProxy

// StateLocation - Sets TOR_PT_STATE_LOCATION
var StateLocation string

// LyrebirdVersion - The version of Lyrebird bundled with IPtProxy.
//
//goland:noinspection GoUnusedExportedFunction
func LyrebirdVersion() string {
	return lyrebird.LyrebirdVersion
}

// SnowflakeVersion - The version of Snowflake bundled with IPtProxy.
//
//goland:noinspection GoUnusedExportedFunction
func SnowflakeVersion() string {
	return version.GetVersion()
}

// LyrebirdLogFile - The log file name used by Lyrebird.
//
// The Lyrebird log file can be found at `filepath.Join(StateLocation, LyrebirdLogFile())`.
//
//goland:noinspection GoUnusedExportedFunction
func LyrebirdLogFile() string {
	return lyrebird.LyrebirdLogFile
}

// StartLyrebird - Start Lyrebird.
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
// @param proxy HTTP, SOCKS4 or SOCKS5 proxy to be used behind Lyrebird. E.g. "socks5://127.0.0.1:12345"
//
// @return Port number where Lyrebird will listen on for Obfs4(!), if no error happens during start up.
//
//	If you need the other ports, check MeekPort, Obfs2Port, Obfs3Port and ScramblesuitPort properties!
//
//goland:noinspection GoUnusedExportedFunction
func StartLyrebird(logLevel string, enableLogging, unsafeLogging bool, proxy string) int {
	if lyrebirdRunning {
		return obfs4Port
	}

	lyrebirdRunning = true

	for !IsPortAvailable(meekPort) {
		meekPort++
	}

	if meekPort >= obfs2Port {
		obfs2Port = meekPort + 1
	}

	for !IsPortAvailable(obfs2Port) {
		obfs2Port++
	}

	if obfs2Port >= obfs3Port {
		obfs3Port = obfs2Port + 1
	}

	for !IsPortAvailable(obfs3Port) {
		obfs3Port++
	}

	if obfs3Port >= obfs4Port {
		obfs4Port = obfs3Port + 1
	}

	for !IsPortAvailable(obfs4Port) {
		obfs4Port++
	}

	if obfs4Port >= scramblesuitPort {
		scramblesuitPort = obfs4Port + 1
	}

	for !IsPortAvailable(scramblesuitPort) {
		scramblesuitPort++
	}

	fixEnv()

	if len(proxy) > 0 {
		_ = os.Setenv("TOR_PT_PROXY", proxy)
	} else {
		_ = os.Unsetenv("TOR_PT_PROXY")
	}

	go lyrebird.Start(&meekPort, &obfs2Port, &obfs3Port, &obfs4Port, &scramblesuitPort, &logLevel, &enableLogging, &unsafeLogging)

	return obfs4Port
}

// StopLyrebird - Stop Lyrebird.
//
//goland:noinspection GoUnusedExportedFunction
func StopLyrebird() {
	if !lyrebirdRunning {
		return
	}

	go lyrebird.Stop()

	lyrebirdRunning = false
}

// StartSnowflake - Start the Snowflake client.
//
// @param ice Comma-separated list of ICE servers.
//
// @param url URL of signaling broker.
//
// @param front Front domain.
//
// @param ampCache OPTIONAL. URL of AMP cache to use as a proxy for signaling.
//
//	Only needed when you want to do the rendezvous over AMP instead of a domain fronted server.
//
// @param logFile Name of log file. OPTIONAL. Defaults to no log.
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
func StartSnowflake(ice, url, front, ampCache, logFile string, logToStateDir, keepLocalAddresses, unsafeLogging bool, maxPeers int) int {
	if snowflakeRunning {
		return snowflakePort
	}

	snowflakeRunning = true

	for !IsPortAvailable(snowflakePort) {
		snowflakePort++
	}

	fixEnv()

	go snowflakeclient.Start(&snowflakePort, &ice, &url, &front, &ampCache, &logFile, &logToStateDir, &keepLocalAddresses, &unsafeLogging, &maxPeers)

	return snowflakePort
}

// StopSnowflake - Stop the Snowflake client.
//
//goland:noinspection GoUnusedExportedFunction
func StopSnowflake() {
	if !snowflakeRunning {
		return
	}

	go snowflakeclient.Stop()

	snowflakeRunning = false
}

// SnowflakeClientConnected - Interface to use when clients connect
// to the snowflake proxy. For use with StartSnowflakeProxy
type SnowflakeClientConnected interface {
	// Connected - callback method to handle snowflake proxy client connections.
	Connected()
}

// StartSnowflakeProxy - Start the Snowflake proxy.
//
// @param capacity the maximum number of clients a Snowflake will serve. If set to 0, the proxy will accept an unlimited number of clients.
//
// @param broker Broker URL. OPTIONAL. Defaults to https://snowflake-broker.torproject.net/, if empty.
//
// @param relay WebSocket relay URL. OPTIONAL. Defaults to wss://snowflake.bamsoftware.com/, if empty.
//
// @param stun STUN URL. OPTIONAL. Defaults to stun:stun.l.google.com:19302, if empty.
//
// @param natProbe OPTIONAL. Defaults to https://snowflake-broker.torproject.net:8443/probe, if empty.
//
// @param logFile Name of log file. OPTIONAL. Defaults to STDERR.
//
// @param keepLocalAddresses Keep local LAN address ICE candidates.
//
// @param unsafeLogging Prevent logs from being scrubbed.
//
// @param clientConnected A delegate which is called when a client successfully connected.
//
//	Will be called on its own thread! You will need to switch to your own UI thread,
//	if you want to do UI stuff!! OPTIONAL
//
//goland:noinspection GoUnusedExportedFunction
func StartSnowflakeProxy(capacity int, broker, relay, stun, natProbe, logFile string, keepLocalAddresses, unsafeLogging bool, clientConnected SnowflakeClientConnected) {
	if snowflakeProxy != nil {
		return
	}

	if capacity < 1 {
		capacity = 0
	}

	snowflakeProxy = &sfp.SnowflakeProxy{
		Capacity:               uint(capacity),
		STUNURL:                stun,
		BrokerURL:              broker,
		KeepLocalAddresses:     keepLocalAddresses,
		RelayURL:               relay,
		NATProbeURL:            natProbe,
		ProxyType:              "iptproxy",
		RelayDomainNamePattern: "snowflake.torproject.net$",
		AllowNonTLSRelay:       false,
		EventDispatcher:        event.NewSnowflakeEventDispatcher(),
		ClientConnectedCallback: func() {
			if clientConnected != nil {
				clientConnected.Connected()
			}
		},
	}

	fixEnv()

	go func(snowflakeProxy *sfp.SnowflakeProxy) {
		var logOutput io.Writer = os.Stdout
		log.SetFlags(log.LstdFlags | log.LUTC)

		if logFile != "" {
			f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
			if err != nil {
				log.Fatal(err)
			}
			defer func(f *os.File) {
				_ = f.Close()
			}(f)
			logOutput = io.MultiWriter(os.Stdout, f)
		}
		if unsafeLogging {
			log.SetOutput(logOutput)
		} else {
			log.SetOutput(&safelog.LogScrubber{Output: logOutput})
		}
		err := snowflakeProxy.Start()
		if err != nil {
			log.Fatal(err)
		}
	}(snowflakeProxy)
}

// IsSnowflakeProxyRunning - Checks to see if a snowflake proxy is running in your app.
//
//goland:noinspection GoUnusedExportedFunction
func IsSnowflakeProxyRunning() bool {
	return snowflakeProxy != nil
}

// StopSnowflakeProxy - Stop the Snowflake proxy.
//
//goland:noinspection GoUnusedExportedFunction
func StopSnowflakeProxy() {
	if snowflakeProxy == nil {
		return
	}

	go func(snowflakeProxy *sfp.SnowflakeProxy) {
		snowflakeProxy.Stop()
	}(snowflakeProxy)

	snowflakeProxy = nil
}

// IsPortAvailable - Checks to see if a given port is not in use.
//
// @param port The port to check.
func IsPortAvailable(port int) bool {
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))

	conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)

	if err != nil {
		return true
	}

	_ = conn.Close()

	return false
}

// Hack: Set some environment variables that are either
// required, or values that we want. Have to do this here, since we can only
// launch this in a thread and the manipulation of environment variables
// from within an iOS app won't end up in goptlib properly.
//
// Note: This might be called multiple times when using different functions here,
// but that doesn't necessarily mean, that the values set are independent each
// time this is called. It's still the ENVIRONMENT, we're changing here, so there might
// be race conditions.
func fixEnv() {
	info, err := os.Stat(StateLocation)

	// If dir does not exist, try to create it.
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(StateLocation, 0700)

		if err == nil {
			info, err = os.Stat(StateLocation)
		}
	}

	// If it is not a dir, panic.
	if err == nil && !info.IsDir() {
		err = fs.ErrInvalid
	}

	// Create a file within dir to test writability.
	if err == nil {
		tempFile := StateLocation + "/.iptproxy-writetest"
		var file *os.File
		file, err = os.Create(tempFile)

		// Remove the test file again.
		if err == nil {
			file.Close()

			err = os.Remove(tempFile)
		}
	}

	if err != nil {
		panic("Error with StateLocation directory \"" + StateLocation + "\":\n" +
			"  " + err.Error() + "\n" +
			"  StateLocation needs to be set to a writable directory.\n" +
			"  Use an app-private directory to avoid information leaks.\n" +
			"  Use a non-temporary directory to allow reuse of potentially stored state.")
	}

	_ = os.Setenv("TOR_PT_CLIENT_TRANSPORTS", "meek_lite,obfs2,obfs3,obfs4,scramblesuit,snowflake")
	_ = os.Setenv("TOR_PT_MANAGED_TRANSPORT_VER", "1")
	_ = os.Setenv("TOR_PT_STATE_LOCATION", StateLocation)
}
