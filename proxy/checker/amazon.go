package checker

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func (c *Checker) AmazonTest() {
	ctx, cancel := context.WithCancel(c.Proxy.Ctx)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.primevideo.com", nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")

	resp, err := c.Proxy.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	bodyStr := string(body)
	// 查找 "currentTerritory": 字段
	territoryIndex := strings.Index(bodyStr, `"currentTerritory":`)
	if territoryIndex != -1 {
		// 查找第三个引号后的内容
		start := territoryIndex + len(`"currentTerritory":`)
		// 跳过可能的空格和第一个引号
		for start < len(bodyStr) && (bodyStr[start] == ' ' || bodyStr[start] == '\t' || bodyStr[start] == '"') {
			start++
		}
		// 如果是引号开始，再跳过一个引号找到实际内容
		if start > 0 && bodyStr[start-1] == '"' {
			end := strings.Index(bodyStr[start:], `"`)
			if end != -1 && end > 0 {
				territory := bodyStr[start : start+end]
				if territory != "" {
					c.Proxy.Info.Unlock.Amazon = true
				}
			}
		}
	}
}
