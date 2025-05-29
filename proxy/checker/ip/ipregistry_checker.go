package ip

import (
	"fmt"
	"net/http"
)

// JSON Structs for IPRegistry
type ipRegistryConnection struct {
	Type string `json:"type"`
}
type ipRegistryCompany struct {
	Type string `json:"type"`
}
type ipRegistryLocationCountry struct {
	Code string `json:"code"`
}
type ipRegistryLocation struct {
	Country ipRegistryLocationCountry `json:"country"`
}
type ipRegistrySecurity struct {
	IsProxy         bool `json:"is_proxy"`
	IsTor           bool `json:"is_tor"`
	IsTorExit       bool `json:"is_tor_exit"`
	IsVPN           bool `json:"is_vpn"`
	IsCloudProvider bool `json:"is_cloud_provider"`
	IsAbuser        bool `json:"is_abuser"`
}
type ipRegistryResponse struct {
	Connection ipRegistryConnection `json:"connection"`
	Company    ipRegistryCompany    `json:"company"`
	Location   ipRegistryLocation   `json:"location"`
	Security   ipRegistrySecurity   `json:"security"`
}

// FetchIPRegistryUsageData fetches usage types from ipinfo.check.place (for ipregistry)
func FetchIPRegistryUsageData(ipAddr string, client *http.Client) (connectionType string, companyType string, err error) {
	var respData ipRegistryResponse
	url := fmt.Sprintf("https://ipinfo.check.place/%s?db=ipregistry", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		return "", "", fmt.Errorf("FetchIPRegistryUsageData failed for IP %s: %w", ipAddr, err)
	}
	return respData.Connection.Type, respData.Company.Type, nil
}

// FetchIPRegistryRiskFactors fetches risk factors from IPRegistry via ipinfo.check.place
func FetchIPRegistryRiskFactors(ipAddr string, client *http.Client) (factors info.IPRiskFactor, err error) {
	var respData ipRegistryResponse
	url := fmt.Sprintf("https://ipinfo.check.place/%s?db=ipregistry", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		return info.IPRiskFactor{}, fmt.Errorf("FetchIPRegistryRiskFactors failed for IP %s: %w", ipAddr, err)
	}

	factors.Country = respData.Location.Country.Code
	factors.Proxy = respData.Security.IsProxy
	factors.Tor = respData.Security.IsTor || respData.Security.IsTorExit
	factors.VPN = respData.Security.IsVPN
	factors.Hosting = respData.Security.IsCloudProvider
	factors.Abuse = respData.Security.IsAbuser

	return factors, nil
} 