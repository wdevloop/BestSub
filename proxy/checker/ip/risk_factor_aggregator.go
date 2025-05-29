package ip

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bestruirui/bestsub/proxy/info"
)

// GetIPRiskFactorInfo fetches and aggregates IP risk factors from various APIs.
func GetIPRiskFactorInfo(ipAddr string) (info.IPRiskFactorInfo, error) {
	client := &http.Client{Timeout: defaultTimeout} // Using defaultTimeout from utils.go
	factorInfo := info.IPRiskFactorInfo{}

	var collectedErrors []string

	// IP2Location Factors
	ip2locationFactors, errIP2Location := FetchIP2LocationRiskFactors(ipAddr, client)
	if errIP2Location != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("IP2LocationFactors error: %v", errIP2Location))
	} else {
		factorInfo.IP2Location = ip2locationFactors
	}

	// IPApi Factors
	ipapiFactors, errIPAPI := FetchIPAPIRiskFactors(ipAddr, client)
	if errIPAPI != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("IPAPIFactors error: %v", errIPAPI))
	} else {
		factorInfo.IPApi = ipapiFactors
	}

	// IPRegistry Factors
	ipregistryFactors, errIPRegistry := FetchIPRegistryRiskFactors(ipAddr, client)
	if errIPRegistry != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("IPRegistryFactors error: %v", errIPRegistry))
	} else {
		factorInfo.IPRegistry = ipregistryFactors
	}

	// IPQS Factors
	ipqsFactors, errIPQS := FetchIPQSRiskFactors(ipAddr, client)
	if errIPQS != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("IPQSFactors error: %v", errIPQS))
	} else {
		factorInfo.IPQS = ipqsFactors
	}

	// Scamalytics Factors
	// FetchScamalyticsRiskData returns: riskScore, countryCode, vpn, tor, server, proxy, err
	_, smCountry, smVPN, smTor, smServer, smProxy, errScamalytics := FetchScamalyticsRiskData(ipAddr, client)
	if errScamalytics != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("ScamalyticsFactors error: %v", errScamalytics))
	} else {
		factorInfo.Scamalytics.Country = smCountry
		factorInfo.Scamalytics.VPN = smVPN
		factorInfo.Scamalytics.Tor = smTor
		factorInfo.Scamalytics.Hosting = smServer // server maps to Hosting
		factorInfo.Scamalytics.Proxy = smProxy
		// Abuse from Scamalytics is primarily reflected in its score, not a direct boolean factor here.
	}

	// IPData Factors
	ipdataFactors, errIPData := FetchIPDataRiskFactors(ipAddr, client)
	if errIPData != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("IPDataFactors error: %v", errIPData))
	} else {
		factorInfo.IPData = ipdataFactors
	}

	// IPInfo Factors
	ipinfoFactors, errIPInfo := FetchIPInfoRiskFactors(ipAddr, client)
	if errIPInfo != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("IPInfoFactors error: %v", errIPInfo))
	} else {
		factorInfo.IPInfo = ipinfoFactors
	}

	// IPWhois Factors
	ipwhoisFactors, errIPWhois := FetchIPWhoisRiskFactors(ipAddr, client)
	if errIPWhois != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("IPWhoisFactors error: %v", errIPWhois))
	} else {
		factorInfo.IPWhois = ipwhoisFactors
	}

	if len(collectedErrors) > 0 {
		numFactorAPIs := 8 // Number of APIs/sources attempted for factors
		if len(collectedErrors) == numFactorAPIs {
			return factorInfo, fmt.Errorf("all IP risk factor API calls failed: %s", strings.Join(collectedErrors, "; "))
		}
		return factorInfo, fmt.Errorf("some IP risk factor API calls failed: %s", strings.Join(collectedErrors, "; "))
	}

	return factorInfo, nil
} 