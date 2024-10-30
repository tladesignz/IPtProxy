package IPtProxy

/**
Package IPtProxy combines tor pluggable transport clients into a single library for use
with mobile applications. Transports, when started, will listen for incoming SOCKS
connections on an available local address and proxy traffic between those connections
and a configured bridge.

Sample gomobile usage:
 import IPtProxy.Controller;

 // Create a new IPtProxy instance with provided state directory
 Controller iptproxy = Controller.newController("/path/to/statedir", true, false, "DEBUG");

 // Start listening for obfs4 and meek connections, using an outgoing proxy
 iptproxy.start("obfs4", "socks5://localhost:8001");
 iptproxy.start("meek_lite", "socks5://localhost:8001");

 // Get the address that is listening for SOCKS connections for each transport
 obfs4Addr = iptproxy.getLocalAddress("obfs4");
 meekAddr = iptproxy.getLocalAddress("meek_lite");

 // Start listening for snowflake connections
 // Note that snowflake setup can happen either here or with SOCKS arguments on
 // a per-connection basis.
 iptproxy.setSnowflakeIceServers("stun:stun.l.google.com:19302,stun:stun.l.google.com:5349");
 iptproxy.start("snowflake", "");


 // Stop transports
 iptproxy.stop("snowflake");
 iptproxy.stop("obfs4");
 iptproxy.stop("meek_lite");


Sample pure go usage:
 import github.com/tladesignz/IPtProxy

 func main() {
	 iptproxy := NewController("/path/to/statedir", true, false, "DEBUG")

	 iptproxy.Start("snowflake")
	 iptproxy.Start("meek_lite")
	 iptproxy.Start("obfs4")
	 addr := iptproxy.GetLocalAddress("snowflake")
	 fmt.Printf("Listening for snowflake connections on: %s", addr)

	 // ...

	 iptproxy.Stop("snowflake")
	 iptproxy.Stop("obfs4")
	 iptproxy.Stop("meek_lite")
 }

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
	"strings"

	pt "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib"
	ptlog "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/common/log"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/transports"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/transports/base"
	sf "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/client/lib"
	sproxy "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/common/proxy"
	sfversion "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/common/version"
	"golang.org/x/net/proxy"
)

const LogFileName = "ipt.log"

const ScrambleSuit = "scramblesuit"

const Obfs2 = "obfs2"

const Obfs3 = "obfs3"

const Obfs4 = "obfs4"

const MeekLite = "meek_lite"

const Webtunnel = "webtunnel"

const Snowflake = "snowflake"

type Controller struct {

	// SnowflakeIceServers is a comma-separated list of ICE server addresses
	SnowflakeIceServers string
	SnowflakeBrokerUrl  string
	// SnowflakeFrontDomains is a comma-separated list of domains for either
	// the domain fronting or AMP cache rendezvous methods
	SnowflakeFrontDomains string
	SnowflakeAmpCacheUrl  string
	SnowflakeSqsUrl       string
	SnowflakeSqsCreds     string

	stateDir  string
	listeners map[string]*pt.SocksListener
	shutdown  map[string]chan struct{}
}

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

func (c *Controller) StateDir() string {
	return c.stateDir
}

func acceptLoop(f base.ClientFactory, ln *pt.SocksListener, proxyURL *url.URL, shutdown chan struct{}) error {
	defer ln.Close()
	for {
		conn, err := ln.AcceptSocks()
		if err != nil {
			if e, ok := err.(net.Error); ok && !e.Temporary() {
				return err
			}
			continue
		}
		go clientHandler(f, conn, proxyURL, shutdown)
	}
}

func clientHandler(f base.ClientFactory, conn *pt.SocksConn, proxyURL *url.URL, shutdown chan struct{}) {
	defer conn.Close()
	args, err := f.ParseArgs(&conn.Req.Args)
	if err != nil {
		log.Printf("Error parsing PT args: %s", err.Error())
		conn.Reject()
		return
	}
	dialFn := proxy.Direct.Dial
	if proxyURL != nil {
		dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			log.Printf("Error getting proxy dialer: %s", err.Error())
			conn.Reject()
		}
		dialFn = dialer.Dial
	}
	remote, err := f.Dial("tcp", conn.Req.Target, dialFn, args)
	if err != nil {
		log.Printf("Error dialing PT: %s", err.Error())
		return
	}
	err = conn.Grant(&net.TCPAddr{IP: net.IPv4zero, Port: 0})
	if err != nil {
		log.Printf("conn.Grant error: %s", err)
		return
	}
	defer remote.Close()
	done := make(chan struct{}, 2)
	go copyLoop(conn, remote, done)
	// wait for copy loop to finish or for shutdown signal
	select {
	case <-shutdown:
	case <-done:
		log.Println("copy loop ended")
	}
}

// Exchanges bytes between two ReadWriters.
// (In this case, between a SOCKS connection and a pt conn)
func copyLoop(socks, sfconn io.ReadWriter, done chan struct{}) {
	go func() {
		if _, err := io.Copy(socks, sfconn); err != nil {
			log.Printf("copying transport to SOCKS resulted in error: %v", err)
		}
		done <- struct{}{}
	}()
	go func() {
		if _, err := io.Copy(sfconn, socks); err != nil {
			log.Printf("copying SOCKS to transport resulted in error: %v", err)
		}
		done <- struct{}{}
	}()
}

func (c *Controller) GetLocalAddress(methodName string) string {
	if ln, ok := c.listeners[methodName]; ok {
		return ln.Addr().String()
	}
	return ""
}

func (c *Controller) GetPort(methodName string) int {
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
		file.Close()

		err = os.Remove(tempFile)
	}
	return err
}

func (c *Controller) Start(methodName string, proxy string) {
	var proxyURL *url.URL
	var err error

	ptlog.Noticef("Launched for transport: %v", methodName)

	if proxy != "" {
		proxyURL, err = url.Parse(proxy)
		if err != nil {
			log.Fatalf("Failed to parse proxy address: %s", err.Error())
		}
	}

	switch methodName {
	case "snowflake":
		iceServers := strings.Split(strings.TrimSpace(c.SnowflakeIceServers), ",")
		frontDomains := strings.Split(strings.TrimSpace(c.SnowflakeFrontDomains), ",")
		config := &sf.ClientConfig{
			BrokerURL:    c.SnowflakeBrokerUrl,
			AmpCacheURL:  c.SnowflakeAmpCacheUrl,
			SQSQueueURL:  c.SnowflakeSqsUrl,
			SQSCredsStr:  c.SnowflakeSqsCreds,
			FrontDomains: frontDomains,
			ICEAddresses: iceServers,
		}
		if proxyURL != nil {
			if err := sproxy.CheckProxyProtocolSupport(proxyURL); err != nil {
				log.Printf("Error setting up proxy: %s", err.Error())
				return
			} else {
				config.CommunicationProxy = proxyURL
				client := sproxy.NewSocks5UDPClient(proxyURL)
				conn, err := client.ListenPacket("udp", nil)
				if err != nil {
					log.Printf("Failed to initialize %s: proxy test failure: %s",
						methodName, err.Error())
					conn.Close()
					return
				}
				conn.Close()
			}
		}
		f := newSnowflakeClientFactory(config)
		ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
		if err != nil {
			log.Printf("Failed to initialize %s: %s", methodName, err.Error())
			return
		}
		c.shutdown[methodName] = make(chan struct{})
		c.listeners[methodName] = ln
		go acceptLoop(f, ln, nil, c.shutdown[methodName])
	default:
		// at the moment, everything else is in lyrebird
		t := transports.Get(methodName)
		if t == nil {
			log.Printf("Failed to initialize %s: no such method", methodName)
			return
		}
		f, err := t.ClientFactory(c.stateDir)
		if err != nil {
			log.Printf("Failed to initialize %s: %s", methodName, err.Error())
			return
		}
		ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
		if err != nil {
			log.Printf("Failed to initialize %s: %s", methodName, err.Error())
			return
		}
		c.listeners[methodName] = ln
		c.shutdown[methodName] = make(chan struct{})
		go acceptLoop(f, ln, proxyURL, c.shutdown[methodName])

	}
}

func (c *Controller) Stop(methodName string) {
	if ln, ok := c.listeners[methodName]; ok {
		ln.Close()
		log.Printf("Shutting down %s", methodName)
		close(c.shutdown[methodName])
		delete(c.shutdown, methodName)
		delete(c.listeners, methodName)
	} else {
		log.Printf("No listener for %s", methodName)
	}
}

func SnowflakeVersion() string {
	return sfversion.GetVersion()
}

func LyrebirdVersion() string {
	return "lyrebird-0.2.0"
}
