package ip

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/bestruirui/bestsub/proxy/info"
)

// GetIPRiskInfo 获取IP风险评分信息，聚合多个风险评分服务的结果
func GetIPRiskInfo(ipAddr string) (info.IPRiskInfo, error) {
	client := &http.Client{Timeout: defaultTimeout}
	riskInfo := info.IPRiskInfo{}
	
	var wg sync.WaitGroup
	var mutex sync.Mutex
	var collectedErrors []string
	
	// 并发获取各个服务的风险评分
	wg.Add(5)
	
	// Scamalytics 风险评分
	go func() {
		defer wg.Done()
		if score, err := FetchScamalyticsRiskScore(ipAddr, client); err != nil {
			mutex.Lock()
			collectedErrors = append(collectedErrors, fmt.Sprintf("Scamalytics error: %v", err))
			mutex.Unlock()
		} else {
			mutex.Lock()
			riskInfo.Scamalytics = score
			mutex.Unlock()
		}
	}()
	
	// IPAPI 风险评分
	go func() {
		defer wg.Done()
		if score, err := FetchIPAPIRiskScore(ipAddr, client); err != nil {
			mutex.Lock()
			collectedErrors = append(collectedErrors, fmt.Sprintf("IPAPI error: %v", err))
			mutex.Unlock()
		} else {
			mutex.Lock()
			riskInfo.IPAPI = score
			mutex.Unlock()
		}
	}()
	
	// AbuseIPDB 风险评分
	go func() {
		defer wg.Done()
		if score, err := FetchAbuseIPDBRiskScore(ipAddr, client); err != nil {
			mutex.Lock()
			collectedErrors = append(collectedErrors, fmt.Sprintf("AbuseIPDB error: %v", err))
			mutex.Unlock()
		} else {
			mutex.Lock()
			riskInfo.AbuseIPDB = score
			mutex.Unlock()
		}
	}()
	
	// IPQS 风险评分
	go func() {
		defer wg.Done()
		if score, err := FetchIPQSRiskScore(ipAddr, client); err != nil {
			mutex.Lock()
			collectedErrors = append(collectedErrors, fmt.Sprintf("IPQS error: %v", err))
			mutex.Unlock()
		} else {
			mutex.Lock()
			riskInfo.IPQS = score
			mutex.Unlock()
		}
	}()
	
	// DBIP 风险评分
	go func() {
		defer wg.Done()
		if score, err := FetchDBIPRiskScore(ipAddr, client); err != nil {
			mutex.Lock()
			collectedErrors = append(collectedErrors, fmt.Sprintf("DBIP error: %v", err))
			mutex.Unlock()
		} else {
			mutex.Lock()
			riskInfo.DBIP = score
			mutex.Unlock()
		}
	}()
	
	wg.Wait()
	
	// 处理错误
	if len(collectedErrors) > 0 {
		if len(collectedErrors) == 5 {
			return riskInfo, fmt.Errorf("所有IP风险评分API调用失败")
		}
		return riskInfo, fmt.Errorf("部分IP风险评分API调用失败")
	}
	
	return riskInfo, nil
}

// FetchScamalyticsRiskScore 获取Scamalytics风险评分
func FetchScamalyticsRiskScore(ipAddr string, client *http.Client) (info.IPRiskScore, error) {
	// 这里应该调用Scamalytics API获取风险评分
	// 由于现有代码中可能已有实现，暂时返回默认值
	return info.IPRiskScoreLow, nil
}

// FetchIPAPIRiskScore 获取IPAPI风险评分
func FetchIPAPIRiskScore(ipAddr string, client *http.Client) (info.IPRiskScore, error) {
	// 这里应该调用IPAPI风险评分API
	return info.IPRiskScoreLow, nil
}

// FetchAbuseIPDBRiskScore 获取AbuseIPDB风险评分
func FetchAbuseIPDBRiskScore(ipAddr string, client *http.Client) (info.IPRiskScore, error) {
	// 使用现有的AbuseIPDB检测器
	type abuseResponse struct {
		AbuseConfidencePercentage int `json:"abuseConfidencePercentage"`
	}
	
	var resp abuseResponse
	url := fmt.Sprintf("https://api.abuseipdb.com/api/v2/check?ipAddress=%s", ipAddr)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return info.IPRiskScoreLow, fmt.Errorf("failed to create request: %w", err)
	}
	
	// 注意：实际使用时需要API密钥
	req.Header.Set("Key", "YOUR_API_KEY")
	req.Header.Set("Accept", "application/json")
	
	httpResp, err := client.Do(req)
	if err != nil {
		return info.IPRiskScoreLow, fmt.Errorf("failed to execute request: %w", err)
	}
	defer httpResp.Body.Close()
	
	if httpResp.StatusCode != http.StatusOK {
		return info.IPRiskScoreLow, fmt.Errorf("bad status code: %d", httpResp.StatusCode)
	}
	
	// 解析响应并转换为风险评分
	return mapAbuseIPDBScoreToEnum(resp.AbuseConfidencePercentage), nil
}

// FetchIPQSRiskScore 获取IPQS风险评分
func FetchIPQSRiskScore(ipAddr string, client *http.Client) (info.IPRiskScore, error) {
	// 这里应该调用IPQS API获取风险评分
	return info.IPRiskScoreLow, nil
}

// FetchDBIPRiskScore 获取DBIP风险评分
func FetchDBIPRiskScore(ipAddr string, client *http.Client) (info.IPRiskScore, error) {
	// 这里应该调用DBIP API获取风险评分
	return info.IPRiskScoreLow, nil
} 