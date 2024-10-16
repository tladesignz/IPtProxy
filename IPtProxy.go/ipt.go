package IPtProxy

/**
Package IPtProxy combines tor pluggable transport clients into a single library for use
with mobile applications. Transports, when started, will listen for incoming SOCKS
connections on an available local address and proxy traffic between those connections
and a configured bridge.

Sample gomobile usage:
 import IPtProxy.IPtProxy;

 // Create a new IPtProxy instance with provided state directory
 iptproxy = IPtProxy.newIPtProxy(/path/to/statedir)
 iptproxy.init()

 // Start listening for obfs4 connections, using an outgoing proxy
 String[] transports = {"obfs4", "meek"}
 iptproxy.start(transports, "socks5://localhost:8001")

 // Get the address that is listening for SOCKS connections for each transport
 obfs4Addr = iptproxy.getLocalAddress("obfs4")
 meekAddr = iptproxy.getLocalAddress("meek")

 // Start listening for snowflake connections
 // Note that snowflake setup can happen either here or with SOCKS arguments on
 // a per-connection basis.
 String[] iceServers = {"stun:stun.l.google.com:19302", "stun:stun.l.google.com:5349"};
 iptproxy.setSnowflakeIceServers(iceServers)
 transports = {"snowflake"}
 iptproxy.start(transports, "")


 // Stop transports, either all at once, or individually
 tranports = {"snowflake", "obfs4", "meek_lite"}
 iptproxy.stop(transports)


Sample pure go usage:
 import github.com/tladesignz/IPtProxy

 func main() {
	 iptproxy := NewIPtProxy(/path/to/statedir)
	 iptproxy.Init()

	 iptproxy.Start([]string{"snowflake", "meek_lite", "obfs4"}, "")
	 addr := iptproxy.GetLocalAddress("snowflake")
	 fmt.Printf("Listening for snowflake connections on: %s", addr)

	 // ...

	 iptproxy.Stop([]string{"snowflake", "obfs4", "meek_lite"})
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

	pt "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib"
	ptlog "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/common/log"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/transports"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/transports/base"
	sf "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/client/lib"
	sproxy "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/common/proxy"
	"golang.org/x/net/proxy"
)

type IPtProxy struct {
	EnableLogging bool
	UnsafeLogging bool
	LogLevel      string
	StateDir      string

	SnowflakeIceServers   []string
	SnowflakeBrokerUrl    string
	SnowflakeFrontDomains []string
	SnowflakeAmpCacheUrl  string
	SnowflakeSqsUrl       string
	SnowflakeSqsCreds     string

	listeners map[string]*pt.SocksListener
	shutdown  map[string]chan struct{}
}

func NewIPtProxy(stateDir string) *IPtProxy {
	return &IPtProxy{
		LogLevel:      "ERROR",
		StateDir:      stateDir,
		EnableLogging: true,
	}
}

func (p *IPtProxy) Init() {
	if err := createStateDir(p.StateDir); err != nil {
		log.Fatalf("Failed to set up state directory: %s", err)
	}
	if err := ptlog.Init(p.EnableLogging,
		path.Join(p.StateDir, "ipt.log"), p.UnsafeLogging); err != nil {
		log.Fatalf("Failed to set initialize log: %s", err.Error())
	}
	if p.LogLevel != "" {
		if err := ptlog.SetLogLevel(p.LogLevel); err != nil {
			log.Fatalf("Failed to set log level: %s", err.Error())
		}
	}
	if err := transports.Init(); err != nil {
		log.Fatalf("Failed to initialize transports: %s", err.Error())
	}
	p.listeners = make(map[string]*pt.SocksListener, 0)
	p.shutdown = make(map[string]chan struct{}, 0)
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

func (p *IPtProxy) GetLocalAddress(methodName string) string {
	if ln, ok := p.listeners[methodName]; ok {
		return ln.Addr().String()
	}
	return ""
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

func (p *IPtProxy) Start(methodNames []string, proxy string) {
	var proxyURL *url.URL
	var err error

	ptlog.Noticef("Launced iptproxy for transports: %v", methodNames)

	if proxy != "" {
		proxyURL, err = url.Parse(proxy)
		if err != nil {
			log.Fatalf("Failed to parse proxy address: %s", err.Error())
		}
	}

	for _, methodName := range methodNames {
		switch methodName {
		case "snowflake":
			config := &sf.ClientConfig{
				BrokerURL:    p.SnowflakeBrokerUrl,
				AmpCacheURL:  p.SnowflakeAmpCacheUrl,
				SQSQueueURL:  p.SnowflakeSqsUrl,
				SQSCredsStr:  p.SnowflakeSqsCreds,
				FrontDomains: p.SnowflakeFrontDomains,
				ICEAddresses: p.SnowflakeIceServers,
			}
			if proxyURL != nil {
				if err := sproxy.CheckProxyProtocolSupport(proxyURL); err != nil {
					log.Printf("Error setting up proxy: %s", err.Error())
					continue
				} else {
					config.CommunicationProxy = proxyURL
					client := sproxy.NewSocks5UDPClient(proxyURL)
					conn, err := client.ListenPacket("udp", nil)
					if err != nil {
						log.Printf("Failed to initialize %s: proxy test failure: %s",
							methodName, err.Error())
						conn.Close()
						continue
					}
					conn.Close()
				}
			}
			f := newSnowflakeClientFactory(config)
			ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
			if err != nil {
				log.Printf("Failed to initialize %s: %s", methodName, err.Error())
				break
			}
			p.shutdown[methodName] = make(chan struct{})
			p.listeners[methodName] = ln
			go acceptLoop(f, ln, nil, p.shutdown[methodName])
		default:
			// at the moment, everything else is in lyrebird
			t := transports.Get(methodName)
			if t == nil {
				log.Printf("Failed to initialize %s: no such method", methodName)
				continue
			}
			f, err := t.ClientFactory(p.StateDir)
			if err != nil {
				log.Printf("Failed to initialize %s: %s", methodName, err.Error())
				continue
			}
			ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
			if err != nil {
				log.Printf("Failed to initialize %s: %s", methodName, err.Error())
				break
			}
			p.listeners[methodName] = ln
			p.shutdown[methodName] = make(chan struct{})
			go acceptLoop(f, ln, proxyURL, p.shutdown[methodName])

		}
	}
}

func (p *IPtProxy) Stop(methodNames []string) {
	for _, methodName := range methodNames {
		if ln, ok := p.listeners[methodName]; ok {
			ln.Close()
			log.Printf("Shutting down %s", methodName)
			close(p.shutdown[methodName])
			delete(p.shutdown, methodName)
			delete(p.listeners, methodName)
		} else {
			log.Printf("No listener for %s", methodName)
		}
	}
}
