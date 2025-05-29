package ip

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bestruirui/bestsub/proxy/info"
)

// JSON Structs for IPApi
type ipApiASN struct {
	Type string `json:"type"`
}
type ipApiCompany struct {
	Type string `json:"type"`
	// For Risk Score
	AbuserScore string `json:"abuser_score"` // e.g., "0 (Very Low)"
}
type ipApiLocation struct {
	CountryCode string `json:"country_code"`
}
type ipApiResponse struct {
	ASN     ipApiASN     `json:"asn"`
	Company ipApiCompany `json:"company"`
	// For Risk Factors
	Location     ipApiLocation `json:"location"`
	IsProxy      bool          `json:"is_proxy"`
	IsTor        bool          `json:"is_tor"`
	IsVPN        bool          `json:"is_vpn"`
	IsDatacenter bool          `json:"is_datacenter"`
	IsAbuser     bool          `json:"is_abuser"`
}

// FetchIPApiUsageData fetches usage types from api.ipapi.is
func FetchIPApiUsageData(ipAddr string, client *http.Client) (asnType string, companyType string, err error) {
	var respData ipApiResponse
	url := fmt.Sprintf("https://api.ipapi.is/?q=%s", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		return "", "", fmt.Errorf("FetchIPApiUsageData failed for IP %s: %w", ipAddr, err)
	}
	return respData.ASN.Type, respData.Company.Type, nil
}

// FetchIPAPIRiskScore fetches risk score from api.ipapi.is
func FetchIPAPIRiskScore(ipAddr string, client *http.Client) (riskScore info.IPRiskScore, err error) {
	var respData ipApiResponse
	url := fmt.Sprintf("https://api.ipapi.is/?q=%s", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		return info.IPRiskScoreLow, fmt.Errorf("FetchIPAPIRiskScore failed for IP %s: %w", ipAddr, err)
	}

	// Extract the text part of the score, e.g., "Very Low" from "0 (Very Low)"
	scoreText := respData.Company.AbuserScore
	if strings.Contains(scoreText, "(") {
		parts := strings.SplitN(scoreText, "(", 2)
		if len(parts) == 2 {
			scoreText = strings.TrimRight(parts[1], ")")
		}
	}
	return mapIPAPIScoreToEnum(scoreText), nil
}

// FetchIPAPIRiskFactors fetches risk factors from api.ipapi.is
func FetchIPAPIRiskFactors(ipAddr string, client *http.Client) (factors info.IPRiskFactor, err error) {
	var respData ipApiResponse
	url := fmt.Sprintf("https://api.ipapi.is/?q=%s", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		return info.IPRiskFactor{}, fmt.Errorf("FetchIPAPIRiskFactors failed for IP %s: %w", ipAddr, err)
	}

	factors.Country = respData.Location.CountryCode
	factors.Proxy = respData.IsProxy
	factors.Tor = respData.IsTor
	factors.VPN = respData.IsVPN
	factors.Hosting = respData.IsDatacenter // Map IsDatacenter to Hosting
	factors.Abuse = respData.IsAbuser
	// factors.Spam remains false as per plan

	return factors, nil
}
