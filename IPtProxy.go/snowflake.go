package IPtProxy

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	pt "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/goptlib"
	"gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/lyrebird/transports/base"
	sf "gitlab.torproject.org/tpo/anti-censorship/pluggable-transports/snowflake/v2/client/lib"
)

type snowflakeClientFactory struct {
	transport sf.Transport
	config    sf.ClientConfig
}

func newSnowflakeClientFactory(config sf.ClientConfig) *snowflakeClientFactory {

	return &snowflakeClientFactory{
		config: config,
	}
}

func (cf *snowflakeClientFactory) Transport() base.Transport {
	log.Printf("Transport method not implemented for snowflakeClientFactory")
	return nil
}

func (cf *snowflakeClientFactory) ParseArgs(args *pt.Args) (interface{}, error) {
	if arg, ok := args.Get("ampcache"); ok {
		cf.config.AmpCacheURL = arg
	}
	if arg, ok := args.Get("sqsqueue"); ok {
		cf.config.SQSQueueURL = arg
	}
	if arg, ok := args.Get("sqscreds"); ok {
		cf.config.SQSCredsStr = arg
	}
	if arg, ok := args.Get("fronts"); ok {
		if arg != "" {
			cf.config.FrontDomains = strings.Split(strings.TrimSpace(arg), ",")
		}
	} else if arg, ok := args.Get("front"); ok {
		cf.config.FrontDomains = strings.Split(strings.TrimSpace(arg), ",")
	}
	if arg, ok := args.Get("ice"); ok {
		cf.config.ICEAddresses = strings.Split(strings.TrimSpace(arg), ",")
	}
	if arg, ok := args.Get("max"); ok {
		max, err := strconv.Atoi(arg)
		if err != nil {
			return nil, fmt.Errorf("Invalid SOCKS arg: max=", arg)
		}
		cf.config.Max = max
	}
	if arg, ok := args.Get("url"); ok {
		cf.config.BrokerURL = arg
	}
	if arg, ok := args.Get("utls-nosni"); ok {
		switch strings.ToLower(arg) {
		case "true":
			fallthrough
		case "yes":
			cf.config.UTLSRemoveSNI = true
		}
	}
	if arg, ok := args.Get("utls-imitate"); ok {
		cf.config.UTLSClientID = arg
	}
	if arg, ok := args.Get("fingerprint"); ok {
		cf.config.BridgeFingerprint = arg
	}
	return cf.config, nil
}

func (cf *snowflakeClientFactory) Dial(network, address string, dialFn base.DialFunc, args interface{}) (net.Conn, error) {
	config, ok := args.(*sf.ClientConfig)
	if !ok {
		return nil, fmt.Errorf("invalid type for args")
	}
	transport, err := sf.NewSnowflakeClient(*config)
	if err != nil {
		return nil, err
	}
	return transport.Dial()
}
