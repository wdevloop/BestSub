package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bestruirui/bestsub/utils/log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// This file will house the core logic for network quality checks.

// Province represents a Chinese province for testing
type Province struct {
	Code     string `json:"code"`
	Short    string `json:"short"`
	Zipcode  int    `json:"zipcode"`
	Name     string `json:"name"`
	Province int    `json:"province"`
}

// DelayTestTarget represents a target for delay testing
type DelayTestTarget struct {
	Province string
	ISP      string // CT, CU, CM (ChinaTelecom, ChinaUnicom, ChinaMobile)
	Host     string
}

// RouteTestTarget represents a target for route testing
type RouteTestTarget struct {
	Province string
	ISP      string
	Host     string
	Protocol string // TCP, UDP
}

// InternationalTestTarget represents international targets
type InternationalTestTarget struct {
	Country string
	City    string
	Host    string
}

// NetQualityChecker provides network quality testing functionality
type NetQualityChecker struct {
	provinces            []Province
	delayTargets         []DelayTestTarget
	routeTargets         []RouteTestTarget
	internationalTargets []InternationalTestTarget
}

// NewNetQualityChecker creates a new NetQualityChecker instance
func NewNetQualityChecker() (*NetQualityChecker, error) {
	checker := &NetQualityChecker{}

	// Ensure resource files are available
	if err := EnsureResourceFiles(); err != nil {
		return nil, fmt.Errorf("failed to ensure resource files: %w", err)
	}

	// Load provinces data
	if err := checker.loadProvinces(); err != nil {
		return nil, fmt.Errorf("failed to load provinces: %w", err)
	}

	// Initialize test targets
	checker.initializeTestTargets()

	return checker, nil
}

// loadProvinces loads province data from the resource file
func (nqc *NetQualityChecker) loadProvinces() error {
	provincePath, err := GetResourcePath("province.json")
	if err != nil {
		return fmt.Errorf("failed to get province.json path: %w", err)
	}

	data, err := os.ReadFile(provincePath)
	if err != nil {
		return fmt.Errorf("failed to read province file: %w", err)
	}

	if err := json.Unmarshal(data, &nqc.provinces); err != nil {
		return fmt.Errorf("failed to unmarshal province data: %w", err)
	}

	return nil
}

// initializeTestTargets initializes delay and route test targets
func (nqc *NetQualityChecker) initializeTestTargets() {
	// Initialize delay test targets for major provinces
	majorProvinces := []string{"BJ", "SH", "GD"} // Beijing, Shanghai, Guangdong

	for _, provinceCode := range majorProvinces {
		// Add targets for each ISP
		nqc.delayTargets = append(nqc.delayTargets, []DelayTestTarget{
			{Province: provinceCode, ISP: "CT", Host: fmt.Sprintf("%s-ct-v4.ip.zstaticcdn.com", strings.ToLower(provinceCode))},
			{Province: provinceCode, ISP: "CU", Host: fmt.Sprintf("%s-cu-v4.ip.zstaticcdn.com", strings.ToLower(provinceCode))},
			{Province: provinceCode, ISP: "CM", Host: fmt.Sprintf("%s-cm-v4.ip.zstaticcdn.com", strings.ToLower(provinceCode))},
		}...)

		// Add route test targets
		nqc.routeTargets = append(nqc.routeTargets, []RouteTestTarget{
			{Province: provinceCode, ISP: "CT", Host: fmt.Sprintf("%s-ct-v4.ip.zstaticcdn.com", strings.ToLower(provinceCode)), Protocol: "TCP"},
			{Province: provinceCode, ISP: "CU", Host: fmt.Sprintf("%s-cu-v4.ip.zstaticcdn.com", strings.ToLower(provinceCode)), Protocol: "TCP"},
			{Province: provinceCode, ISP: "CM", Host: fmt.Sprintf("%s-cm-v4.ip.zstaticcdn.com", strings.ToLower(provinceCode)), Protocol: "TCP"},
			{Province: provinceCode, ISP: "CT", Host: fmt.Sprintf("%s-ct-v4.ip.zstaticcdn.com", strings.ToLower(provinceCode)), Protocol: "UDP"},
			{Province: provinceCode, ISP: "CU", Host: fmt.Sprintf("%s-cu-v4.ip.zstaticcdn.com", strings.ToLower(provinceCode)), Protocol: "UDP"},
			{Province: provinceCode, ISP: "CM", Host: fmt.Sprintf("%s-cm-v4.ip.zstaticcdn.com", strings.ToLower(provinceCode)), Protocol: "UDP"},
		}...)
	}

	// Initialize international test targets
	nqc.internationalTargets = []InternationalTestTarget{
		{Country: "US", City: "Los Angeles", Host: "lax.connectivitycheck.gstatic.com"},
		{Country: "US", City: "New York", Host: "ny.connectivitycheck.gstatic.com"},
		{Country: "JP", City: "Tokyo", Host: "nrt.connectivitycheck.gstatic.com"},
		{Country: "KR", City: "Seoul", Host: "icn.connectivitycheck.gstatic.com"},
		{Country: "SG", City: "Singapore", Host: "sin.connectivitycheck.gstatic.com"},
		{Country: "DE", City: "Frankfurt", Host: "fra.connectivitycheck.gstatic.com"},
		{Country: "GB", City: "London", Host: "lhr.connectivitycheck.gstatic.com"},
	}
}

// CheckNetworkQuality performs comprehensive network quality checks
func (c *Checker) CheckNetworkQuality() error {
	nqc, err := NewNetQualityChecker()
	if err != nil {
		log.Error("Failed to create NetQualityChecker: %v", err)
		return err
	}

	log.Info("Starting network quality check for proxy: %v", c.Proxy.Raw["name"])

	// Initialize NetInfo if not already done
	if c.Proxy.Info.Net.Latency.ChinaTelecom == nil {
		c.Proxy.Info.Net.Latency.ChinaTelecom = make(map[string]uint16)
		c.Proxy.Info.Net.Latency.ChinaUnicom = make(map[string]uint16)
		c.Proxy.Info.Net.Latency.ChinaMobile = make(map[string]uint16)
		c.Proxy.Info.Net.Latency.International = make(map[string]uint16)
		c.Proxy.Info.Net.Route = make(map[string]string)
	}

	// Check delays (with timeout for each test)
	if err := nqc.checkDelays(c); err != nil {
		log.Debug("Delays check completed with some errors: %v", err)
	}

	// Check routes (with timeout for each test)
	if err := nqc.checkRoutes(c); err != nil {
		log.Debug("Routes check completed with some errors: %v", err)
	}

	// Check international delays using iperf targets
	if err := nqc.checkInternationalDelaysFromIperf(c); err != nil {
		log.Debug("International iperf delays check completed with some errors: %v", err)
	}

	// Check international delays using standard targets
	if err := nqc.checkInternationalDelays(c); err != nil {
		log.Debug("International delays check completed with some errors: %v", err)
	}

	log.Info("Completed network quality check for proxy: %v", c.Proxy.Raw["name"])
	return nil
}

// checkDelays performs TCP delay tests to Chinese ISPs
func (nqc *NetQualityChecker) checkDelays(c *Checker) error {
	ctx, cancel := context.WithTimeout(c.Proxy.Ctx, 60*time.Second)
	defer cancel()

	for _, target := range nqc.delayTargets {
		delay, err := nqc.measureTCPDelay(ctx, target.Host, c)
		if err != nil {
			log.Debug("Failed to measure delay to %s: %v", target.Host, err)
			continue
		}

		// Store delay based on ISP
		key := fmt.Sprintf("%s-%s", target.Province, target.ISP)
		switch target.ISP {
		case "CT":
			c.Proxy.Info.Net.Latency.ChinaTelecom[key] = delay
		case "CU":
			c.Proxy.Info.Net.Latency.ChinaUnicom[key] = delay
		case "CM":
			c.Proxy.Info.Net.Latency.ChinaMobile[key] = delay
		}

		log.Debug("Delay to %s (%s): %d ms", target.Host, key, delay)
	}

	return nil
}

// checkInternationalDelaysFromIperf uses iperf targets for international delay testing
func (nqc *NetQualityChecker) checkInternationalDelaysFromIperf(c *Checker) error {
	// Load iperf targets from resource or use defaults
	var iperfTargets []InternationalTestTarget

	// Try to load from iperf.json, fallback to defaults
	targets, err := nqc.loadIperfTargetsAsInternational()
	if err != nil {
		log.Debug("Failed to load iperf targets from file, using defaults: %v", err)
		iperfTargets = nqc.getDefaultInternationalTargets()
	} else {
		iperfTargets = targets
	}

	ctx, cancel := context.WithTimeout(c.Proxy.Ctx, 120*time.Second)
	defer cancel()

	// Test each target with limited concurrency
	for _, target := range iperfTargets {
		delay, err := nqc.measureTCPDelay(ctx, target.Host, c)
		if err != nil {
			log.Debug("Failed to measure international delay to %s: %v", target.Host, err)
			continue
		}

		key := fmt.Sprintf("%s-%s", target.Country, target.City)
		c.Proxy.Info.Net.Latency.International[key] = delay

		log.Debug("International delay to %s (%s): %d ms", target.Host, key, delay)

		// Add small delay between tests to avoid overwhelming the proxy
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

// loadIperfTargetsAsInternational loads iperf targets and converts them to international targets
func (nqc *NetQualityChecker) loadIperfTargetsAsInternational() ([]InternationalTestTarget, error) {
	iperfPath, err := GetResourcePath("iperf.json")
	if err != nil {
		return nil, fmt.Errorf("failed to get iperf.json path: %w", err)
	}

	data, err := os.ReadFile(iperfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read iperf file: %w", err)
	}

	// Define a local struct that matches the iperf.json format
	type IperfTarget struct {
		Code   int    `json:"code"`
		Server string `json:"server"`
		PortL  int    `json:"portl"`
		PortU  int    `json:"portu"`
		City   string `json:"city"`
		CityZh string `json:"cityzh"`
	}

	var iperfTargets []IperfTarget
	if err := json.Unmarshal(data, &iperfTargets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal iperf data: %w", err)
	}

	// Convert to international targets
	var targets []InternationalTestTarget
	for _, iperf := range iperfTargets {
		country := nqc.getCountryFromServer(iperf.Server)
		targets = append(targets, InternationalTestTarget{
			Country: country,
			City:    iperf.City,
			Host:    iperf.Server,
		})
	}

	return targets, nil
}

// getCountryFromServer attempts to determine country from server hostname
func (nqc *NetQualityChecker) getCountryFromServer(server string) string {
	serverLower := strings.ToLower(server)

	if strings.Contains(serverLower, "he.net") || strings.Contains(serverLower, "fremont") {
		return "US"
	} else if strings.Contains(serverLower, "lon") || strings.Contains(serverLower, "london") {
		return "GB"
	} else if strings.Contains(serverLower, "fra") || strings.Contains(serverLower, "frankfurt") {
		return "DE"
	} else if strings.Contains(serverLower, "nyc") || strings.Contains(serverLower, "newyork") {
		return "US"
	} else if strings.Contains(serverLower, "dal") || strings.Contains(serverLower, "dallas") {
		return "US"
	} else if strings.Contains(serverLower, "la") || strings.Contains(serverLower, "losangeles") {
		return "US"
	} else if strings.Contains(serverLower, "wdc") || strings.Contains(serverLower, "washington") {
		return "US"
	} else if strings.Contains(serverLower, "ams") || strings.Contains(serverLower, "amsterdam") {
		return "NL"
	} else if strings.Contains(serverLower, "tokyo") {
		return "JP"
	} else if strings.Contains(serverLower, "singapore") {
		return "SG"
	}

	return "Unknown"
}

// getDefaultInternationalTargets returns default international targets
func (nqc *NetQualityChecker) getDefaultInternationalTargets() []InternationalTestTarget {
	return []InternationalTestTarget{
		{Country: "US", City: "Fremont", Host: "iperf3.he.net"},
		{Country: "GB", City: "London", Host: "lon.speedtest.clouvider.net"},
		{Country: "DE", City: "Frankfurt", Host: "fra.speedtest.clouvider.net"},
		{Country: "US", City: "New York", Host: "nyc.speedtest.clouvider.net"},
		{Country: "US", City: "Dallas", Host: "dal.speedtest.clouvider.net"},
		{Country: "US", City: "Los Angeles", Host: "la.speedtest.clouvider.net"},
		{Country: "JP", City: "Tokyo", Host: "speedtest.tokyo2.linode.com"},
		{Country: "SG", City: "Singapore", Host: "speedtest.singapore.linode.com"},
	}
}

// checkRoutes performs route testing to determine network paths
func (nqc *NetQualityChecker) checkRoutes(c *Checker) error {
	for _, target := range nqc.routeTargets {
		route, err := nqc.traceRoute(target.Host, target.Protocol)
		if err != nil {
			log.Debug("Failed to trace route to %s: %v", target.Host, err)
			continue
		}

		key := fmt.Sprintf("%s-%s-%s", target.Province, target.ISP, target.Protocol)
		c.Proxy.Info.Net.Route[key] = route

		log.Debug("Route to %s (%s): %s", target.Host, key, route)
	}

	return nil
}

// traceRoute performs route tracing using available tools
func (nqc *NetQualityChecker) traceRoute(host, protocol string) (string, error) {
	// Try to use nexttrace first, fallback to mtr
	if route, err := nqc.useNexttrace(host, protocol); err == nil {
		return route, nil
	}

	if route, err := nqc.useMTR(host, protocol); err == nil {
		return route, nil
	}

	// Fallback to basic route detection
	return nqc.basicRouteDetection(host)
}

// useNexttrace uses nexttrace tool for route tracing
func (nqc *NetQualityChecker) useNexttrace(host, protocol string) (string, error) {
	var args []string
	args = append(args, "-4", "--raw", "--psize", "1400", "-q", "8")

	switch strings.ToLower(protocol) {
	case "tcp":
		args = append(args, "--tcp", "-p", "80")
	case "udp":
		args = append(args, "--udp", "-p", "80")
	}

	args = append(args, host)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "nexttrace", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("nexttrace failed: %w", err)
	}

	return nqc.parseNexttraceOutput(string(output))
}

// useMTR uses mtr tool for route tracing
func (nqc *NetQualityChecker) useMTR(host, protocol string) (string, error) {
	var args []string
	args = append(args, "-4", "-C", "-G", "1", "-s", "1400", "-c", "1", "-f", "100")

	switch strings.ToLower(protocol) {
	case "tcp":
		args = append(args, "--tcp", "-P", "80")
	case "udp":
		args = append(args, "--udp", "-P", "80")
	}

	args = append(args, host)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "mtr", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("mtr failed: %w", err)
	}

	return nqc.parseMTROutput(string(output))
}

// basicRouteDetection provides basic route detection without external tools
func (nqc *NetQualityChecker) basicRouteDetection(host string) (string, error) {
	// Resolve host to IP
	ips, err := net.LookupIP(host)
	if err != nil {
		return "", fmt.Errorf("failed to resolve %s: %w", host, err)
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no IPs found for %s", host)
	}

	ip := ips[0].String()

	// Basic classification based on IP ranges (simplified)
	if strings.HasPrefix(ip, "202.97") {
		return "CN2GT", nil
	} else if strings.HasPrefix(ip, "59.43") {
		return "CN2GIA", nil
	} else if strings.Contains(host, "ct") || strings.Contains(host, "telecom") {
		return "163", nil
	} else if strings.Contains(host, "cu") || strings.Contains(host, "unicom") {
		return "4837", nil
	} else if strings.Contains(host, "cm") || strings.Contains(host, "mobile") {
		return "CMI", nil
	}

	return "Unknown", nil
}

// parseNexttraceOutput parses nexttrace output to determine route type
func (nqc *NetQualityChecker) parseNexttraceOutput(output string) (string, error) {
	lines := strings.Split(output, "\n")
	var asns []string

	for _, line := range lines {
		// Look for AS numbers in the output
		if strings.Contains(line, "AS") {
			re := regexp.MustCompile(`AS(\d+)`)
			matches := re.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					asns = append(asns, match[1])
				}
			}
		}
	}

	return nqc.classifyRouteByASN(asns), nil
}

// parseMTROutput parses mtr output to determine route type
func (nqc *NetQualityChecker) parseMTROutput(output string) (string, error) {
	lines := strings.Split(output, "\n")
	var asns []string

	for _, line := range lines {
		if strings.Contains(line, "AS") {
			parts := strings.Split(line, ",")
			if len(parts) > 6 {
				asn := strings.TrimSpace(parts[6])
				if strings.HasPrefix(asn, "AS") {
					asns = append(asns, strings.TrimPrefix(asn, "AS"))
				}
			}
		}
	}

	return nqc.classifyRouteByASN(asns), nil
}

// classifyRouteByASN classifies route type based on ASN sequence
func (nqc *NetQualityChecker) classifyRouteByASN(asns []string) string {
	asnSet := make(map[string]bool)
	for _, asn := range asns {
		asnSet[asn] = true
	}

	// Check for specific ASN patterns
	if asnSet["4809"] && asnSet["23764"] {
		return "CTGGIA"
	} else if asnSet["4809"] {
		return "CN2GIA"
	} else if asnSet["23764"] {
		return "CTGGIA"
	} else if asnSet["4134"] && (asnSet["4809"] || asnSet["23764"]) {
		return "CN2GT"
	} else if asnSet["58807"] {
		return "CMIN2"
	} else if asnSet["9929"] {
		return "9929"
	} else if asnSet["10099"] {
		return "10099"
	} else if asnSet["9808"] || asnSet["58453"] {
		return "CMI"
	} else if asnSet["4837"] {
		return "4837"
	} else if asnSet["4134"] {
		return "163"
	} else if asnSet["4538"] {
		return "CERNET"
	} else if asnSet["7497"] {
		return "CSTNET"
	}

	return "Unknown"
}

// checkInternationalDelays performs delay tests to international targets
func (nqc *NetQualityChecker) checkInternationalDelays(c *Checker) error {
	ctx, cancel := context.WithTimeout(c.Proxy.Ctx, 60*time.Second)
	defer cancel()

	for _, target := range nqc.internationalTargets {
		delay, err := nqc.measureTCPDelay(ctx, target.Host, c)
		if err != nil {
			log.Debug("Failed to measure international delay to %s: %v", target.Host, err)
			continue
		}

		key := fmt.Sprintf("%s-%s", target.Country, target.City)
		c.Proxy.Info.Net.Latency.International[key] = delay

		log.Debug("International delay to %s (%s): %d ms", target.Host, key, delay)
	}

	return nil
}

// measureTCPDelay measures TCP connection delay to a host using the proxy
func (nqc *NetQualityChecker) measureTCPDelay(ctx context.Context, host string, c *Checker) (uint16, error) {
	start := time.Now()

	// Create a connection through the proxy
	conn, err := c.Proxy.Client.Transport.(*http.Transport).DialContext(ctx, "tcp", host+":80")
	if err != nil {
		return 0, fmt.Errorf("failed to connect to %s: %w", host, err)
	}
	defer conn.Close()

	delay := time.Since(start)
	delayMs := uint16(delay.Milliseconds())

	// Cap delay at max uint16 value
	if delay.Milliseconds() > 65535 {
		delayMs = 65535
	}

	return delayMs, nil
}
