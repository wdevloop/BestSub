package ip

import (
	"fmt"
	"net/http"

	"github.com/bestruirui/bestsub/proxy/info"
)

// JSON Structs for IPData
type ipDataThreat struct {
	IsProxy         bool `json:"is_proxy"`
	IsTor           bool `json:"is_tor"`
	IsDatacenter    bool `json:"is_datacenter"`
	IsThreat        bool `json:"is_threat"`
	IsKnownAbuser   bool `json:"is_known_abuser"`
	IsKnownAttacker bool `json:"is_known_attacker"`
}
type ipDataResponse struct {
	CountryCode string       `json:"country_code"`
	Threat      ipDataThreat `json:"threat"`
}

// FetchIPDataRiskFactors fetches risk factors from IPData via ipinfo.check.place
func FetchIPDataRiskFactors(ipAddr string, client *http.Client) (factors info.IPRiskFactor, err error) {
	var respData ipDataResponse
	url := fmt.Sprintf("https://ipinfo.check.place/%s?db=ipdata", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		return info.IPRiskFactor{}, fmt.Errorf("FetchIPDataRiskFactors failed for IP %s: %w", ipAddr, err)
	}

	factors.Country = respData.CountryCode
	factors.Proxy = respData.Threat.IsProxy
	factors.Tor = respData.Threat.IsTor
	factors.Hosting = respData.Threat.IsDatacenter
	factors.Abuse = respData.Threat.IsThreat || respData.Threat.IsKnownAbuser || respData.Threat.IsKnownAttacker

	return factors, nil
} 