package ip

import (
	"fmt"
	"net/http"

	"github.com/bestruirui/bestsub/proxy/info"
)

// JSON Structs for IPWhois
type ipWhoisSecurity struct {
	Proxy   bool `json:"proxy"`
	Tor     bool `json:"tor"`
	VPN     bool `json:"vpn"`
	Hosting bool `json:"hosting"`
}
type ipWhoisResponse struct {
	CountryCode string            `json:"country_code"`
	Security    ipWhoisSecurity `json:"security"`
}

// FetchIPWhoisRiskFactors fetches risk factors from IPWhois
func FetchIPWhoisRiskFactors(ipAddr string, client *http.Client) (factors info.IPRiskFactor, err error) {
	var respData ipWhoisResponse
	url := fmt.Sprintf("https://ipwhois.io/widget?ip=%s&lang=en", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		return info.IPRiskFactor{}, fmt.Errorf("FetchIPWhoisRiskFactors failed for IP %s: %w", ipAddr, err)
	}

	factors.Country = respData.CountryCode
	factors.Proxy = respData.Security.Proxy
	factors.Tor = respData.Security.Tor
	factors.VPN = respData.Security.VPN
	factors.Hosting = respData.Security.Hosting
	// Abuse is not directly provided by IPWhois in this structure

	return factors, nil
} 