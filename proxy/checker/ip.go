package checker

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	ipcheck "github.com/bestruirui/bestsub/proxy/checker/ip"
	"github.com/bestruirui/bestsub/utils/log"
)

// getPublicIP retrieves the outbound IP address using the proxy client.
func (c *Checker) getPublicIP() (string, error) {
	ctx, cancel := context.WithTimeout(c.Proxy.Ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api64.ipify.org?format=text", nil)
	if err != nil {
		return "", err
	}

	resp, err := c.Proxy.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ip service returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

// IPTest gathers IP related information using various online services.
func (c *Checker) IPTest() {
	ipAddr, err := c.getPublicIP()
	if err != nil {
		log.Debug("failed to obtain public IP: %v", err)
		return
	}

	if ipAddr == "" {
		return
	}

	if usage, err := ipcheck.GetIPUsageInfo(ipAddr); err == nil {
		c.Proxy.Info.IP.IPUsage = usage
	} else {
		log.Debug("get IP usage info failed: %v", err)
	}

	if risk, err := ipcheck.GetIPRiskInfo(ipAddr); err == nil {
		c.Proxy.Info.IP.IPRisk = risk
	} else {
		log.Debug("get IP risk info failed: %v", err)
	}

	if factors, err := ipcheck.GetIPRiskFactorInfo(ipAddr); err == nil {
		c.Proxy.Info.IP.IPRiskFactor = factors
	} else {
		log.Debug("get IP risk factor info failed: %v", err)
	}

	if banned, err := ipcheck.GetIPBannedInfo(ipAddr); err == nil {
		c.Proxy.Info.IP.IPBanned = banned
	} else {
		log.Debug("get IP banned info failed: %v", err)
	}
}
