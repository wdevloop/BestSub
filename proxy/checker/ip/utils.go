package ip

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/bestruirui/bestsub/proxy/info"
)

const (
	defaultTimeout = 10 * time.Second
	userAgent      = "BestSub IP Checker/1.0"
)

func genericFetchJSON(url string, client *http.Client, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request to %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code from %s: %d", url, resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("failed to decode JSON from %s: %w", url, err)
	}
	return nil
}

// --- Score Mapping Functions ---

func mapScamalyticsScoreToEnum(rawScore int) info.IPRiskScore {
	if rawScore >= 75 {
		return info.IPRiskScoreVeryHigh
	}
	if rawScore >= 50 {
		return info.IPRiskScoreHigh
	}
	if rawScore >= 25 {
		return info.IPRiskScoreMedium
	}
	return info.IPRiskScoreLow
}

func mapIPAPIScoreToEnum(rawScoreText string) info.IPRiskScore {
	switch strings.ToLower(rawScoreText) {
	case "very low":
		return info.IPRiskScoreVeryLow
	case "low":
		return info.IPRiskScoreLow
	case "elevated": // Or "medium"
		return info.IPRiskScoreMedium
	case "high":
		return info.IPRiskScoreHigh
	case "very high":
		return info.IPRiskScoreVeryHigh
	default:
		return info.IPRiskScoreLow // Default for unknown/unparseable
	}
}

func mapAbuseIPDBScoreToEnum(rawScore int) info.IPRiskScore {
	if rawScore >= 75 { // 75+ is considered DoS/VeryHigh risk
		return info.IPRiskScoreVeryHigh
	}
	if rawScore >= 25 { // Between 25 and 74
		return info.IPRiskScoreHigh
	}
	return info.IPRiskScoreLow // 0-24
}

func mapIPQSScoreToEnum(rawScore int) info.IPRiskScore {
	if rawScore >= 90 {
		return info.IPRiskScoreVeryHigh
	}
	if rawScore >= 85 {
		return info.IPRiskScoreHigh // Risky
	}
	if rawScore >= 75 {
		return info.IPRiskScoreMedium // Suspicious
	}
	return info.IPRiskScoreLow
}

func mapDBIPScoreToEnum(rawScoreText string) info.IPRiskScore {
	switch strings.ToLower(rawScoreText) {
	case "low":
		return info.IPRiskScoreLow
	case "medium":
		return info.IPRiskScoreMedium
	case "high":
		return info.IPRiskScoreHigh
	default:
		return info.IPRiskScoreLow // Default for unknown
	}
}

// ValidateIP 验证IP地址格式并返回清理后的IP字符串
func ValidateIP(ipStr string) string {
	// 移除换行符和空白字符
	cleanIP := strings.TrimSpace(ipStr)
	cleanIP = strings.TrimSuffix(cleanIP, "\n")
	cleanIP = strings.TrimSuffix(cleanIP, "\r")

	// 使用正则表达式提取IP地址
	ipv4Regex := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)
	ipv6Regex := regexp.MustCompile(`([0-9a-fA-F:]+::[0-9a-fA-F:]*|[0-9a-fA-F:]*::[0-9a-fA-F:]+|[0-9a-fA-F:]+:[0-9a-fA-F:]+:[0-9a-fA-F:]+:[0-9a-fA-F:]+:[0-9a-fA-F:]+:[0-9a-fA-F:]+:[0-9a-fA-F:]+:[0-9a-fA-F:]+)`)

	// 先尝试IPv4
	if match := ipv4Regex.FindString(cleanIP); match != "" {
		if net.ParseIP(match) != nil {
			return match
		}
	}

	// 再尝试IPv6
	if match := ipv6Regex.FindString(cleanIP); match != "" {
		if net.ParseIP(match) != nil {
			return match
		}
	}

	// 直接验证清理后的字符串
	if net.ParseIP(cleanIP) != nil {
		return cleanIP
	}

	return ""
}
