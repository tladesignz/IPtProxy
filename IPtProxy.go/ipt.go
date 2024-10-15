package IPtProxy

import (
	"errors"
	"io"
	"io/fs"
	"log"
	"net"
	"net/url"
	"os"

	pt "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/transports"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/transports/base"
	sf "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/client/lib"
	sproxy "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/common/proxy"
	"golang.org/x/net/proxy"
)

type IPtProxy struct {
	SnowflakeConfig sf.ClientConfig
	listeners       map[string]*pt.SocksListener
	shutdown        chan struct{}
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

func (p *IPtProxy) GetLocalAddress(methodName string) net.Addr {
	return p.listeners[methodName].Addr()
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

func (p *IPtProxy) StartTransports(methodNames []string, stateDir, logLevel string, enableLogging, unsafeLogging bool, proxyURL *url.URL) {

	// TODO: set up logging

	err := createStateDir(stateDir)
	if err != nil {
		log.Fatalf("Failed to set up state directory: %s", err)
	}

	p.shutdown = make(chan struct{})
	listeners := make(map[string]*pt.SocksListener, 0)
	for _, methodName := range methodNames {
		switch methodName {
		case "snowflake":
			if proxyURL != nil {
				if err := sproxy.CheckProxyProtocolSupport(proxyURL); err != nil {
					continue
				} else {
					p.SnowflakeConfig.CommunicationProxy = proxyURL
					client := sproxy.NewSocks5UDPClient(p.SnowflakeConfig.CommunicationProxy)
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
			f := newSnowflakeClientFactory(p.SnowflakeConfig)
			ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
			if err != nil {
				log.Printf("Failed to initialize %s: %s", methodName, err.Error())
				break
			}
			go acceptLoop(f, ln, nil, p.shutdown)
			listeners[methodName] = ln
		default:
			// at the moment, everything else is in lyrebird
			t := transports.Get(methodName)
			if t == nil {
				log.Printf("Failed to initialize %s: no such method", methodName)
				continue
			}
			f, err := t.ClientFactory(stateDir)
			if err != nil {
				log.Printf("Failed to initialize %s: %s", methodName, err.Error())
				continue
			}
			ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
			if err != nil {
				log.Printf("Failed to initialize %s: %s", methodName, err.Error())
				break
			}
			listeners[methodName] = ln
			go acceptLoop(f, ln, proxyURL, p.shutdown)

		}
	}
	p.listeners = listeners
}

func (p *IPtProxy) StopTransports() {
	for _, ln := range p.listeners {
		ln.Close()
	}
	close(p.shutdown)
}
