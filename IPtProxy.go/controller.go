package IPtProxy

/**
Package IPtProxy combines tor pluggable transport clients into a single library for use
with mobile applications. Transports, when started, will listen for incoming SOCKS
connections on an available local address and proxy traffic between those connections
and a configured bridge.

Sample gomobile usage:

```java
	import IPtProxy.Controller;

	// Create a new Controller instance with provided state directory
	Controller ptController = Controller.newController("/path/to/statedir", true, false, "DEBUG");

	// Start listening for obfs4 and meek connections, using an outgoing proxy
	ptController.start(IPtProxy.Obfs4, "socks5://localhost:8001");
	ptController.start(IPtProxy.MeekLite, "socks5://localhost:8001");

	// Get the address that is listening for SOCKS connections for each transport
	String obfs4Addr = ptController.localAddress(IPtProxy.Obfs4);
	String meekAddr = ptController.localAddress(IPtProxy.MeekLite);

	// Start listening for snowflake connections
	// Note that snowflake setup can happen either here or with SOCKS arguments on
	// a per-connection basis.
	ptController.setSnowflakeIceServers("stun:stun.l.google.com:19302,stun:stun.l.google.com:5349");
	ptController.start(IPtProxy.Snowflake, "");

	// Stop transports
	ptController.stop(IPtProxy.Snowflake);
	ptController.stop(IPtProxy.Obfs4);
	ptController.stop(IPtProxy.MeekLite);
```

Sample pure go usage:

```go
	import github.com/tladesignz/IPtProxy

	func main() {
		ptController := NewController("/path/to/statedir", true, false, "DEBUG")

		ptController.Start(Snowflake)
		ptController.Start(MeekLite)
		ptController.Start(Obfs4)
		addr := ptController.LocalAddress(Snowflake)
		fmt.Printf("Listening for snowflake connections on: %s", addr)

		// ...

		ptController.Stop(Snowflake)
		ptController.Stop(Obfs4)
		ptController.Stop(MeekLite)
	}
```
*/

import (
	"errors"
	"io"
	"io/fs"
	"log"
	"net"
	"net/url"
	"os"
	"path"

	pt "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib"
	ptlog "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/common/log"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/transports"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/transports/base"
	sfversion "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/common/version"
	"golang.org/x/net/proxy"
)

// LogFileName - the filename of the log residing in `StateDir`.
const LogFileName = "ipt.log"

// ScrambleSuit - DEPRECATED transport implemented in Lyrebird.
//
//goland:noinspection GoUnusedConst
const ScrambleSuit = "scramblesuit"

// Obfs2 - DEPRECATED transport implemented in Lyrebird.
//
//goland:noinspection GoUnusedConst
const Obfs2 = "obfs2"

// Obfs3 - DEPRECATED transport implemented in Lyrebird.
//
//goland:noinspection GoUnusedConst
const Obfs3 = "obfs3"

// Obfs4 - Transport implemented in Lyrebird.
//
//goland:noinspection GoUnusedConst
const Obfs4 = "obfs4"

// MeekLite - Transport implemented in Lyrebird.
//
//goland:noinspection GoUnusedConst
const MeekLite = "meek_lite"

// Webtunnel - Transport implemented in Lyrebird.
//
//goland:noinspection GoUnusedConst
const Webtunnel = "webtunnel"

// Snowflake - Transport implemented in Snowflake.
//
//goland:noinspection GoUnusedConst
const Snowflake = "snowflake"

// Controller - Class to start and stop transports.
type Controller struct {

	// SnowflakeIceServers is a comma-separated list of ICE server addresses.
	SnowflakeIceServers string

	// SnowflakeBrokerUrl - URL of signaling broker.
	SnowflakeBrokerUrl string

	// SnowflakeFrontDomains is a comma-separated list of domains for either
	// the domain fronting or AMP cache rendezvous methods.
	SnowflakeFrontDomains string

	// SnowflakeAmpCacheUrl - URL of AMP cache to use as a proxy for signaling.
	// Only needed when you want to do the rendezvous over AMP instead of a domain fronted server.
	SnowflakeAmpCacheUrl string

	// SnowflakeSqsUrl - URL of SQS Queue to use as a proxy for signaling.
	SnowflakeSqsUrl string

	// SnowflakeSqsCreds - Credentials to access SQS Queue.
	SnowflakeSqsCreds string

	// SnowflakeMaxPeers - Capacity for number of multiplexed WebRTC peers. DEFAULTs to 1 if less than that.
	SnowflakeMaxPeers string

	stateDir  string
	listeners map[string]*pt.SocksListener
	shutdown  map[string]chan struct{}
}

// NewController - Create a new Controller object.
//
// @param enableLogging Log to StateDir/ipt.log.
//
// @param unsafeLogging Disable the address scrubber.
//
// @param logLevel Log level (ERROR/WARN/INFO/DEBUG). Defaults to ERROR if empty string.
//
//goland:noinspection GoUnusedExportedFunction
func NewController(stateDir string, enableLogging, unsafeLogging bool, logLevel string) *Controller {
	c := &Controller{
		stateDir: stateDir,
	}

	if logLevel == "" {
		logLevel = "ERROR"
	}

	if err := createStateDir(c.stateDir); err != nil {
		log.Printf("Failed to set up state directory: %s", err)
		return nil
	}
	if err := ptlog.Init(enableLogging,
		path.Join(c.stateDir, LogFileName), unsafeLogging); err != nil {
		log.Printf("Failed to set initialize log: %s", err.Error())
		return nil
	}
	if err := ptlog.SetLogLevel(logLevel); err != nil {
		log.Printf("Failed to set log level: %s", err.Error())
		ptlog.Warnf("Failed to set log level: %s", err.Error())
	}

	if err := transports.Init(); err != nil {
		ptlog.Warnf("Failed to initialize transports: %s", err.Error())
		return nil
	}

	c.listeners = make(map[string]*pt.SocksListener)
	c.shutdown = make(map[string]chan struct{})

	return c
}

// StateDir - The StateDir set in the constructor.
//
// @returns the directory you set in the constructor, where transports store their state and where the log file resides.
func (c *Controller) StateDir() string {
	return c.stateDir
}

// addExtraArgs adds the args in extraArgs to the connection args
func addExtraArgs(args *pt.Args, extraArgs *pt.Args) {
	for name, _ := range *extraArgs {
		//only overwrite if connection arg doesn't exist
		if arg, ok := args.Get(name); !ok {
			args.Add(name, arg)
		}
	}
}

func acceptLoop(f base.ClientFactory, ln *pt.SocksListener, proxyURL *url.URL,
	extraArgs *pt.Args, shutdown chan struct{}) error {
	defer ln.Close()
	for {
		conn, err := ln.AcceptSocks()
		if err != nil {
			if e, ok := err.(net.Error); ok && !e.Temporary() {
				return err
			}
			continue
		}
		go clientHandler(f, conn, proxyURL, extraArgs, shutdown)
	}
}

func clientHandler(f base.ClientFactory, conn *pt.SocksConn, proxyURL *url.URL,
	extraArgs *pt.Args, shutdown chan struct{}) {
	defer conn.Close()
	addExtraArgs(&conn.Req.Args, extraArgs)
	args, err := f.ParseArgs(&conn.Req.Args)
	if err != nil {
		ptlog.Errorf("Error parsing PT args: %s", err.Error())
		_ = conn.Reject()
		return
	}
	dialFn := proxy.Direct.Dial
	if proxyURL != nil {
		dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			ptlog.Errorf("Error getting proxy dialer: %s", err.Error())
			_ = conn.Reject()
		}
		dialFn = dialer.Dial
	}
	remote, err := f.Dial("tcp", conn.Req.Target, dialFn, args)
	if err != nil {
		ptlog.Errorf("Error dialing PT: %s", err.Error())
		return
	}
	err = conn.Grant(&net.TCPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		ptlog.Errorf("conn.Grant error: %s", err)
		return
	}
	defer remote.Close()
	done := make(chan struct{}, 2)
	go copyLoop(conn, remote, done)
	// wait for copy loop to finish or for shutdown signal
	select {
	case <-shutdown:
	case <-done:
		ptlog.Noticef("copy loop ended")
	}
}

// Exchanges bytes between two ReadWriters.
// (In this case, between a SOCKS connection and a pt conn)
func copyLoop(socks, sfconn io.ReadWriter, done chan struct{}) {
	go func() {
		if _, err := io.Copy(socks, sfconn); err != nil {
			ptlog.Errorf("copying transport to SOCKS resulted in error: %v", err)
		}
		done <- struct{}{}
	}()
	go func() {
		if _, err := io.Copy(sfconn, socks); err != nil {
			ptlog.Errorf("copying SOCKS to transport resulted in error: %v", err)
		}
		done <- struct{}{}
	}()
}

// LocalAddress - Address of the given transport.
//
// @param methodName one of the constants `ScrambleSuit` (deprecated), `Obfs2` (deprecated), `Obfs3` (deprecated),
// `Obfs4`, `MeekLite`, `Webtunnel` or `Snowflake`.
//
// @return address string containing host and port where the given transport listens.
func (c *Controller) LocalAddress(methodName string) string {
	if ln, ok := c.listeners[methodName]; ok {
		return ln.Addr().String()
	}
	return ""
}

// Port - Port of the given transport.
//
// @param methodName one of the constants `ScrambleSuit` (deprecated), `Obfs2` (deprecated), `Obfs3` (deprecated),
// `Obfs4`, `MeekLite`, `Webtunnel` or `Snowflake`.
//
// @return port number on localhost where the given transport listens.
func (c *Controller) Port(methodName string) int {
	if ln, ok := c.listeners[methodName]; ok {
		return int(ln.Addr().(*net.TCPAddr).AddrPort().Port())
	}
	return 0
}

func createStateDir(path string) error {
	info, err := os.Stat(path)

	// If dir does not exist, try to create it.
	if errors.Is(err, os.ErrNotExist) {
		err = os.MkdirAll(path, 0700)

		if err == nil {
			info, err = os.Stat(path)
		}
	}

	// If it is not a dir, return error
	if err == nil && !info.IsDir() {
		err = fs.ErrInvalid
		return err
	}

	// Create a file within dir to test writability.
	tempFile := path + "/.iptproxy-writetest"
	var file *os.File
	file, err = os.Create(tempFile)

	// Remove the test file again.
	if err == nil {
		_ = file.Close()

		err = os.Remove(tempFile)
	}
	return err
}

// Start - Start given transport.
//
// @param methodName one of the constants `ScrambleSuit` (deprecated), `Obfs2` (deprecated), `Obfs3` (deprecated),
// `Obfs4`, `MeekLite`, `Webtunnel` or `Snowflake`.
//
// @param proxy HTTP, SOCKS4 or SOCKS5 proxy to be used behind Lyrebird. E.g. "socks5://127.0.0.1:12345"
func (c *Controller) Start(methodName string, proxy string) {
	var proxyURL *url.URL
	var err error

	ptlog.Noticef("Launched transport: %v", methodName)

	if proxy != "" {
		proxyURL, err = url.Parse(proxy)
		if err != nil {
			ptlog.Warnf("Failed to parse proxy address: %s", err.Error())
		}
	}

	switch methodName {
	case "snowflake":
		extraArgs := &pt.Args{}
		extraArgs.Add("fronts", c.SnowflakeFrontDomains)
		extraArgs.Add("ice", c.SnowflakeIceServers)
		extraArgs.Add("max", c.SnowflakeMaxPeers)
		extraArgs.Add("url", c.SnowflakeBrokerUrl)
		extraArgs.Add("ampcache", c.SnowflakeAmpCacheUrl)
		extraArgs.Add("sqsqueue", c.SnowflakeSqsUrl)
		extraArgs.Add("sqscreds", c.SnowflakeSqsCreds)
		extraArgs.Add("proxy", proxy)

		t := transports.Get(methodName)
		if t == nil {
			ptlog.Errorf("Failed to initialize %s: no such method", methodName)
			return
		}
		f, err := t.ClientFactory(c.stateDir)
		if err != nil {
			ptlog.Errorf("Failed to initialize %s: %s", methodName, err.Error())
			return
		}
		ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
		if err != nil {
			ptlog.Errorf("Failed to initialize %s: %s", methodName, err.Error())
			return
		}

		c.shutdown[methodName] = make(chan struct{})
		c.listeners[methodName] = ln

		go acceptLoop(f, ln, nil, extraArgs, c.shutdown[methodName])

	default:
		// at the moment, everything else is in lyrebird
		t := transports.Get(methodName)
		if t == nil {
			ptlog.Errorf("Failed to initialize %s: no such method", methodName)
			return
		}

		f, err := t.ClientFactory(c.stateDir)
		if err != nil {
			ptlog.Errorf("Failed to initialize %s: %s", methodName, err.Error())
			return
		}

		ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
		if err != nil {
			ptlog.Errorf("Failed to initialize %s: %s", methodName, err.Error())
			return
		}

		c.listeners[methodName] = ln
		c.shutdown[methodName] = make(chan struct{})

		go acceptLoop(f, ln, proxyURL, nil, c.shutdown[methodName])
	}
}

// Stop - Stop given transport.
//
// @param methodName one of the constants `ScrambleSuit` (deprecated), `Obfs2` (deprecated), `Obfs3` (deprecated),
// `Obfs4`, `MeekLite`, `Webtunnel` or `Snowflake`.
func (c *Controller) Stop(methodName string) {
	if ln, ok := c.listeners[methodName]; ok {
		_ = ln.Close()

		ptlog.Noticef("Shutting down %s", methodName)

		close(c.shutdown[methodName])
		delete(c.shutdown, methodName)
		delete(c.listeners, methodName)
	} else {
		ptlog.Warnf("No listener for %s", methodName)
	}
}

// SnowflakeVersion - The version of Snowflake bundled with IPtProxy.
//
//goland:noinspection GoUnusedExportedFunction
func SnowflakeVersion() string {
	return sfversion.GetVersion()
}

// LyrebirdVersion - The version of Lyrebird bundled with IPtProxy.
//
//goland:noinspection GoUnusedExportedFunction
func LyrebirdVersion() string {
	return "lyrebird-0.2.0"
}
