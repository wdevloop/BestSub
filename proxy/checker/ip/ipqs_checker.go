package ip

import (
	"fmt"
	"net/http"

	"github.com/bestruirui/bestsub/proxy/info"
)

// JSON Structs for IPQS (IPQualityScore)
type ipqsResponse struct {
	FraudScore  int    `json:"fraud_score"`
	CountryCode string `json:"country_code"`
	Proxy       bool   `json:"proxy"`
	Tor         bool   `json:"tor"`
	VPN         bool   `json:"vpn"`
	RecentAbuse bool   `json:"recent_abuse"`
	BotStatus   bool   `json:"bot_status"` // Not directly in IPRiskFactor, but available
}

// FetchIPQSRiskScore fetches risk score from IPQS via ipinfo.check.place
func FetchIPQSRiskScore(ipAddr string, client *http.Client) (riskScore info.IPRiskScore, err error) {
	var respData ipqsResponse
	url := fmt.Sprintf("https://ipinfo.check.place/%s?db=ipqualityscore", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		return info.IPRiskScoreLow, fmt.Errorf("FetchIPQSRiskScore failed for IP %s: %w", ipAddr, err)
	}
	return mapIPQSScoreToEnum(respData.FraudScore), nil
}

// FetchIPQSRiskFactors fetches risk factors from IPQS via ipinfo.check.place
func FetchIPQSRiskFactors(ipAddr string, client *http.Client) (factors info.IPRiskFactor, err error) {
	var respData ipqsResponse
	url := fmt.Sprintf("https://ipinfo.check.place/%s?db=ipqualityscore", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		return info.IPRiskFactor{}, fmt.Errorf("FetchIPQSRiskFactors failed for IP %s: %w", ipAddr, err)
	}

	factors.Country = respData.CountryCode
	factors.Proxy = respData.Proxy
	factors.Tor = respData.Tor
	factors.VPN = respData.VPN
	factors.Abuse = respData.RecentAbuse
	// factors.Hosting is not directly provided by IPQS in this structure
	// factors.Spam is also not distinct here (covered by Abuse)

	return factors, nil
} 