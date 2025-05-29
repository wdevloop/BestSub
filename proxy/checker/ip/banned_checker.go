package ip

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bestruirui/bestsub/proxy/info"
)

// BannedCheckResult 单个平台的封禁检测结果
type BannedCheckResult struct {
	Platform string
	Status   int // 0: 正常, 1: 标记, 2: 封禁
	Error    error
}

// GetIPBannedInfo 获取IP在各个平台的封禁状态信息
func GetIPBannedInfo(ipAddr string) (info.IPBannedInfo, error) {
	client := &http.Client{Timeout: defaultTimeout}
	bannedInfo := info.IPBannedInfo{}
	
	// 执行多个平台的封禁检测
	results := make(chan BannedCheckResult, 6)
	
	// 启动并发检测
	go checkGoogleBanned(ipAddr, client, results)
	go checkNetflixBanned(ipAddr, client, results)
	go checkAmazonBanned(ipAddr, client, results)
	go checkCloudFlareBanned(ipAddr, client, results)
	go checkSpotifyBanned(ipAddr, client, results)
	go checkOpenAIBanned(ipAddr, client, results)
	
	// 收集结果
	for i := 0; i < 6; i++ {
		result := <-results
		switch result.Status {
		case 0:
			bannedInfo.Normal++
		case 1:
			bannedInfo.Tagged++
		case 2:
			bannedInfo.Banned++
		}
	}
	
	return bannedInfo, nil
}

// checkGoogleBanned 检测Google平台封禁状态
func checkGoogleBanned(ipAddr string, client *http.Client, results chan<- BannedCheckResult) {
	status := 0 // 默认正常
	
	// 检测Google访问状态
	req, err := http.NewRequest("GET", "https://www.google.com/generate_204", nil)
	if err != nil {
		results <- BannedCheckResult{Platform: "Google", Status: status, Error: err}
		return
	}
	req.Header.Set("User-Agent", userAgent)
	
	resp, err := client.Do(req)
	if err != nil {
		status = 2 // 无法访问，视为封禁
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 429 {
			status = 1 // 限制访问，视为标记
		} else if resp.StatusCode != 204 {
			status = 2 // 非正常响应，视为封禁
		}
	}
	
	results <- BannedCheckResult{Platform: "Google", Status: status, Error: err}
}

// checkNetflixBanned 检测Netflix平台封禁状态
func checkNetflixBanned(ipAddr string, client *http.Client, results chan<- BannedCheckResult) {
	status := 0
	
	req, err := http.NewRequest("GET", "https://www.netflix.com/title/70143836", nil)
	if err != nil {
		results <- BannedCheckResult{Platform: "Netflix", Status: status, Error: err}
		return
	}
	req.Header.Set("User-Agent", userAgent)
	
	resp, err := client.Do(req)
	if err != nil {
		status = 2
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 403 {
			status = 2 // 禁止访问
		} else if resp.StatusCode >= 400 {
			status = 1 // 其他错误状态
		}
	}
	
	results <- BannedCheckResult{Platform: "Netflix", Status: status, Error: err}
}

// checkAmazonBanned 检测Amazon平台封禁状态
func checkAmazonBanned(ipAddr string, client *http.Client, results chan<- BannedCheckResult) {
	status := 0
	
	req, err := http.NewRequest("GET", "https://www.amazon.com/dp/B07HXRHDVR", nil)
	if err != nil {
		results <- BannedCheckResult{Platform: "Amazon", Status: status, Error: err}
		return
	}
	req.Header.Set("User-Agent", userAgent)
	
	resp, err := client.Do(req)
	if err != nil {
		status = 2
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 503 {
			status = 2 // 服务不可用，可能是IP封禁
		} else if resp.StatusCode >= 400 {
			status = 1
		}
	}
	
	results <- BannedCheckResult{Platform: "Amazon", Status: status, Error: err}
}

// checkCloudFlareBanned 检测CloudFlare平台封禁状态
func checkCloudFlareBanned(ipAddr string, client *http.Client, results chan<- BannedCheckResult) {
	status := 0
	
	req, err := http.NewRequest("GET", "https://www.cloudflare.com/cdn-cgi/trace", nil)
	if err != nil {
		results <- BannedCheckResult{Platform: "CloudFlare", Status: status, Error: err}
		return
	}
	req.Header.Set("User-Agent", userAgent)
	
	resp, err := client.Do(req)
	if err != nil {
		status = 2
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 403 {
			status = 2 // CloudFlare阻止
		} else if resp.StatusCode == 429 {
			status = 1 // 限速
		}
	}
	
	results <- BannedCheckResult{Platform: "CloudFlare", Status: status, Error: err}
}

// checkSpotifyBanned 检测Spotify平台封禁状态
func checkSpotifyBanned(ipAddr string, client *http.Client, results chan<- BannedCheckResult) {
	status := 0
	
	req, err := http.NewRequest("GET", "https://open.spotify.com/", nil)
	if err != nil {
		results <- BannedCheckResult{Platform: "Spotify", Status: status, Error: err}
		return
	}
	req.Header.Set("User-Agent", userAgent)
	
	resp, err := client.Do(req)
	if err != nil {
		status = 2
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 403 {
			status = 2
		} else if resp.StatusCode >= 400 {
			status = 1
		}
	}
	
	results <- BannedCheckResult{Platform: "Spotify", Status: status, Error: err}
}

// checkOpenAIBanned 检测OpenAI平台封禁状态
func checkOpenAIBanned(ipAddr string, client *http.Client, results chan<- BannedCheckResult) {
	status := 0
	
	req, err := http.NewRequest("GET", "https://chat.openai.com/", nil)
	if err != nil {
		results <- BannedCheckResult{Platform: "OpenAI", Status: status, Error: err}
		return
	}
	req.Header.Set("User-Agent", userAgent)
	
	resp, err := client.Do(req)
	if err != nil {
		status = 2
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 403 {
			status = 2
		} else if resp.StatusCode == 429 {
			status = 1
		}
	}
	
	results <- BannedCheckResult{Platform: "OpenAI", Status: status, Error: err}
} 