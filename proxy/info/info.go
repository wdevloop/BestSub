package info

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/bestruirui/bestsub/config"
	"github.com/metacubex/mihomo/adapter"
	"github.com/metacubex/mihomo/constant"
)

type Unlock struct {
	Google     bool
	Chatgpt    bool
	Netflix    bool
	Disney     bool
	Youtube    bool
	Cloudflare bool
	Tiktok     bool
	Spotify    bool
	Amazon     bool
}

type IPUsageType int
const (
	HomeUsage IPUsageType = iota
	BusinessUsage
	HostingUsage
)

type IPUsageInfo struct {
	IPInfo 			IPUsageType
	IPInfoHost 		IPUsageType
	IPRegistry 		IPUsageType
	IPRegistryHost 	IPUsageType
	IPApi 			IPUsageType
	IPApiHost 		IPUsageType
	AbuseIPDB 		IPUsageType
	IP2Location 	IPUsageType
}

type IPRiskScore int
const (
	IPRiskScoreVeryLow IPRiskScore = iota
	IPRiskScoreLow
	IPRiskScoreMedium
	IPRiskScoreHigh
	IPRiskScoreVeryHigh
)

type IPRiskInfo struct {
	Scamalytics 	IPRiskScore
	IPAPI 			IPRiskScore
	AbuseIPDB 		IPRiskScore
	IPQS 			IPRiskScore
	DBIP 			IPRiskScore
}

type IPRiskFactor struct {
	Country 	string
	Proxy		bool
	Tor 		bool
	VPN 		bool
	Hosting 	bool
	Abuse 		bool
	Spam 		bool
}

type IPRiskFactorInfo struct {
	IP2Location 	IPRiskFactor
	IPApi 			IPRiskFactor
	IPRegistry 		IPRiskFactor
	IPQS 			IPRiskFactor
	Scamalytics		IPRiskFactor
	IPData		 	IPRiskFactor
	IPInfo 			IPRiskFactor
	IPWhois 		IPRiskFactor
}

type IPBannedInfo struct {
	Normal 		int
	Tagged 		int
	Banned 		int
}

type IPInfo struct {
	IPUsage			IPUsageInfo
	IPRisk			IPRiskInfo
	IPRiskFactor	IPRiskFactorInfo
	IPBanned		IPBannedInfo
}

type NetLatencyInfo struct {
	ChinaTelecom	map[string]uint16
	ChinaUnicom		map[string]uint16
	ChinaMobile		map[string]uint16
	International	map[string]uint16
}

type NetInfo struct {
	Latency		NetLatencyInfo
	Route		map[string]string	// 指不同回程的线路，例如Beijing-ChinaUnicom-TCP的线路等等
}

type ProxyInfo struct {
	Unlock   	 Unlock
	IP			 IPInfo
	Net			 NetInfo
	Speed    	 int
	SpeedSkip 	 bool
	Rate     	 float32
	Risk     	 int
	Delay     	 uint16
	Alive     	 bool
	Country   	 string
	Flag     	 string
}

type Proxy struct {
	Raw    map[string]any
	Id     int
	Ctx    context.Context
	Cancel context.CancelFunc
	Client *http.Client
	Info   ProxyInfo
}

func (p *Proxy) Close() {
	if p.Cancel != nil {
		p.Cancel()
	}
	if transport, ok := p.Client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
}
func (p *Proxy) CloseTransport() {
	if transport, ok := p.Client.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
}
func (p *Proxy) New() error {
	p.Ctx, p.Cancel = context.WithCancel(context.Background())
	proxy, err := adapter.ParseProxy(p.Raw)
	if err != nil {
		return err
	}
	p.Client = &http.Client{
		Timeout:   time.Duration(config.GlobalConfig.Check.Timeout) * time.Millisecond,
		Transport: BuildTransport(proxy, p.Ctx),
	}
	return nil
}
func BuildTransport(proxy constant.Proxy, ctx context.Context) *http.Transport {
	transport := &http.Transport{
		DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			var u16Port uint16
			if port, err := strconv.ParseUint(port, 10, 16); err == nil {
				u16Port = uint16(port)
			}

			return proxy.DialContext(ctx, &constant.Metadata{
				Host:    host,
				DstPort: u16Port,
			})
		},
		MaxIdleConns:          0,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DisableKeepAlives:     true,
		ForceAttemptHTTP2:     false,
		MaxConnsPerHost:       0,
		MaxIdleConnsPerHost:   0,
	}
	return transport
}
