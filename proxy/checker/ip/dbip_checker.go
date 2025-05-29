package ip

import (
	"fmt"
	"net/http"

	"github.com/bestruirui/bestsub/proxy/info"
)

// JSON Structs for DB-IP
type dbipDemoInfo struct {
	ThreatLevel string `json:"threatLevel"` // e.g., "low", "medium", "high"
}
type dbipResponse struct {
	DemoInfo dbipDemoInfo `json:"demoInfo"`
}

// FetchDBIPRiskScore fetches risk score from DB-IP
func FetchDBIPRiskScore(ipAddr string, client *http.Client) (riskScore info.IPRiskScore, err error) {
	var respData dbipResponse
	// The ip.sh script uses `curl ... "https://db-ip.com/demo/home.php?s=$IP"` and then `echo "$RESPONSE"|jq .`, implying JSON output.
	url := fmt.Sprintf("https://db-ip.com/demo/home.php?s=%s", ipAddr)

	err = genericFetchJSON(url, client, &respData)
	if err != nil {
		return info.IPRiskScoreLow, fmt.Errorf("FetchDBIPRiskScore failed for IP %s: %w", ipAddr, err)
	}
	return mapDBIPScoreToEnum(respData.DemoInfo.ThreatLevel), nil
} 