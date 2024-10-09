package IPtProxy

import (
	"io"
	"log"
	"net"
	"os"

	pt "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/transports"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/transports/base"
	sf "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/client/lib"
	"golang.org/x/net/proxy"
)

type IPtProxy struct {
	SnowflakeConfig sf.ClientConfig
	listeners       map[string]*pt.SocksListener
}

func acceptLoop(f base.ClientFactory, ln *pt.SocksListener) error {
	defer ln.Close()
	for {
		conn, err := ln.AcceptSocks()
		if err != nil {
			if e, ok := err.(net.Error); ok && !e.Temporary() {
				return err
			}
			continue
		}
		go clientHandler(f, conn)
		//TODO: maybe a shutdown channel, make sure to close conns
	}
}

func clientHandler(f base.ClientFactory, conn *pt.SocksConn) {
	args, err := f.ParseArgs(&conn.Req.Args)
	if err != nil {
		log.Printf("Error parsing PT args: %s", err.Error())
		return
	}
	//TODO: when proxy support is added back, update this
	dialFn := proxy.Direct.Dial
	remote, err := f.Dial("tcp", conn.Req.Target, dialFn, args)
	if err != nil {
		log.Printf("Error dialing PT: %s", err.Error())
		return
	}
	copyLoop(conn, remote)
}

// Exchanges bytes between two ReadWriters.
// (In this case, between a SOCKS connection and a pt conn)
func copyLoop(socks, sfconn io.ReadWriter) {
	done := make(chan struct{}, 2)
	go func() {
		if _, err := io.Copy(socks, sfconn); err != nil {
			log.Printf("copying Snowflake to SOCKS resulted in error: %v", err)
		}
		done <- struct{}{}
	}()
	go func() {
		if _, err := io.Copy(sfconn, socks); err != nil {
			log.Printf("copying SOCKS to Snowflake resulted in error: %v", err)
		}
		done <- struct{}{}
	}()
	<-done
	log.Println("copy loop ended")
}

func (p *IPtProxy) GetLocalAddress(methodName string) net.Addr {
	return p.listeners[methodName].Addr()
}

func (p *IPtProxy) StartTransports(stateLocation, logLevel string, enableLogging, unsafeLogging bool, proxy string) {

	// TODO: set up logging

	// goptlib will check for the environemnt variables that PT processes usually
	// get from the tor parent process. In this case, we are starting the PT processes
	// independently and then reporting the SOCKS5 port in the torrc configuration,
	// so we need to set these environment variables manually
	fixEnv(stateLocation)

	stateDir, err := pt.MakeStateDir()
	if err != nil {
		log.Fatalf("No state directory: %s", err)
	}

	ptInfo, err := pt.ClientSetup(nil)
	if err != nil {
		log.Fatal(err)
	}

	// This assumes a major version bump due to breaking changes
	pt.ReportVersion("iptproxy", "4.0.0")

	// TODO: add back in proxy support, which is necessary for iOS
	if ptInfo.ProxyURL != nil {
		pt.ProxyError("proxy is not supported")
		os.Exit(1)
	}

	listeners := make(map[string]*pt.SocksListener, 0)
	for _, methodName := range ptInfo.MethodNames {
		switch methodName {
		case "snowflake":
			f := newSnowflakeClientFactory(p.SnowflakeConfig)
			ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
			if err != nil {
				pt.CmethodError(methodName, err.Error())
				break
			}
			go acceptLoop(f, ln)
			pt.Cmethod(methodName, ln.Version(), ln.Addr())
			listeners[methodName] = ln
		default:
			// at the moment, everything else is in lyrebird
			t := transports.Get(methodName)
			if t == nil {
				pt.CmethodError(methodName, "no such method")
				continue
			}
			f, err := t.ClientFactory(stateDir)
			if err != nil {
				pt.CmethodError(methodName, err.Error())
				continue
			}
			ln, err := pt.ListenSocks("tcp", "127.0.0.1:0")
			if err != nil {
				pt.CmethodError(methodName, err.Error())
				break
			}
			listeners[methodName] = ln
			go acceptLoop(f, ln)

		}
	}
	p.listeners = listeners
	pt.CmethodsDone()
}

func (p *IPtProxy) StopTransports() {
	for _, ln := range p.listeners {
		ln.Close()
	}
}
