package ip

import (
	"fmt"
	"net/http"

	"github.com/bestruirui/bestsub/proxy/info"
)

// JSON Structs for AbuseIPDB
type abuseIPDBData struct {
	UsageType string `json:"usageType"`
	// For Risk Score
	AbuseConfidenceScore int `json:"abuseConfidenceScore"`
	// For Risk Factors (limited)
	CountryCode string `json:"countryCode"`
}
type abuseIPDBResponse struct {
	Data abuseIPDBData `json:"data"`
}

// FetchAbuseIPDBUsageData fetches usage type from ipinfo.check.place (for abuseipdb)
func FetchAbuseIPDBUsageData(ipAddr string, client *http.Client) (usageType string, err error) {
	var respData abuseIPDBResponse
	url := fmt.Sprintf("https://ipinfo.check.place/%s?db=abuseipdb", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		return "", fmt.Errorf("FetchAbuseIPDBUsageData failed for IP %s: %w", ipAddr, err)
	}
	return respData.Data.UsageType, nil
}

// FetchAbuseIPDBRiskScore fetches risk score from AbuseIPDB via ipinfo.check.place
func FetchAbuseIPDBRiskScore(ipAddr string, client *http.Client) (riskScore info.IPRiskScore, err error) {
	var respData abuseIPDBResponse
	url := fmt.Sprintf("https://ipinfo.check.place/%s?db=abuseipdb", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		// Return a default score and the error
		return info.IPRiskScoreLow, fmt.Errorf("FetchAbuseIPDBRiskScore failed for IP %s: %w", ipAddr, err)
	}
	return mapAbuseIPDBScoreToEnum(respData.Data.AbuseConfidenceScore), nil
}

// FetchAbuseIPDBRiskFactors fetches risk factors from AbuseIPDB via ipinfo.check.place
// Note: AbuseIPDB via this endpoint primarily gives score and usageType.
// It provides CountryCode, but not other boolean factors like proxy, vpn etc.
func FetchAbuseIPDBRiskFactors(ipAddr string, client *http.Client) (factors info.IPRiskFactor, err error) {
	var respData abuseIPDBResponse
	url := fmt.Sprintf("https://ipinfo.check.place/%s?db=abuseipdb", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		// Return empty factors and the error
		return info.IPRiskFactor{}, fmt.Errorf("FetchAbuseIPDBRiskFactors failed for IP %s: %w", ipAddr, err)
	}

	factors.Country = respData.Data.CountryCode
	// Other factors (Proxy, Tor, VPN, Hosting, Abuse) are not directly provided by this endpoint specific factor list.
	// The 'Abuse' status is generally inferred from the AbuseConfidenceScore, not as a separate boolean factor from this source.
	
	return factors, nil
} 