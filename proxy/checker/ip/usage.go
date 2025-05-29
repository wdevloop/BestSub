package ip

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bestruirui/bestsub/proxy/info"
)

const (
	defaultTimeout = 10 * time.Second
	userAgent      = "BestSub IP Checker/1.0" // Example User-Agent
)

// toIPUsageType maps API type strings to info.IPUsageType
func toIPUsageType(apiTypeRaw string) info.IPUsageType {
	if apiTypeRaw == "" {
		return info.HomeUsage // Default if empty
	}
	normalizedType := strings.ToLower(apiTypeRaw)

	if strings.Contains(normalizedType, "/") {
		parts := strings.Split(normalizedType, "/")
		if len(parts) > 0 {
			normalizedType = parts[0]
		}
	}
	normalizedType = strings.TrimSpace(normalizedType)

	switch normalizedType {
	case "isp", "education", "university/college/school", "fixed line isp", "mobile isp", "edu", "mob", "library":
		return info.HomeUsage
	case "business", "government", "banking", "commercial", "com", "organization", "org", "military", "mil":
		return info.BusinessUsage
	case "hosting", "data center/web hosting/transit", "dch", "cdn", "content delivery network", "ses", "search engine spider":
		return info.HostingUsage
	default:
		// Consider logging unknown types if a logger is available
		// fmt.Printf("Unknown IP usage type: %s, defaulting to HomeUsage\n", apiTypeRaw)
		return info.HomeUsage
	}
}

// GetIPUsageInfo fetches IP usage information from various APIs.
func GetIPUsageInfo(ipAddr string) (info.IPUsageInfo, error) {
	// client is initialized using defaultTimeout from utils.go (same package, so no import needed for defaultTimeout)
	client := &http.Client{Timeout: defaultTimeout}
	usageInfo := info.IPUsageInfo{}

	var collectedErrors []string

	// Fetch from IPInfo
	ipInfoAsnType, ipInfoCompanyType, err := FetchIPInfoUsageData(ipAddr, client)
	if err != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("IPInfo error: %v", err))
	} else {
		usageInfo.IPInfo = toIPUsageType(ipInfoAsnType)
		usageInfo.IPInfoHost = toIPUsageType(ipInfoCompanyType)
	}

	// Fetch from IPRegistry
	ipRegistryConnType, ipRegistryCompType, err := FetchIPRegistryUsageData(ipAddr, client)
	if err != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("IPRegistry error: %v", err))
	} else {
		usageInfo.IPRegistry = toIPUsageType(ipRegistryConnType)
		usageInfo.IPRegistryHost = toIPUsageType(ipRegistryCompType)
	}

	// Fetch from IPApi
	ipApiAsnType, ipApiCompType, err := FetchIPApiUsageData(ipAddr, client)
	if err != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("IPApi error: %v", err))
	} else {
		usageInfo.IPApi = toIPUsageType(ipApiAsnType)
		usageInfo.IPApiHost = toIPUsageType(ipApiCompType)
	}

	// Fetch from AbuseIPDB
	abuseIPDBUsageType, err := FetchAbuseIPDBUsageData(ipAddr, client)
	if err != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("AbuseIPDB error: %v", err))
	} else {
		usageInfo.AbuseIPDB = toIPUsageType(abuseIPDBUsageType)
	}

	// Fetch from IP2Location
	ip2LocationUsageType, err := FetchIP2LocationUsageData(ipAddr, client)
	if err != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("IP2Location error: %v", err))
	} else {
		usageInfo.IP2Location = toIPUsageType(ip2LocationUsageType)
	}

	if len(collectedErrors) > 0 {
		numApis := 5 // Total number of APIs being called
		if len(collectedErrors) == numApis {
			return usageInfo, fmt.Errorf("all IP usage API calls failed: %s", strings.Join(collectedErrors, "; "))
		}
		// Return partially filled info and a combined error message for partial failures
		return usageInfo, fmt.Errorf("some IP usage API calls failed: %s", strings.Join(collectedErrors, "; "))
	}

	return usageInfo, nil
}
