package checker

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func (c *Checker) TiktokTest() {
	ctx, cancel := context.WithCancel(c.Proxy.Ctx)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.tiktok.com/", nil)
	if err != nil {
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	req.Header.Set("Accept-Language", "en")

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
	// 查找 "region": 字段
	regionIndex := strings.Index(bodyStr, `"region":`)
	if regionIndex != -1 {
		// 提取 region 值
		start := regionIndex + len(`"region":"`)
		end := strings.Index(bodyStr[start:], `"`)
		if end != -1 && end > 0 {
			region := bodyStr[start : start+end]
			if region != "" {
				c.Proxy.Info.Unlock.Tiktok = true
			}
		}
	}
}
