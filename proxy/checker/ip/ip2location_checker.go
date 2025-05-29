package ip

import (
	"fmt"
	"net/http"
)

// JSON Structs for IP2Location
type ip2LocationProxyInfo struct {
	IsPublicProxy bool `json:"is_public_proxy"`
	IsWebProxy    bool `json:"is_web_proxy"`
	IsTor         bool `json:"is_tor"`
	IsVPN         bool `json:"is_vpn"`
	IsDataCenter  bool `json:"is_data_center"`
	IsSpammer     bool `json:"is_spammer"`
}
type ip2LocationResponse struct {
	UsageType   string               `json:"usage_type"`
	CountryCode string               `json:"country_code"`
	Proxy       ip2LocationProxyInfo `json:"proxy"`
}

// FetchIP2LocationUsageData fetches usage type from ipinfo.check.place (for ip2location)
func FetchIP2LocationUsageData(ipAddr string, client *http.Client) (usageType string, err error) {
	var respData ip2LocationResponse
	url := fmt.Sprintf("https://ipinfo.check.place/%s?db=ip2location", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		return "", fmt.Errorf("FetchIP2LocationUsageData failed for IP %s: %w", ipAddr, err)
	}
	return respData.UsageType, nil
}

// FetchIP2LocationRiskFactors fetches risk factors from IP2Location via ipinfo.check.place
func FetchIP2LocationRiskFactors(ipAddr string, client *http.Client) (factors info.IPRiskFactor, err error) {
	var respData ip2LocationResponse
	url := fmt.Sprintf("https://ipinfo.check.place/%s?db=ip2location", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		return info.IPRiskFactor{}, fmt.Errorf("FetchIP2LocationRiskFactors failed for IP %s: %w", ipAddr, err)
	}

	factors.Country = respData.CountryCode
	factors.Proxy = respData.Proxy.IsPublicProxy || respData.Proxy.IsWebProxy
	factors.Tor = respData.Proxy.IsTor
	factors.VPN = respData.Proxy.IsVPN
	factors.Hosting = respData.Proxy.IsDataCenter
	factors.Abuse = respData.Proxy.IsSpammer // Mapping IsSpammer to Abuse
	// factors.Spam remains false

	return factors, nil
} 