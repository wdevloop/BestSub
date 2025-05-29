package checker

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (c *Checker) SpotifyTest() {
	ctx, cancel := context.WithCancel(c.Proxy.Ctx)
	defer cancel()

	// 准备 POST 数据
	data := url.Values{}
	data.Set("birth_day", "11")
	data.Set("birth_month", "11")
	data.Set("birth_year", "2000")
	data.Set("collect_personal_info", "undefined")
	data.Set("creation_flow", "")
	data.Set("creation_point", "https://www.spotify.com/hk-en/")
	data.Set("displayname", "Gay Lord")
	data.Set("gender", "male")
	data.Set("iagree", "1")
	data.Set("key", "a1e486e2729f46d6bb368d6b2bcda326")
	data.Set("platform", "www")
	data.Set("referrer", "")
	data.Set("send-email", "0")
	data.Set("thirdpartyemail", "0")
	data.Set("identifier_token", "AgE6YTvEzkReHNfJpO114514")

	req, err := http.NewRequestWithContext(ctx, "POST", "https://spclient.wg.spotify.com/signup/public/v1/account", strings.NewReader(data.Encode()))
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept-Language", "en")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.Proxy.Client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return
	}

	// 检查响应状态
	status, statusOk := result["status"].(string)
	isLaunched, launchedOk := result["is_country_launched"].(bool)

	if statusOk && launchedOk {
		// 根据 Shell 脚本逻辑：status = "311" 且 is_country_launched = true 表示支持
		if status == "311" && isLaunched {
			c.Proxy.Info.Unlock.Spotify = true
		}
	}
}
