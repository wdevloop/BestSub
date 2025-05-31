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
		score, _, _, _, _, _, err := FetchScamalyticsRiskData(ipAddr, client)
		if err != nil {
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
