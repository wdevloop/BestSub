package ip

import (
	"fmt"
	"net/http"

	"github.com/bestruirui/bestsub/proxy/info"
)

// JSON Structs for IPInfo
type ipInfoASN struct {
	Type string `json:"type"`
}
type ipInfoCompany struct {
	Type string `json:"type"`
}
type ipInfoPrivacy struct {
	Proxy   bool `json:"proxy"`
	Tor     bool `json:"tor"`
	VPN     bool `json:"vpn"`
	Hosting bool `json:"hosting"`
}
type ipInfoData struct {
	ASN     ipInfoASN     `json:"asn"`
	Company ipInfoCompany `json:"company"`
	Country string        `json:"country"` // Country code for factors
	Privacy ipInfoPrivacy `json:"privacy"` // For factors
}
type ipInfoResponse struct {
	Data ipInfoData `json:"data"`
}

// FetchIPInfoUsageData fetches usage types from ipinfo.io
func FetchIPInfoUsageData(ipAddr string, client *http.Client) (asnType string, companyType string, err error) {
	var respData ipInfoResponse
	url := fmt.Sprintf("https://ipinfo.io/widget/demo/%s", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		return "", "", fmt.Errorf("FetchIPInfoUsageData failed for IP %s: %w", ipAddr, err)
	}
	return respData.Data.ASN.Type, respData.Data.Company.Type, nil
}

// FetchIPInfoRiskFactors fetches risk factors from ipinfo.io
func FetchIPInfoRiskFactors(ipAddr string, client *http.Client) (factors info.IPRiskFactor, err error) {
	var respData ipInfoResponse
	url := fmt.Sprintf("https://ipinfo.io/widget/demo/%s", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		return info.IPRiskFactor{}, fmt.Errorf("FetchIPInfoRiskFactors failed for IP %s: %w", ipAddr, err)
	}

	factors.Country = respData.Data.Country
	factors.Proxy = respData.Data.Privacy.Proxy
	factors.Tor = respData.Data.Privacy.Tor
	factors.VPN = respData.Data.Privacy.VPN
	factors.Hosting = respData.Data.Privacy.Hosting
	// Abuse is not directly provided by IPInfo in this structure

	return factors, nil
}
